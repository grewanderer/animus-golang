package auditexport

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type Worker struct {
	store  Store
	audit  repo.AuditEventAppender
	logger Logger
	cfg    Config
	now    func() time.Time
	client *http.Client
}

func NewWorker(store Store, audit repo.AuditEventAppender, logger Logger, cfg Config) *Worker {
	return &Worker{
		store:  store,
		audit:  audit,
		logger: logger,
		cfg:    cfg,
		now:    time.Now,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.store == nil {
		return
	}
	if !w.cfg.Enabled() {
		return
	}
	if w.cfg.PollInterval <= 0 {
		w.cfg.PollInterval = 5 * time.Second
	}
	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processOnce(ctx)
		}
	}
}

func (w *Worker) processOnce(ctx context.Context) {
	now := w.now().UTC()
	inflight := w.cfg.InflightTimeout
	if inflight <= 0 {
		inflight = 2 * time.Minute
	}
	limit := w.cfg.BatchSize
	if limit <= 0 {
		limit = 50
	}
	batch, err := w.store.ClaimDue(ctx, now, inflight, limit)
	if err != nil {
		w.logWarn("audit export claim failed", "error", err)
		return
	}
	if len(batch) == 0 {
		return
	}

	sinkCache := map[string]Sink{}
	for _, job := range batch {
		sink, ok := sinkCache[job.SinkID]
		if !ok {
			s, err := w.store.GetSink(ctx, job.SinkID)
			if err != nil {
				w.handleJobFailure(ctx, job, now, fmt.Errorf("sink lookup failed"))
				continue
			}
			sink = s
			sinkCache[job.SinkID] = s
		}
		event, err := w.store.GetEvent(ctx, job.EventID)
		if err != nil {
			w.handleJobFailure(ctx, job, now, fmt.Errorf("event lookup failed"))
			continue
		}

		w.auditAttempt(ctx, sink, event, job)
		if err := Deliver(ctx, sink, event, w.client); err != nil {
			w.handleJobFailure(ctx, job, now, err)
			continue
		}
		if err := w.store.MarkDelivered(ctx, job.OutboxID, now); err != nil {
			w.logWarn("audit export delivery mark failed", "outbox_id", job.OutboxID, "error", err)
			continue
		}
		w.auditDelivered(ctx, sink, event, job)
	}
}

func (w *Worker) handleJobFailure(ctx context.Context, job OutboxRecord, now time.Time, err error) {
	if w == nil {
		return
	}
	delay := backoffDelay(job.Attempt, w.cfg.RetryBaseDelay, w.cfg.RetryMaxDelay)
	nextAttempt := now.Add(delay)
	if markErr := w.store.MarkFailed(ctx, job.OutboxID, trimError(err), nextAttempt); markErr != nil {
		w.logWarn("audit export failure mark failed", "outbox_id", job.OutboxID, "error", markErr)
	}
	w.logWarn("audit export delivery failed", "outbox_id", job.OutboxID, "attempt", job.Attempt, "error", err)
}

func (w *Worker) auditAttempt(ctx context.Context, sink Sink, event domain.AuditEvent, job OutboxRecord) {
	if w == nil || w.audit == nil {
		return
	}
	payload := domain.Metadata{
		"event_id":    event.EventID,
		"sink_id":     sink.SinkID,
		"destination": sink.Destination,
		"format":      sink.Format,
		"attempt":     job.Attempt,
		"outbox_id":   job.OutboxID,
	}
	_, _ = w.audit.Append(ctx, domain.AuditEvent{
		OccurredAt:   w.now().UTC(),
		Actor:        "system:audit-exporter",
		Action:       "audit.export.attempted",
		ResourceType: "audit_export",
		ResourceID:   fmt.Sprintf("%s:%d", sink.SinkID, event.EventID),
		Payload:      payload,
	})
}

func (w *Worker) auditDelivered(ctx context.Context, sink Sink, event domain.AuditEvent, job OutboxRecord) {
	if w == nil || w.audit == nil {
		return
	}
	payload := domain.Metadata{
		"event_id":    event.EventID,
		"sink_id":     sink.SinkID,
		"destination": sink.Destination,
		"format":      sink.Format,
		"attempt":     job.Attempt,
		"outbox_id":   job.OutboxID,
	}
	_, _ = w.audit.Append(ctx, domain.AuditEvent{
		OccurredAt:   w.now().UTC(),
		Actor:        "system:audit-exporter",
		Action:       "audit.export.delivered",
		ResourceType: "audit_export",
		ResourceID:   fmt.Sprintf("%s:%d", sink.SinkID, event.EventID),
		Payload:      payload,
	})
}

func (w *Worker) logWarn(msg string, args ...any) {
	if w != nil && w.logger != nil {
		w.logger.Warn(msg, args...)
	}
}

func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	if base <= 0 {
		base = 5 * time.Second
	}
	if max <= 0 {
		max = 5 * time.Minute
	}
	if attempt <= 1 {
		if base > max {
			return max
		}
		return base
	}
	delay := base
	for i := 1; i < attempt; i++ {
		if delay >= max/2 {
			delay = max
			break
		}
		delay *= 2
	}
	if delay > max {
		delay = max
	}
	return delay
}

func trimError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if len(msg) > 500 {
		return msg[:500]
	}
	return msg
}

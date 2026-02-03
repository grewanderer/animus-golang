package auditexport

import (
	"context"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
)

type Store interface {
	UpsertSink(ctx context.Context, sink Sink) (Sink, error)
	BackfillOutbox(ctx context.Context, sinkID string) error
	ClaimDue(ctx context.Context, now time.Time, inflightTimeout time.Duration, limit int) ([]OutboxRecord, error)
	GetEvent(ctx context.Context, eventID int64) (domain.AuditEvent, error)
	GetSink(ctx context.Context, sinkID string) (Sink, error)
	MarkDelivered(ctx context.Context, outboxID int64, deliveredAt time.Time) error
	MarkFailed(ctx context.Context, outboxID int64, lastError string, nextAttemptAt time.Time) error
}

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type auditAPI struct {
	logger *slog.Logger
	db     *sql.DB
}

func newAuditAPI(logger *slog.Logger, db *sql.DB) *auditAPI {
	return &auditAPI{
		logger: logger,
		db:     db,
	}
}

func (api *auditAPI) register(mux *http.ServeMux) {
	mux.HandleFunc("GET /events", api.handleListEvents)
	mux.HandleFunc("GET /events/{event_id}", api.handleGetEvent)
	mux.HandleFunc("POST /export", api.handleExport)
}

func (api *auditAPI) handleExport(w http.ResponseWriter, r *http.Request) {
	api.writeJSON(w, http.StatusNotImplemented, map[string]any{
		"error":  "export_not_configured",
		"detail": "configure an audit exporter to enable streaming exports",
	})
}

type auditEvent struct {
	EventID         int64           `json:"event_id"`
	OccurredAt      time.Time       `json:"occurred_at"`
	Actor           string          `json:"actor"`
	Action          string          `json:"action"`
	ResourceType    string          `json:"resource_type"`
	ResourceID      string          `json:"resource_id"`
	RequestID       string          `json:"request_id,omitempty"`
	IP              string          `json:"ip,omitempty"`
	UserAgent       string          `json:"user_agent,omitempty"`
	Payload         json.RawMessage `json:"payload"`
	IntegritySHA256 string          `json:"integrity_sha256"`
}

func (api *auditAPI) handleListEvents(w http.ResponseWriter, r *http.Request) {
	limit := clampInt(parseIntQuery(r, "limit", 100), 1, 500)
	beforeID := parseInt64Query(r, "before_event_id", 0)

	actor := strings.TrimSpace(r.URL.Query().Get("actor"))
	action := strings.TrimSpace(r.URL.Query().Get("action"))
	resourceType := strings.TrimSpace(r.URL.Query().Get("resource_type"))
	resourceID := strings.TrimSpace(r.URL.Query().Get("resource_id"))
	requestID := strings.TrimSpace(r.URL.Query().Get("request_id"))

	where := make([]string, 0, 6)
	args := make([]any, 0, 8)

	if beforeID > 0 {
		args = append(args, beforeID)
		where = append(where, "event_id < $"+strconv.Itoa(len(args)))
	}
	if actor != "" {
		args = append(args, actor)
		where = append(where, "actor = $"+strconv.Itoa(len(args)))
	}
	if action != "" {
		args = append(args, action)
		where = append(where, "action = $"+strconv.Itoa(len(args)))
	}
	if resourceType != "" {
		args = append(args, resourceType)
		where = append(where, "resource_type = $"+strconv.Itoa(len(args)))
	}
	if resourceID != "" {
		args = append(args, resourceID)
		where = append(where, "resource_id = $"+strconv.Itoa(len(args)))
	}
	if requestID != "" {
		args = append(args, requestID)
		where = append(where, "request_id = $"+strconv.Itoa(len(args)))
	}

	args = append(args, limit)
	query := `SELECT event_id, occurred_at, actor, action, resource_type, resource_id, request_id, ip, user_agent, payload, integrity_sha256
		FROM audit_events`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY event_id DESC LIMIT $" + strconv.Itoa(len(args))

	rows, err := api.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer rows.Close()

	events := make([]auditEvent, 0, limit)
	for rows.Next() {
		var (
			ev         auditEvent
			reqID      sql.NullString
			ip         sql.NullString
			userAgent  sql.NullString
			payloadRaw []byte
		)
		if err := rows.Scan(
			&ev.EventID,
			&ev.OccurredAt,
			&ev.Actor,
			&ev.Action,
			&ev.ResourceType,
			&ev.ResourceID,
			&reqID,
			&ip,
			&userAgent,
			&payloadRaw,
			&ev.IntegritySHA256,
		); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}

		ev.RequestID = strings.TrimSpace(reqID.String)
		ev.IP = strings.TrimSpace(ip.String)
		ev.UserAgent = strings.TrimSpace(userAgent.String)
		ev.Payload = normalizeJSON(payloadRaw)
		events = append(events, ev)
	}
	if err := rows.Err(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	resp := map[string]any{"events": events}
	if len(events) > 0 {
		resp["next_before_event_id"] = events[len(events)-1].EventID
	}
	api.writeJSON(w, http.StatusOK, resp)
}

func (api *auditAPI) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	rawID := strings.TrimSpace(r.PathValue("event_id"))
	if rawID == "" {
		api.writeError(w, r, http.StatusBadRequest, "event_id_required")
		return
	}
	eventID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || eventID <= 0 {
		api.writeError(w, r, http.StatusBadRequest, "event_id_required")
		return
	}

	var (
		ev         auditEvent
		reqID      sql.NullString
		ip         sql.NullString
		userAgent  sql.NullString
		payloadRaw []byte
	)
	err = api.db.QueryRowContext(
		r.Context(),
		`SELECT event_id, occurred_at, actor, action, resource_type, resource_id, request_id, ip, user_agent, payload, integrity_sha256
		 FROM audit_events
		 WHERE event_id = $1`,
		eventID,
	).Scan(
		&ev.EventID,
		&ev.OccurredAt,
		&ev.Actor,
		&ev.Action,
		&ev.ResourceType,
		&ev.ResourceID,
		&reqID,
		&ip,
		&userAgent,
		&payloadRaw,
		&ev.IntegritySHA256,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	ev.RequestID = strings.TrimSpace(reqID.String)
	ev.IP = strings.TrimSpace(ip.String)
	ev.UserAgent = strings.TrimSpace(userAgent.String)
	ev.Payload = normalizeJSON(payloadRaw)

	api.writeJSON(w, http.StatusOK, ev)
}

func (api *auditAPI) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(body)
}

func (api *auditAPI) writeError(w http.ResponseWriter, r *http.Request, status int, code string) {
	api.writeJSON(w, status, map[string]any{
		"error":      code,
		"request_id": r.Header.Get("X-Request-Id"),
	})
}

func normalizeJSON(raw []byte) json.RawMessage {
	raw = bytesTrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("{}")
	}
	return raw
}

func bytesTrimSpace(in []byte) []byte {
	return []byte(strings.TrimSpace(string(in)))
}

func parseIntQuery(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return parsed
}

func parseInt64Query(r *http.Request, key string, def int64) int64 {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	parsed, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return parsed
}

func clampInt(v int, min int, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

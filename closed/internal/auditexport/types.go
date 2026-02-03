package auditexport

import (
	"encoding/json"
	"time"
)

const DefaultSinkID = "default"

const (
	OutboxStatusPending  = "pending"
	OutboxStatusInflight = "inflight"
	OutboxStatusRetry    = "retry"
	OutboxStatusDelivered = "delivered"
)

type Sink struct {
	SinkID      string
	Name        string
	Destination string
	Format      string
	Config      json.RawMessage
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OutboxRecord struct {
	OutboxID int64
	EventID  int64
	SinkID   string
	Attempt  int
}

type SinkConfig struct {
	WebhookURL     string            `json:"webhook_url,omitempty"`
	WebhookHeaders map[string]string `json:"webhook_headers,omitempty"`
	SyslogAddr     string            `json:"syslog_addr,omitempty"`
	SyslogProtocol string            `json:"syslog_protocol,omitempty"`
	SyslogTag      string            `json:"syslog_tag,omitempty"`
}

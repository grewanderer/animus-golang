package auditexport

import (
	"encoding/json"
	"strings"
	"time"
)

func DefaultSinkFromConfig(cfg Config, now time.Time) (Sink, error) {
	dest := normalizeDestination(cfg.Destination)
	format := strings.ToLower(strings.TrimSpace(cfg.Format))
	if format == "" {
		format = "ndjson"
	}
	config := SinkConfig{}
	switch dest {
	case "webhook":
		config.WebhookURL = strings.TrimSpace(cfg.WebhookURL)
		config.WebhookHeaders = cfg.WebhookHeaders
	case "syslog":
		config.SyslogAddr = strings.TrimSpace(cfg.SyslogAddr)
		config.SyslogProtocol = strings.TrimSpace(cfg.SyslogProtocol)
		config.SyslogTag = strings.TrimSpace(cfg.SyslogTag)
	default:
		dest = "none"
	}
	blob, _ := json.Marshal(config)
	return Sink{
		SinkID:      DefaultSinkID,
		Name:        "default",
		Destination: dest,
		Format:      format,
		Config:      blob,
		Enabled:     dest != "none",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

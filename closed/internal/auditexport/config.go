package auditexport

import (
	"fmt"
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/platform/env"
)

// Config controls audit export format and destination.
type Config struct {
	Format      string
	Destination string
}

func ConfigFromEnv() (Config, error) {
	cfg := Config{
		Format:      env.String("AUDIT_EXPORT_FORMAT", "ndjson"),
		Destination: env.String("AUDIT_EXPORT_DESTINATION", "http"),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	format := strings.ToLower(strings.TrimSpace(c.Format))
	destination := strings.ToLower(strings.TrimSpace(c.Destination))
	if format == "" {
		format = "ndjson"
	}
	if destination == "" {
		destination = "http"
	}
	if format != "ndjson" {
		return fmt.Errorf("unsupported audit export format: %s", format)
	}
	if destination != "http" {
		return fmt.Errorf("unsupported audit export destination: %s", destination)
	}
	return nil
}

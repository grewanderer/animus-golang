package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/internal/platform/env"
)

type Mode string

const (
	ModeOIDC     Mode = "oidc"
	ModeDev      Mode = "dev"
	ModeDisabled Mode = "disabled"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type Config struct {
	Mode Mode

	RolesClaim string
	EmailClaim string

	SessionCookieName     string
	SessionCookieSecure   bool
	SessionCookieMaxAge   time.Duration
	SessionCookieSameSite string

	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	OIDCScopes       []string

	DevSubject string
	DevEmail   string
	DevRoles   []string
}

func ConfigFromEnv() (Config, error) {
	modeRaw := strings.ToLower(strings.TrimSpace(env.String("AUTH_MODE", string(ModeOIDC))))
	var mode Mode
	switch modeRaw {
	case string(ModeOIDC):
		mode = ModeOIDC
	case string(ModeDev):
		mode = ModeDev
	case string(ModeDisabled):
		mode = ModeDisabled
	default:
		return Config{}, fmt.Errorf("AUTH_MODE must be one of: oidc, dev, disabled (got %q)", modeRaw)
	}

	sessionCookieSecure, err := env.Bool("AUTH_SESSION_COOKIE_SECURE", true)
	if err != nil {
		return Config{}, err
	}
	maxAgeSeconds, err := env.Int("AUTH_SESSION_MAX_AGE_SECONDS", 3600)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Mode:                  mode,
		RolesClaim:            env.String("AUTH_ROLES_CLAIM", "roles"),
		EmailClaim:            env.String("AUTH_EMAIL_CLAIM", "email"),
		SessionCookieName:     env.String("AUTH_SESSION_COOKIE_NAME", "animus_session"),
		SessionCookieSecure:   sessionCookieSecure,
		SessionCookieMaxAge:   time.Duration(maxAgeSeconds) * time.Second,
		SessionCookieSameSite: env.String("AUTH_SESSION_COOKIE_SAMESITE", "Lax"),
		OIDCIssuerURL:         env.String("OIDC_ISSUER_URL", ""),
		OIDCClientID:          env.String("OIDC_CLIENT_ID", ""),
		OIDCClientSecret:      env.String("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:       env.String("OIDC_REDIRECT_URL", ""),
		OIDCScopes:            parseScopes(env.String("OIDC_SCOPES", "openid profile email")),
		DevSubject:            env.String("DEV_AUTH_SUBJECT", "dev-user"),
		DevEmail:              env.String("DEV_AUTH_EMAIL", "dev-user@example.local"),
		DevRoles:              parseCSV(env.String("DEV_AUTH_ROLES", "admin")),
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(string(c.Mode)) == "" {
		return errors.New("AUTH_MODE is required")
	}
	if strings.TrimSpace(c.RolesClaim) == "" {
		return errors.New("AUTH_ROLES_CLAIM is required")
	}
	if strings.TrimSpace(c.EmailClaim) == "" {
		return errors.New("AUTH_EMAIL_CLAIM is required")
	}
	if strings.TrimSpace(c.SessionCookieName) == "" {
		return errors.New("AUTH_SESSION_COOKIE_NAME is required")
	}
	if c.SessionCookieMaxAge <= 0 {
		return errors.New("AUTH_SESSION_MAX_AGE_SECONDS must be positive")
	}
	if strings.TrimSpace(c.SessionCookieSameSite) == "" {
		return errors.New("AUTH_SESSION_COOKIE_SAMESITE is required")
	}

	switch c.Mode {
	case ModeOIDC:
		if strings.TrimSpace(c.OIDCIssuerURL) == "" {
			return errors.New("OIDC_ISSUER_URL is required when AUTH_MODE=oidc")
		}
		if strings.TrimSpace(c.OIDCClientID) == "" {
			return errors.New("OIDC_CLIENT_ID is required when AUTH_MODE=oidc")
		}
	case ModeDev:
		if strings.TrimSpace(c.DevSubject) == "" {
			return errors.New("DEV_AUTH_SUBJECT is required when AUTH_MODE=dev")
		}
		if len(c.DevRoles) == 0 {
			return errors.New("DEV_AUTH_ROLES must be non-empty when AUTH_MODE=dev")
		}
	case ModeDisabled:
	default:
		return fmt.Errorf("unsupported auth mode: %q", c.Mode)
	}

	return nil
}

func (c Config) ValidateForLogin() error {
	if c.Mode != ModeOIDC {
		return fmt.Errorf("login requires AUTH_MODE=oidc (got %q)", c.Mode)
	}
	if strings.TrimSpace(c.OIDCClientSecret) == "" {
		return errors.New("OIDC_CLIENT_SECRET is required for login endpoints")
	}
	if strings.TrimSpace(c.OIDCRedirectURL) == "" {
		return errors.New("OIDC_REDIRECT_URL is required for login endpoints")
	}
	return nil
}

func parseScopes(value string) []string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return []string{"openid", "profile", "email"}
	}
	return fields
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		item := strings.ToLower(strings.TrimSpace(part))
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

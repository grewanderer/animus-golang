package auth

import (
	"os"
	"testing"
)

func TestConfigFromEnv_Dev(t *testing.T) {
	t.Setenv("AUTH_MODE", "dev")
	t.Setenv("DEV_AUTH_SUBJECT", "dev")
	t.Setenv("DEV_AUTH_EMAIL", "dev@example.local")
	t.Setenv("DEV_AUTH_ROLES", "admin,viewer")

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv() err=%v", err)
	}
	if cfg.Mode != ModeDev {
		t.Fatalf("Mode=%q, want dev", cfg.Mode)
	}
	if cfg.DevSubject != "dev" {
		t.Fatalf("DevSubject=%q, want dev", cfg.DevSubject)
	}
	if len(cfg.DevRoles) != 2 {
		t.Fatalf("DevRoles=%v, want 2 roles", cfg.DevRoles)
	}
}

func TestConfigFromEnv_OIDC_RequiresIssuerAndClientID(t *testing.T) {
	_ = os.Unsetenv("OIDC_ISSUER_URL")
	_ = os.Unsetenv("OIDC_CLIENT_ID")
	t.Setenv("AUTH_MODE", "oidc")

	_, err := ConfigFromEnv()
	if err == nil {
		t.Fatalf("expected error")
	}
}

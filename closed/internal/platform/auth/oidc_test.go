package auth

import (
	"testing"
)

func TestSafeReturnTo(t *testing.T) {
	if got := safeReturnTo(""); got != "/" {
		t.Fatalf("safeReturnTo()=%q, want /", got)
	}
	if got := safeReturnTo("/app"); got != "/app" {
		t.Fatalf("safeReturnTo()=%q, want /app", got)
	}
	if got := safeReturnTo("https://evil.test/phish"); got != "/" {
		t.Fatalf("safeReturnTo()=%q, want /", got)
	}
	if got := safeReturnTo("//evil"); got != "/" {
		t.Fatalf("safeReturnTo()=%q, want /", got)
	}
}

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/repo"
)

type stubSessionStore struct {
	record       repo.SessionRecord
	revoked      bool
	revokedBy    string
	revokeReason string
}

func (s *stubSessionStore) Create(ctx context.Context, record repo.SessionRecord) error {
	s.record = record
	return nil
}

func (s *stubSessionStore) Get(ctx context.Context, sessionID string) (repo.SessionRecord, error) {
	return s.record, nil
}

func (s *stubSessionStore) ListActiveBySubject(ctx context.Context, subject string, limit int) ([]repo.SessionRecord, error) {
	return nil, nil
}

func (s *stubSessionStore) UpdateLastSeen(ctx context.Context, sessionID string, at time.Time) error {
	return nil
}

func (s *stubSessionStore) Revoke(ctx context.Context, sessionID, revokedBy, reason string, at time.Time) (bool, error) {
	s.revoked = true
	s.revokedBy = revokedBy
	s.revokeReason = reason
	return true, nil
}

func (s *stubSessionStore) RevokeBySubject(ctx context.Context, subject, revokedBy, reason string, at time.Time) (int, error) {
	return 0, nil
}

func TestSessionManagerExpiresSession(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	store := &stubSessionStore{
		record: repo.SessionRecord{
			SessionID: "sess-1",
			Subject:   "user-1",
			ExpiresAt: now.Add(-1 * time.Minute),
		},
	}
	mgr := SessionManager{
		Store: store,
		Now:   func() time.Time { return now },
	}

	_, err := mgr.GetSession(context.Background(), "sess-1")
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if !store.revoked {
		t.Fatalf("expected session revoked on expiry")
	}
	if store.revokeReason != "expired" {
		t.Fatalf("expected revoke reason expired, got %q", store.revokeReason)
	}
}

package artifacts

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/repo"
	store "github.com/animus-labs/animus-go/closed/internal/storage/objectstore"
)

type stubArtifactRepo struct {
	createCalled bool
	getCalled    bool
	projectID    string
	created      domain.Artifact
	getArtifact  domain.Artifact
}

func (s *stubArtifactRepo) CreateArtifact(ctx context.Context, projectID string, artifact domain.Artifact) (domain.Artifact, error) {
	s.createCalled = true
	s.projectID = projectID
	s.created = artifact
	return artifact, nil
}

func (s *stubArtifactRepo) GetArtifact(ctx context.Context, projectID, id string) (domain.Artifact, error) {
	s.getCalled = true
	if s.getArtifact.ID == "" {
		return domain.Artifact{}, repo.ErrNotFound
	}
	return s.getArtifact, nil
}

func (s *stubArtifactRepo) ListArtifacts(ctx context.Context, filter repo.ArtifactFilter) ([]domain.Artifact, error) {
	return nil, nil
}

func (s *stubArtifactRepo) UpdateRetention(ctx context.Context, projectID, id string, retentionUntil *time.Time, legalHold bool) error {
	return nil
}

type stubStore struct {
	putCalls int
	getCalls int
}

func (s *stubStore) Put(ctx context.Context, bucket, key string, body io.Reader, size int64, contentType string) error {
	return errors.New("not implemented")
}

func (s *stubStore) Get(ctx context.Context, bucket, key string) (io.ReadCloser, store.ObjectInfo, error) {
	return nil, store.ObjectInfo{}, errors.New("not implemented")
}

func (s *stubStore) Stat(ctx context.Context, bucket, key string) (store.ObjectInfo, error) {
	return store.ObjectInfo{}, errors.New("not implemented")
}

func (s *stubStore) Delete(ctx context.Context, bucket, key string) error {
	return errors.New("not implemented")
}

func (s *stubStore) PresignPut(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	s.putCalls++
	return "https://example.com/upload", nil
}

func (s *stubStore) PresignGet(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	s.getCalls++
	return "https://example.com/download", nil
}

type stubAudit struct {
	events []domain.AuditEvent
}

func (s *stubAudit) Append(ctx context.Context, event domain.AuditEvent) (int64, error) {
	s.events = append(s.events, event)
	return int64(len(s.events)), nil
}

func TestCreateArtifactUploadAppendsAuditAndPresigns(t *testing.T) {
	repo := &stubArtifactRepo{}
	store := &stubStore{}
	audit := &stubAudit{}

	svc, err := NewService(repo, store, "bucket", 5*time.Minute, audit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := svc.CreateArtifactUpload(context.Background(), "proj-1", CreateArtifactInput{
		Kind:      "model",
		SHA256:    "abc123",
		SizeBytes: 42,
	}, AuditContext{Actor: "user", Service: "dataset-registry"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UploadURL == "" || store.putCalls != 1 {
		t.Fatalf("expected presigned upload url")
	}
	if !repo.createCalled {
		t.Fatalf("expected repo create call")
	}
	if len(audit.events) != 2 {
		t.Fatalf("expected 2 audit events, got %d", len(audit.events))
	}
	if audit.events[0].Action != "artifact.create" {
		t.Fatalf("unexpected first audit action: %s", audit.events[0].Action)
	}
	if audit.events[1].Action != "artifact.upload_url_issued" {
		t.Fatalf("unexpected second audit action: %s", audit.events[1].Action)
	}
}

func TestGetArtifactDownloadAppendsAudit(t *testing.T) {
	repo := &stubArtifactRepo{getArtifact: domain.Artifact{ID: "a1", ProjectID: "proj-1", Kind: "model", ObjectKey: "proj-1/artifacts/a1", SHA256: "abc"}}
	store := &stubStore{}
	audit := &stubAudit{}

	svc, err := NewService(repo, store, "bucket", 5*time.Minute, audit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := svc.GetArtifactDownload(context.Background(), "proj-1", "a1", AuditContext{Actor: "user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DownloadURL == "" || store.getCalls != 1 {
		t.Fatalf("expected presigned download url")
	}
	if len(audit.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(audit.events))
	}
	if audit.events[0].Action != "artifact.download_url_issued" {
		t.Fatalf("unexpected audit action: %s", audit.events[0].Action)
	}
}

func TestCreateArtifactUploadRequiresProjectID(t *testing.T) {
	repo := &stubArtifactRepo{}
	store := &stubStore{}
	audit := &stubAudit{}

	svc, err := NewService(repo, store, "bucket", 5*time.Minute, audit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.CreateArtifactUpload(context.Background(), "", CreateArtifactInput{Kind: "model", SHA256: "abc"}, AuditContext{})
	if err == nil {
		t.Fatalf("expected error for missing project id")
	}
	if repo.createCalled {
		t.Fatalf("repo should not be called when project id missing")
	}
	if store.putCalls != 0 {
		t.Fatalf("store should not be called when project id missing")
	}
}

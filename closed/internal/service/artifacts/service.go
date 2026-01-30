package artifacts

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/repo"
	store "github.com/animus-labs/animus-go/closed/internal/storage/objectstore"
	"github.com/google/uuid"
)

// AuditContext captures request identity details for audit logging.
type AuditContext struct {
	Actor     string
	RequestID string
	IP        net.IP
	UserAgent string
	Path      string
	Service   string
}

// CreateArtifactInput describes the metadata required to register an artifact upload.
type CreateArtifactInput struct {
	Kind           string
	ContentType    string
	SizeBytes      int64
	SHA256         string
	Metadata       map[string]any
	RetentionUntil *time.Time
	LegalHold      bool
}

// UploadResult returns the registered artifact and a pre-signed upload URL.
type UploadResult struct {
	Artifact  domain.Artifact
	UploadURL string
}

// DownloadResult returns the artifact and a pre-signed download URL.
type DownloadResult struct {
	Artifact    domain.Artifact
	DownloadURL string
}

// Service coordinates artifact persistence and object storage links.
type Service struct {
	repo       repo.ArtifactRepository
	store      store.Store
	bucket     string
	presignTTL time.Duration
	audit      repo.AuditEventAppender
	now        func() time.Time
}

func NewService(repo repo.ArtifactRepository, store store.Store, bucket string, presignTTL time.Duration, audit repo.AuditEventAppender) (*Service, error) {
	if repo == nil {
		return nil, errors.New("artifact repository is required")
	}
	if store == nil {
		return nil, errors.New("object store is required")
	}
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return nil, errors.New("bucket is required")
	}
	if presignTTL <= 0 {
		presignTTL = 10 * time.Minute
	}
	return &Service{
		repo:       repo,
		store:      store,
		bucket:     bucket,
		presignTTL: presignTTL,
		audit:      audit,
		now:        time.Now,
	}, nil
}

func (s *Service) CreateArtifactUpload(ctx context.Context, projectID string, input CreateArtifactInput, auditCtx AuditContext) (UploadResult, error) {
	if s == nil || s.repo == nil || s.store == nil {
		return UploadResult{}, errors.New("artifact service not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return UploadResult{}, errors.New("project id is required")
	}
	kind := strings.TrimSpace(input.Kind)
	if kind == "" {
		return UploadResult{}, errors.New("artifact kind is required")
	}
	checksum := strings.TrimSpace(input.SHA256)
	if checksum == "" {
		return UploadResult{}, errors.New("artifact sha256 is required")
	}
	if input.SizeBytes < 0 {
		return UploadResult{}, errors.New("size_bytes must be >= 0")
	}
	metadata := input.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	now := s.now().UTC()
	artifactID := uuid.NewString()
	objectKey := fmt.Sprintf("%s/artifacts/%s", projectID, artifactID)
	retention := normalizeRetention(input.RetentionUntil)

	integrity, err := artifactIntegritySHA256(artifactIntegrityInput{
		ArtifactID:     artifactID,
		ProjectID:      projectID,
		Kind:           kind,
		ObjectKey:      objectKey,
		ContentType:    strings.TrimSpace(input.ContentType),
		SizeBytes:      input.SizeBytes,
		SHA256:         checksum,
		Metadata:       metadata,
		RetentionUntil: retention,
		LegalHold:      input.LegalHold,
		CreatedAt:      now,
		CreatedBy:      strings.TrimSpace(auditCtx.Actor),
	})
	if err != nil {
		return UploadResult{}, fmt.Errorf("integrity: %w", err)
	}

	artifact := domain.Artifact{
		ID:              artifactID,
		ProjectID:       projectID,
		Kind:            kind,
		ContentType:     strings.TrimSpace(input.ContentType),
		ObjectKey:       objectKey,
		SHA256:          checksum,
		SizeBytes:       input.SizeBytes,
		Metadata:        domain.Metadata(metadata),
		RetentionUntil:  retention,
		LegalHold:       input.LegalHold,
		CreatedAt:       now,
		CreatedBy:       strings.TrimSpace(auditCtx.Actor),
		IntegritySHA256: integrity,
	}

	created, err := s.repo.CreateArtifact(ctx, projectID, artifact)
	if err != nil {
		return UploadResult{}, err
	}

	uploadURL, err := s.store.PresignPut(ctx, s.bucket, created.ObjectKey, s.presignTTL)
	if err != nil {
		return UploadResult{}, fmt.Errorf("presign upload: %w", err)
	}

	s.appendAudit(ctx, created, auditCtx, "artifact.create", map[string]any{
		"upload_url_issued": false,
	})
	s.appendAudit(ctx, created, auditCtx, "artifact.upload_url_issued", map[string]any{
		"upload_url_issued": true,
	})

	return UploadResult{Artifact: created, UploadURL: uploadURL}, nil
}

func (s *Service) GetArtifact(ctx context.Context, projectID, artifactID string) (domain.Artifact, error) {
	if s == nil || s.repo == nil {
		return domain.Artifact{}, errors.New("artifact service not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return domain.Artifact{}, errors.New("project id is required")
	}
	artifactID = strings.TrimSpace(artifactID)
	if artifactID == "" {
		return domain.Artifact{}, errors.New("artifact id is required")
	}
	return s.repo.GetArtifact(ctx, projectID, artifactID)
}

func (s *Service) GetArtifactDownload(ctx context.Context, projectID, artifactID string, auditCtx AuditContext) (DownloadResult, error) {
	if s == nil || s.repo == nil || s.store == nil {
		return DownloadResult{}, errors.New("artifact service not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return DownloadResult{}, errors.New("project id is required")
	}
	artifactID = strings.TrimSpace(artifactID)
	if artifactID == "" {
		return DownloadResult{}, errors.New("artifact id is required")
	}

	artifact, err := s.repo.GetArtifact(ctx, projectID, artifactID)
	if err != nil {
		return DownloadResult{}, err
	}
	if strings.TrimSpace(artifact.ObjectKey) == "" {
		return DownloadResult{}, errors.New("object key is required")
	}

	url, err := s.store.PresignGet(ctx, s.bucket, artifact.ObjectKey, s.presignTTL)
	if err != nil {
		return DownloadResult{}, fmt.Errorf("presign download: %w", err)
	}

	s.appendAudit(ctx, artifact, auditCtx, "artifact.download_url_issued", map[string]any{
		"download_url_issued": true,
	})

	return DownloadResult{Artifact: artifact, DownloadURL: url}, nil
}

func (s *Service) appendAudit(ctx context.Context, artifact domain.Artifact, auditCtx AuditContext, action string, extra map[string]any) {
	if s.audit == nil {
		return
	}
	payload := map[string]any{
		"service":         strings.TrimSpace(auditCtx.Service),
		"project_id":      artifact.ProjectID,
		"artifact_id":     artifact.ID,
		"kind":            artifact.Kind,
		"object_key":      artifact.ObjectKey,
		"content_type":    artifact.ContentType,
		"size_bytes":      artifact.SizeBytes,
		"sha256":          artifact.SHA256,
		"legal_hold":      artifact.LegalHold,
		"request_path":    auditCtx.Path,
		"retention_until": artifact.RetentionUntil,
	}
	for k, v := range extra {
		payload[k] = v
	}
	_, _ = s.audit.Append(ctx, domain.AuditEvent{
		OccurredAt:   s.now().UTC(),
		Actor:        strings.TrimSpace(auditCtx.Actor),
		Action:       action,
		ResourceType: "artifact",
		ResourceID:   artifact.ID,
		RequestID:    auditCtx.RequestID,
		IP:           auditCtx.IP,
		UserAgent:    auditCtx.UserAgent,
		Payload:      payload,
	})
}

func normalizeRetention(retention *time.Time) *time.Time {
	if retention == nil {
		return nil
	}
	ret := retention.UTC()
	return &ret
}

type artifactIntegrityInput struct {
	ArtifactID     string         `json:"artifact_id"`
	ProjectID      string         `json:"project_id"`
	Kind           string         `json:"kind"`
	ObjectKey      string         `json:"object_key"`
	ContentType    string         `json:"content_type,omitempty"`
	SizeBytes      int64          `json:"size_bytes"`
	SHA256         string         `json:"sha256"`
	Metadata       map[string]any `json:"metadata"`
	RetentionUntil *time.Time     `json:"retention_until,omitempty"`
	LegalHold      bool           `json:"legal_hold"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedBy      string         `json:"created_by"`
}

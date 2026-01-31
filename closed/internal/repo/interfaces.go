package repo

import (
	"context"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
)

type ProjectFilter struct {
	Name      string
	CreatedBy string
	Limit     int
}

type DatasetFilter struct {
	ProjectID string
	Name      string
	Limit     int
}

type DatasetVersionFilter struct {
	ProjectID string
	DatasetID string
	Limit     int
}

type RunFilter struct {
	ProjectID    string
	ExperimentID string
	Status       string
	Limit        int
}

type ArtifactFilter struct {
	ProjectID string
	Kind      string
	Limit     int
}

type ModelFilter struct {
	ProjectID string
	Name      string
	Status    domain.ModelStatus
	Limit     int
}

type RunRecord struct {
	ID             string
	ProjectID      string
	IdempotencyKey string
	Status         string
	PipelineSpec   []byte
	RunSpec        []byte
	SpecHash       string
	CreatedAt      time.Time
}

type PlanRecord struct {
	ID        string
	RunID     string
	ProjectID string
	Plan      []byte
	CreatedAt time.Time
}

// ProjectRepository manages Projects.
type ProjectRepository interface {
	Create(ctx context.Context, project domain.Project) error
	Get(ctx context.Context, id string) (domain.Project, error)
	List(ctx context.Context, filter ProjectFilter) ([]domain.Project, error)
}

// DatasetRepository manages datasets and immutable versions.
type DatasetRepository interface {
	CreateDataset(ctx context.Context, dataset domain.Dataset) error
	GetDataset(ctx context.Context, projectID, id string) (domain.Dataset, error)
	ListDatasets(ctx context.Context, filter DatasetFilter) ([]domain.Dataset, error)

	CreateDatasetVersion(ctx context.Context, version domain.DatasetVersion) error
	GetDatasetVersion(ctx context.Context, projectID, id string) (domain.DatasetVersion, error)
	ListDatasetVersions(ctx context.Context, filter DatasetVersionFilter) ([]domain.DatasetVersion, error)
	NextDatasetVersionOrdinal(ctx context.Context, projectID, datasetID string) (int64, error)
}

// RunRepository manages run state with immutable identity.
type RunRepository interface {
	CreateRun(ctx context.Context, projectID, idempotencyKey string, pipelineSpecJSON, runSpecJSON []byte, specHash string) (RunRecord, bool, error)
	GetRun(ctx context.Context, projectID, id string) (RunRecord, error)
}

// ArtifactRepository manages project-scoped artifacts.
type ArtifactRepository interface {
	CreateArtifact(ctx context.Context, projectID string, artifact domain.Artifact) (domain.Artifact, error)
	GetArtifact(ctx context.Context, projectID, id string) (domain.Artifact, error)
	ListArtifacts(ctx context.Context, filter ArtifactFilter) ([]domain.Artifact, error)
	UpdateRetention(ctx context.Context, projectID, id string, retentionUntil *time.Time, legalHold bool) error
}

// ModelRepository manages model registry entries.
type ModelRepository interface {
	CreateModel(ctx context.Context, model domain.Model) error
	GetModel(ctx context.Context, projectID, id string) (domain.Model, error)
	ListModels(ctx context.Context, filter ModelFilter) ([]domain.Model, error)
	UpdateModelStatus(ctx context.Context, projectID, id string, status domain.ModelStatus) error
}

// PlanRepository stores deterministic execution plans for runs.
type PlanRepository interface {
	UpsertPlan(ctx context.Context, projectID, runID string, planJSON []byte) (PlanRecord, error)
	GetPlan(ctx context.Context, projectID, runID string) (PlanRecord, error)
}

// AuditEventAppender ensures append-only audit writes.
type AuditEventAppender interface {
	Append(ctx context.Context, event domain.AuditEvent) (int64, error)
}

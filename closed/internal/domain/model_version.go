package domain

import (
	"errors"
	"strings"
	"time"
)

// ModelVersion represents a versioned model artifact with immutable provenance.
type ModelVersion struct {
	ID                   string
	ProjectID            string
	ModelID              string
	Version              string
	Status               ModelStatus
	RunID                string
	ArtifactIDs          []string
	DatasetVersionIDs    []string
	EnvLockID            string
	CodeRef              CodeRef
	PolicySnapshotSHA256 string
	CreatedAt            time.Time
	CreatedBy            string
	IntegritySHA256      string
}

func (v ModelVersion) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return errors.New("model version id is required")
	}
	if strings.TrimSpace(v.ProjectID) == "" {
		return errors.New("project id is required")
	}
	if strings.TrimSpace(v.ModelID) == "" {
		return errors.New("model id is required")
	}
	if strings.TrimSpace(v.Version) == "" {
		return errors.New("version is required")
	}
	if !v.Status.Valid() {
		return errors.New("invalid model status")
	}
	if strings.TrimSpace(v.RunID) == "" {
		return errors.New("run id is required")
	}
	if len(v.ArtifactIDs) == 0 {
		return errors.New("artifact ids are required")
	}
	return nil
}

// ModelVersionTransition captures a state transition for a model version.
type ModelVersionTransition struct {
	TransitionID   int64
	ProjectID      string
	ModelVersionID string
	FromStatus     ModelStatus
	ToStatus       ModelStatus
	Action         string
	RequestID      string
	OccurredAt     time.Time
	Actor          string
}

// ModelExport captures an export request for a model version.
type ModelExport struct {
	ExportID        string
	ProjectID       string
	ModelVersionID  string
	Status          string
	Target          string
	CreatedAt       time.Time
	CreatedBy       string
	IntegritySHA256 string
}

func (e ModelExport) Validate() error {
	if strings.TrimSpace(e.ExportID) == "" {
		return errors.New("export id is required")
	}
	if strings.TrimSpace(e.ProjectID) == "" {
		return errors.New("project id is required")
	}
	if strings.TrimSpace(e.ModelVersionID) == "" {
		return errors.New("model version id is required")
	}
	if strings.TrimSpace(e.Status) == "" {
		return errors.New("status is required")
	}
	return nil
}

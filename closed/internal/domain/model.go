package domain

import (
	"errors"
	"strings"
	"time"
)

// ModelStatus represents the lifecycle state of a model.
type ModelStatus string

const (
	ModelStatusDraft      ModelStatus = "draft"
	ModelStatusValidated  ModelStatus = "validated"
	ModelStatusApproved   ModelStatus = "approved"
	ModelStatusDeprecated ModelStatus = "deprecated"
)

// Model is a versioned model entity with lifecycle state.
type Model struct {
	ID              string
	ProjectID       string
	Name            string
	Version         string
	Status          ModelStatus
	ArtifactID      string
	Metadata        Metadata
	CreatedAt       time.Time
	CreatedBy       string
	IntegritySHA256 string
}

func (m Model) Validate() error {
	if strings.TrimSpace(m.ID) == "" {
		return errors.New("model id is required")
	}
	if strings.TrimSpace(m.ProjectID) == "" {
		return errors.New("project id is required")
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("model name is required")
	}
	if !m.Status.Valid() {
		return errors.New("invalid model status")
	}
	return nil
}

func (s ModelStatus) Valid() bool {
	switch s {
	case ModelStatusDraft, ModelStatusValidated, ModelStatusApproved, ModelStatusDeprecated:
		return true
	default:
		return false
	}
}

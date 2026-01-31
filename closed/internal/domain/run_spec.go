package domain

import "time"

// RunSpec is an immutable execution snapshot that binds a PipelineSpec to concrete inputs.
type RunSpec struct {
	RunSpecVersion  string
	ProjectID       string
	PipelineSpec    PipelineSpec
	DatasetBindings map[string]string
	CodeRef         CodeRef
	EnvLock         EnvLock
	CreatedAt       time.Time
	CreatedBy       string
}

// CodeRef identifies the exact source used for execution.
type CodeRef struct {
	RepoURL   string
	CommitSHA string
}

// EnvLock captures the immutable execution environment bindings.
type EnvLock struct {
	ImageDigests  map[string]string
	EnvTemplateID string
	EnvHash       string
}

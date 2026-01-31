package specvalidator

import (
	"fmt"
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/domain"
)

// ValidateRunSpec performs strict validation of a RunSpec.
func ValidateRunSpec(spec domain.RunSpec) error {
	issues := &ValidationError{}

	if strings.TrimSpace(spec.RunSpecVersion) == "" {
		issues.Add("runSpecVersion is required")
	}
	if strings.TrimSpace(spec.ProjectID) == "" {
		issues.Add("projectId is required")
	}
	if strings.TrimSpace(spec.CodeRef.RepoURL) == "" {
		issues.Add("codeRef.repoUrl is required")
	}
	if strings.TrimSpace(spec.CodeRef.CommitSHA) == "" {
		issues.Add("codeRef.commitSha is required")
	}
	if strings.TrimSpace(spec.EnvLock.EnvHash) == "" {
		issues.Add("envLock.envHash is required")
	}
	if spec.DatasetBindings == nil {
		issues.Add("datasetBindings is required")
	}

	if err := ValidatePipelineSpec(spec.PipelineSpec); err != nil {
		issues.Add(fmt.Sprintf("pipelineSpec invalid: %s", err.Error()))
	}

	if spec.DatasetBindings != nil {
		for key, value := range spec.DatasetBindings {
			if strings.TrimSpace(key) == "" {
				issues.Add("datasetBindings contains empty datasetRef")
			}
			if strings.TrimSpace(value) == "" {
				issues.Add(fmt.Sprintf("datasetBinding for %q is empty", key))
			}
		}
	}

	refs := pipelineDatasetRefs(spec.PipelineSpec)
	for ref := range refs {
		if spec.DatasetBindings == nil {
			break
		}
		if value, ok := spec.DatasetBindings[ref]; !ok || strings.TrimSpace(value) == "" {
			issues.Add(fmt.Sprintf("missing dataset binding for %q", ref))
		}
	}
	if spec.DatasetBindings != nil {
		for ref := range spec.DatasetBindings {
			if _, ok := refs[ref]; !ok {
				issues.Add(fmt.Sprintf("dataset binding %q not referenced in pipeline", ref))
			}
		}
	}

	if spec.EnvLock.ImageDigests != nil {
		for key, value := range spec.EnvLock.ImageDigests {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				issues.Add("envLock.imageDigests must not contain empty keys or values")
			}
		}
	}

	return issues.OrNil()
}

func pipelineDatasetRefs(spec domain.PipelineSpec) map[string]struct{} {
	refs := make(map[string]struct{})
	for _, step := range spec.Spec.Steps {
		for _, dataset := range step.Inputs.Datasets {
			ref := strings.TrimSpace(dataset.DatasetRef)
			if ref == "" {
				continue
			}
			refs[ref] = struct{}{}
		}
	}
	return refs
}

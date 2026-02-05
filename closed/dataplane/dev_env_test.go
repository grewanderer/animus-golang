package main

import (
	"testing"

	"github.com/animus-labs/animus-go/closed/internal/dataplane"
	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/platform/k8s"
)

func TestBuildDevEnvJobSpecSetsTTL(t *testing.T) {
	req := dataplane.DevEnvProvisionRequest{
		DevEnvID:    "dev-1",
		ProjectID:   "proj-1",
		TemplateRef: "tmpl-1",
		ImageRef:    "registry.example/dev:latest",
		TTLSeconds:  3600,
		ResourceDefaults: domain.EnvironmentResources{
			CPU: "1",
		},
		ResourceLimits: domain.EnvironmentResources{
			Memory: "2Gi",
			GPU:    1,
		},
	}

	job, err := buildDevEnvJobSpec(req, "job-1", "ns-1", "sa-1", 120)
	if err != nil {
		t.Fatalf("build job: %v", err)
	}
	if job.Spec.ActiveDeadlineSeconds == nil || *job.Spec.ActiveDeadlineSeconds != 3600 {
		t.Fatalf("expected active deadline 3600, got %+v", job.Spec.ActiveDeadlineSeconds)
	}
	if job.Spec.TTLSecondsAfterFinished == nil || *job.Spec.TTLSecondsAfterFinished != 120 {
		t.Fatalf("expected ttl after finished 120, got %+v", job.Spec.TTLSecondsAfterFinished)
	}
	if got := job.Spec.Template.Spec.ServiceAccountName; got != "sa-1" {
		t.Fatalf("expected service account sa-1, got %q", got)
	}
	if job.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] != "1" {
		t.Fatalf("expected cpu request to be set")
	}
	if job.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] != "2Gi" {
		t.Fatalf("expected memory limit to be set")
	}
	if job.Spec.Template.Spec.Containers[0].Resources.Limits["nvidia.com/gpu"] != "1" {
		t.Fatalf("expected gpu limit to be set")
	}
	if !hasEnvVar(job.Spec.Template.Spec.Containers[0].Env, "ANIMUS_DEV_ENV_TTL_SECONDS", "3600") {
		t.Fatalf("expected ttl env var to be set")
	}
}

func hasEnvVar(env []k8s.EnvVar, name, value string) bool {
	for _, pair := range env {
		if pair.Name == name && pair.Value == value {
			return true
		}
	}
	return false
}

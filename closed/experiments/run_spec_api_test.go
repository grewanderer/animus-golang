package main

import (
	"encoding/json"
	"testing"
)

const (
	validDigest   = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	validImageRef = "ghcr.io/acme/train@" + validDigest
)

func TestBuildRunSpecValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		req     createRunRequest
		wantErr error
	}{
		{
			name: "missing commit sha",
			req: createRunRequest{
				PipelineSpec:    rawSpec(minimalPipelineSpecJSON(validImageRef)),
				DatasetBindings: map[string]string{},
				CodeRef:         runSpecCodeRef{RepoURL: "https://github.com/acme/repo", CommitSHA: ""},
				EnvLock:         runSpecEnvLock{EnvHash: "envhash"},
			},
			wantErr: errInvalidRunSpec,
		},
		{
			name: "missing dataset binding",
			req: createRunRequest{
				PipelineSpec:    rawSpec(pipelineSpecWithDatasetRef(validImageRef, "training")),
				DatasetBindings: map[string]string{},
				CodeRef:         runSpecCodeRef{RepoURL: "https://github.com/acme/repo", CommitSHA: "deadbeef"},
				EnvLock:         runSpecEnvLock{EnvHash: "envhash"},
			},
			wantErr: errInvalidRunSpec,
		},
		{
			name: "non digest image",
			req: createRunRequest{
				PipelineSpec:    rawSpec(minimalPipelineSpecJSON("ghcr.io/acme/train:latest")),
				DatasetBindings: map[string]string{},
				CodeRef:         runSpecCodeRef{RepoURL: "https://github.com/acme/repo", CommitSHA: "deadbeef"},
				EnvLock:         runSpecEnvLock{EnvHash: "envhash"},
			},
			wantErr: errInvalidPipelineSpec,
		},
		{
			name: "duplicate step names",
			req: createRunRequest{
				PipelineSpec:    rawSpec(pipelineSpecWithDuplicateSteps(validImageRef)),
				DatasetBindings: map[string]string{},
				CodeRef:         runSpecCodeRef{RepoURL: "https://github.com/acme/repo", CommitSHA: "deadbeef"},
				EnvLock:         runSpecEnvLock{EnvHash: "envhash"},
			},
			wantErr: errInvalidPipelineSpec,
		},
		{
			name: "cycle detected",
			req: createRunRequest{
				PipelineSpec:    rawSpec(pipelineSpecWithCycle(validImageRef)),
				DatasetBindings: map[string]string{},
				CodeRef:         runSpecCodeRef{RepoURL: "https://github.com/acme/repo", CommitSHA: "deadbeef"},
				EnvLock:         runSpecEnvLock{EnvHash: "envhash"},
			},
			wantErr: errInvalidPipelineSpec,
		},
		{
			name: "missing dataset bindings object",
			req: createRunRequest{
				PipelineSpec: rawSpec(minimalPipelineSpecJSON(validImageRef)),
				CodeRef:      runSpecCodeRef{RepoURL: "https://github.com/acme/repo", CommitSHA: "deadbeef"},
				EnvLock:      runSpecEnvLock{EnvHash: "envhash"},
			},
			wantErr: errDatasetBindingsNeeded,
		},
	}

	for _, tc := range tests {
		if _, _, err := buildRunSpec("proj-1", "actor", tc.req); err != tc.wantErr {
			t.Fatalf("%s: expected err %v, got %v", tc.name, tc.wantErr, err)
		}
	}
}

func rawSpec(value string) json.RawMessage {
	return json.RawMessage(value)
}

func minimalPipelineSpecJSON(image string) string {
	return `{
  "apiVersion":"animus/v1alpha1",
  "kind":"Pipeline",
  "specVersion":"1.0",
  "spec":{
    "steps":[
      {
        "name":"step-a",
        "image":"` + image + `",
        "command":["echo"],
        "args":["ok"],
        "inputs":{"datasets":[],"artifacts":[]},
        "outputs":{"artifacts":[]},
        "env":[],
        "resources":{"cpu":"1","memory":"1Gi","gpu":0},
        "retryPolicy":{"maxAttempts":1,"backoff":{"type":"fixed","initialSeconds":0,"maxSeconds":0,"multiplier":1}}
      }
    ],
    "dependencies":[]
  }
}`
}

func pipelineSpecWithDatasetRef(image string, datasetRef string) string {
	return `{
  "apiVersion":"animus/v1alpha1",
  "kind":"Pipeline",
  "specVersion":"1.0",
  "spec":{
    "steps":[
      {
        "name":"step-a",
        "image":"` + image + `",
        "command":["echo"],
        "args":["ok"],
        "inputs":{"datasets":[{"name":"data","datasetRef":"` + datasetRef + `"}],"artifacts":[]},
        "outputs":{"artifacts":[]},
        "env":[],
        "resources":{"cpu":"1","memory":"1Gi","gpu":0},
        "retryPolicy":{"maxAttempts":1,"backoff":{"type":"fixed","initialSeconds":0,"maxSeconds":0,"multiplier":1}}
      }
    ],
    "dependencies":[]
  }
}`
}

func pipelineSpecWithDuplicateSteps(image string) string {
	return `{
  "apiVersion":"animus/v1alpha1",
  "kind":"Pipeline",
  "specVersion":"1.0",
  "spec":{
    "steps":[
      {
        "name":"step-a",
        "image":"` + image + `",
        "command":["echo"],
        "args":["one"],
        "inputs":{"datasets":[],"artifacts":[]},
        "outputs":{"artifacts":[]},
        "env":[],
        "resources":{"cpu":"1","memory":"1Gi","gpu":0},
        "retryPolicy":{"maxAttempts":1,"backoff":{"type":"fixed","initialSeconds":0,"maxSeconds":0,"multiplier":1}}
      },
      {
        "name":"step-a",
        "image":"` + image + `",
        "command":["echo"],
        "args":["two"],
        "inputs":{"datasets":[],"artifacts":[]},
        "outputs":{"artifacts":[]},
        "env":[],
        "resources":{"cpu":"1","memory":"1Gi","gpu":0},
        "retryPolicy":{"maxAttempts":1,"backoff":{"type":"fixed","initialSeconds":0,"maxSeconds":0,"multiplier":1}}
      }
    ],
    "dependencies":[]
  }
}`
}

func pipelineSpecWithCycle(image string) string {
	return `{
  "apiVersion":"animus/v1alpha1",
  "kind":"Pipeline",
  "specVersion":"1.0",
  "spec":{
    "steps":[
      {
        "name":"step-a",
        "image":"` + image + `",
        "command":["echo"],
        "args":["one"],
        "inputs":{"datasets":[],"artifacts":[]},
        "outputs":{"artifacts":[]},
        "env":[],
        "resources":{"cpu":"1","memory":"1Gi","gpu":0},
        "retryPolicy":{"maxAttempts":1,"backoff":{"type":"fixed","initialSeconds":0,"maxSeconds":0,"multiplier":1}}
      },
      {
        "name":"step-b",
        "image":"` + image + `",
        "command":["echo"],
        "args":["two"],
        "inputs":{"datasets":[],"artifacts":[]},
        "outputs":{"artifacts":[]},
        "env":[],
        "resources":{"cpu":"1","memory":"1Gi","gpu":0},
        "retryPolicy":{"maxAttempts":1,"backoff":{"type":"fixed","initialSeconds":0,"maxSeconds":0,"multiplier":1}}
      }
    ],
    "dependencies":[
      {"from":"step-a","to":"step-b"},
      {"from":"step-b","to":"step-a"}
    ]
  }
}`
}

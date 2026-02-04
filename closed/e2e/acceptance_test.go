//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
)

const (
	testPipelineDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testEnvDigest      = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

type authContext struct {
	secret  string
	subject string
	email   string
	roles   []string
}

func TestE2EAcceptance_ModelPromotion(t *testing.T) {
	infra := ensureInfra(t)
	repoRoot := repoRoot(t)
	tmpDir := t.TempDir()
	applyMigrations(t, repoRoot, infra.databaseURL)

	client := &http.Client{Timeout: 15 * time.Second}
	authCtx := authContext{
		secret:  infra.internalAuthSecret,
		subject: "e2e-user",
		email:   "e2e@example.com",
		roles:   []string{"admin"},
	}

	datasetAddr := freeAddr(t)
	datasetURL := "http://" + datasetAddr
	datasetEnv := []string{
		"DATASET_REGISTRY_HTTP_ADDR=" + datasetAddr,
		"DATABASE_URL=" + infra.databaseURL,
		"ANIMUS_INTERNAL_AUTH_SECRET=" + infra.internalAuthSecret,
		"ANIMUS_MINIO_ENDPOINT=" + infra.minioEndpoint,
		"ANIMUS_MINIO_ACCESS_KEY=" + infra.minioAccessKey,
		"ANIMUS_MINIO_SECRET_KEY=" + infra.minioSecretKey,
		"ANIMUS_MINIO_USE_SSL=false",
		"ANIMUS_MINIO_BUCKET_DATASETS=" + infra.minioBucketDatasets,
		"ANIMUS_MINIO_BUCKET_ARTIFACTS=" + infra.minioBucketArtifacts,
		"DATASET_REGISTRY_RETENTION_INTERVAL=1h",
	}
	datasetCmd := startService(t, repoRoot, tmpDir, "dataset-registry", "./closed/dataset-registry", datasetEnv)
	waitHTTP200(t, datasetURL+"/readyz")

	experimentsAddr := freeAddr(t)
	experimentsURL := "http://" + experimentsAddr
	experimentsEnv := []string{
		"EXPERIMENTS_HTTP_ADDR=" + experimentsAddr,
		"DATABASE_URL=" + infra.databaseURL,
		"ANIMUS_INTERNAL_AUTH_SECRET=" + infra.internalAuthSecret,
		"ANIMUS_CI_WEBHOOK_SECRET=" + infra.ciWebhookSecret,
		"ANIMUS_MINIO_ENDPOINT=" + infra.minioEndpoint,
		"ANIMUS_MINIO_ACCESS_KEY=" + infra.minioAccessKey,
		"ANIMUS_MINIO_SECRET_KEY=" + infra.minioSecretKey,
		"ANIMUS_MINIO_USE_SSL=false",
		"ANIMUS_MINIO_BUCKET_DATASETS=" + infra.minioBucketDatasets,
		"ANIMUS_MINIO_BUCKET_ARTIFACTS=" + infra.minioBucketArtifacts,
		"ANIMUS_TRAINING_EXECUTOR=disabled",
		"ANIMUS_EVALUATION_ENABLED=false",
		"ANIMUS_SCHEDULER_INTERVAL=1h",
		"ANIMUS_DP_RECONCILE_INTERVAL=1h",
		"EXPERIMENTS_RETENTION_INTERVAL=1h",
	}
	experimentsCmd := startService(t, repoRoot, tmpDir, "experiments", "./closed/experiments", experimentsEnv)
	waitHTTP200(t, experimentsURL+"/readyz")

	projectID := createProject(t, client, datasetURL, authCtx)
	datasetID := createDataset(t, client, datasetURL, authCtx, projectID)
	datasetVersionID := uploadDatasetVersion(t, client, datasetURL, authCtx, projectID, datasetID)

	envDefID := createEnvironmentDefinition(t, client, experimentsURL, authCtx, projectID)
	envLockID := createEnvironmentLock(t, client, experimentsURL, authCtx, projectID, envDefID)

	runID := createRun(t, client, experimentsURL, authCtx, projectID, envLockID, datasetVersionID)
	postHeartbeat(t, client, experimentsURL, authCtx, projectID, runID)
	postTerminal(t, client, experimentsURL, authCtx, projectID, runID)

	artifactID := createRunArtifact(t, client, experimentsURL, authCtx, runID)

	modelID := createModel(t, client, experimentsURL, authCtx, projectID)
	modelVersionID := createModelVersion(t, client, experimentsURL, authCtx, projectID, modelID, runID, artifactID)

	transitionModelVersion(t, client, experimentsURL, authCtx, projectID, modelVersionID, "validate")
	transitionModelVersion(t, client, experimentsURL, authCtx, projectID, modelVersionID, "approve")
	exportModelVersion(t, client, experimentsURL, authCtx, projectID, modelVersionID)
	transitionModelVersion(t, client, experimentsURL, authCtx, projectID, modelVersionID, "deprecate")

	_ = datasetCmd
	_ = experimentsCmd
}

func startService(t *testing.T, repoRoot, tmpDir, name, path string, env []string) *exec.Cmd {
	t.Helper()
	bin := filepath.Join(tmpDir, fmt.Sprintf("%s.bin", name))
	build := exec.Command("go", "build", "-o", bin, path)
	build.Dir = repoRoot
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build %s: %v\n%s", path, err, string(out))
	}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), env...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", name, err)
	}
	t.Cleanup(func() { stopProcess(t, cmd, &buf) })
	return cmd
}

func applyMigrations(t *testing.T, repoRoot, databaseURL string) {
	t.Helper()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
  version BIGINT PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}

	migrationsDir := filepath.Join(repoRoot, "closed", "migrations")
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	sort.Strings(files)
	for _, file := range files {
		base := filepath.Base(file)
		verRaw := strings.SplitN(base, "_", 2)[0]
		if verRaw == "" {
			continue
		}
		ver, err := strconv.Atoi(verRaw)
		if err != nil {
			t.Fatalf("invalid migration version %s: %v", base, err)
		}
		var exists int
		if err := db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = $1 LIMIT 1", ver).Scan(&exists); err == nil && exists == 1 {
			continue
		}
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		if _, err := db.Exec(string(raw)); err != nil {
			t.Fatalf("apply migration %s: %v", base, err)
		}
		if _, err := db.Exec("INSERT INTO schema_migrations(version, name) VALUES ($1, $2)", ver, strings.TrimSuffix(base, ".up.sql")); err != nil {
			t.Fatalf("record migration %s: %v", base, err)
		}
	}
}

func createProject(t *testing.T, client *http.Client, baseURL string, authCtx authContext) string {
	payload := map[string]any{"name": "e2e-project"}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/projects", payload, authCtx, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create project status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var out struct {
		ProjectID string `json:"project_id"`
	}
	decodeJSONBody(t, resp.Body, &out)
	if out.ProjectID == "" {
		t.Fatalf("project id missing")
	}
	return out.ProjectID
}

func createDataset(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID string) string {
	payload := map[string]any{"name": "e2e-dataset"}
	resp := doJSON(t, client, http.MethodPost, baseURL+"/datasets", payload, authCtx, projectID)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create dataset status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var out struct {
		DatasetID string `json:"dataset_id"`
	}
	decodeJSONBody(t, resp.Body, &out)
	if out.DatasetID == "" {
		t.Fatalf("dataset id missing")
	}
	return out.DatasetID
}

func uploadDatasetVersion(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, datasetID string) string {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	file, err := writer.CreateFormFile("file", "dataset.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := file.Write([]byte("e2e dataset")); err != nil {
		t.Fatalf("write dataset: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	url := fmt.Sprintf("%s/datasets/%s/versions/upload", baseURL, datasetID)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	signRequest(t, req, authCtx)
	req.Header.Set("X-Project-Id", projectID)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("upload dataset version: %v", err)
	}
	defer resp.Body.Close()
	body := readBody(t, resp.Body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload dataset version status=%d body=%s", resp.StatusCode, body)
	}
	var out struct {
		VersionID string `json:"version_id"`
	}
	decodeJSONBody(t, body, &out)
	if out.VersionID == "" {
		t.Fatalf("dataset version id missing")
	}
	return out.VersionID
}

func createEnvironmentDefinition(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID string) string {
	payload := map[string]any{
		"name": "e2e-env",
		"baseImages": []map[string]any{
			{"name": "python", "ref": "python:3.11"},
		},
	}
	url := fmt.Sprintf("%s/projects/%s/environment-definitions", baseURL, projectID)
	resp := doJSONWithIdempotency(t, client, http.MethodPost, url, payload, authCtx, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create env def status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var out struct {
		Definition struct {
			ID string `json:"environmentDefinitionId"`
		} `json:"definition"`
	}
	decodeJSONBody(t, resp.Body, &out)
	if out.Definition.ID == "" {
		t.Fatalf("env definition id missing")
	}
	return out.Definition.ID
}

func createEnvironmentLock(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, defID string) string {
	payload := map[string]any{
		"environmentDefinitionId": defID,
		"imageDigests":            map[string]string{"python": testEnvDigest},
	}
	url := fmt.Sprintf("%s/projects/%s/environment-locks", baseURL, projectID)
	resp := doJSONWithIdempotency(t, client, http.MethodPost, url, payload, authCtx, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create env lock status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var out struct {
		Lock struct {
			LockID string `json:"lockId"`
		} `json:"lock"`
	}
	decodeJSONBody(t, resp.Body, &out)
	if out.Lock.LockID == "" {
		t.Fatalf("env lock id missing")
	}
	return out.Lock.LockID
}

func createRun(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, lockID, datasetVersionID string) string {
	pipelineSpec := map[string]any{
		"apiVersion":  "animus/v1alpha1",
		"kind":        "Pipeline",
		"specVersion": "1.0",
		"metadata": map[string]any{
			"name": "e2e-pipeline",
		},
		"spec": map[string]any{
			"steps": []map[string]any{
				{
					"name":    "train",
					"image":   "ghcr.io/acme/train@" + testPipelineDigest,
					"command": []string{"echo"},
					"args":    []string{"ok"},
					"inputs": map[string]any{
						"datasets": []map[string]any{
							{"name": "dataset", "datasetRef": "dataset"},
						},
						"artifacts": []any{},
					},
					"outputs": map[string]any{
						"artifacts": []any{},
					},
					"env":       []any{},
					"resources": map[string]any{"cpu": "1", "memory": "1Gi", "gpu": 0},
					"retryPolicy": map[string]any{
						"maxAttempts": 1,
						"backoff": map[string]any{
							"type":           "fixed",
							"initialSeconds": 0,
							"maxSeconds":     0,
							"multiplier":     1,
						},
					},
				},
			},
			"dependencies": []any{},
		},
	}

	payload := map[string]any{
		"pipelineSpec":    pipelineSpec,
		"datasetBindings": map[string]string{"dataset": datasetVersionID},
		"codeRef": map[string]any{
			"repoUrl":   "https://github.com/animus-labs/animus-go",
			"commitSha": "deadbeef",
			"scmType":   "git",
		},
		"envLock":    map[string]any{"lockId": lockID},
		"parameters": map[string]any{},
	}
	url := fmt.Sprintf("%s/projects/%s/runs", baseURL, projectID)
	resp := doJSONWithIdempotency(t, client, http.MethodPost, url, payload, authCtx, "", "")
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("create run status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var out struct {
		RunID string `json:"runId"`
	}
	decodeJSONBody(t, resp.Body, &out)
	if out.RunID == "" {
		t.Fatalf("run id missing")
	}
	return out.RunID
}

func postHeartbeat(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, runID string) {
	payload := map[string]any{
		"eventId":   fmt.Sprintf("hb-%d", time.Now().UnixNano()),
		"runId":     runID,
		"projectId": projectID,
		"emittedAt": time.Now().UTC().Format(time.RFC3339Nano),
	}
	url := fmt.Sprintf("%s/internal/cp/runs/%s/heartbeat", baseURL, runID)
	resp := doJSON(t, client, http.MethodPost, url, payload, authCtx, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%s", resp.StatusCode, resp.Body)
	}
}

func postTerminal(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, runID string) {
	now := time.Now().UTC()
	payload := map[string]any{
		"eventId":    fmt.Sprintf("term-%d", time.Now().UnixNano()),
		"runId":      runID,
		"projectId":  projectID,
		"state":      "succeeded",
		"emittedAt":  now.Format(time.RFC3339Nano),
		"finishedAt": now.Format(time.RFC3339Nano),
	}
	url := fmt.Sprintf("%s/internal/cp/runs/%s/terminal", baseURL, runID)
	resp := doJSON(t, client, http.MethodPost, url, payload, authCtx, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("terminal status=%d body=%s", resp.StatusCode, resp.Body)
	}
}

func createRunArtifact(t *testing.T, client *http.Client, baseURL string, authCtx authContext, runID string) string {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("kind", "model")
	file, err := writer.CreateFormFile("file", "model.bin")
	if err != nil {
		t.Fatalf("create artifact form: %v", err)
	}
	if _, err := file.Write([]byte("model artifact")); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close artifact writer: %v", err)
	}

	url := fmt.Sprintf("%s/experiment-runs/%s/artifacts", baseURL, runID)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		t.Fatalf("new artifact request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	signRequest(t, req, authCtx)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create artifact: %v", err)
	}
	defer resp.Body.Close()
	body := readBody(t, resp.Body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create artifact status=%d body=%s", resp.StatusCode, body)
	}
	var out struct {
		Artifact struct {
			ArtifactID string `json:"artifact_id"`
		} `json:"artifact"`
	}
	decodeJSONBody(t, body, &out)
	if out.Artifact.ArtifactID == "" {
		t.Fatalf("artifact id missing")
	}
	return out.Artifact.ArtifactID
}

func createModel(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID string) string {
	payload := map[string]any{"name": "e2e-model"}
	url := fmt.Sprintf("%s/projects/%s/models", baseURL, projectID)
	resp := doJSONWithIdempotency(t, client, http.MethodPost, url, payload, authCtx, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create model status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var out struct {
		Model struct {
			ModelID string `json:"modelId"`
		} `json:"model"`
	}
	decodeJSONBody(t, resp.Body, &out)
	if out.Model.ModelID == "" {
		t.Fatalf("model id missing")
	}
	return out.Model.ModelID
}

func createModelVersion(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, modelID, runID, artifactID string) string {
	payload := map[string]any{
		"version":     "v1",
		"runId":       runID,
		"artifactIds": []string{artifactID},
	}
	url := fmt.Sprintf("%s/projects/%s/models/%s/versions", baseURL, projectID, modelID)
	resp := doJSONWithIdempotency(t, client, http.MethodPost, url, payload, authCtx, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create model version status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var out struct {
		ModelVersion struct {
			ModelVersionID string `json:"modelVersionId"`
		} `json:"modelVersion"`
	}
	decodeJSONBody(t, resp.Body, &out)
	if out.ModelVersion.ModelVersionID == "" {
		t.Fatalf("model version id missing")
	}
	return out.ModelVersion.ModelVersionID
}

func transitionModelVersion(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, versionID, action string) {
	url := fmt.Sprintf("%s/projects/%s/model-versions/%s:%s", baseURL, projectID, versionID, action)
	resp := doJSON(t, client, http.MethodPost, url, map[string]any{}, authCtx, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("model version %s status=%d body=%s", action, resp.StatusCode, resp.Body)
	}
}

func exportModelVersion(t *testing.T, client *http.Client, baseURL string, authCtx authContext, projectID, versionID string) {
	url := fmt.Sprintf("%s/projects/%s/model-versions/%s:export", baseURL, projectID, versionID)
	resp := doJSONWithIdempotency(t, client, http.MethodPost, url, map[string]any{}, authCtx, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("model export status=%d body=%s", resp.StatusCode, resp.Body)
	}
}

type httpResponse struct {
	StatusCode int
	Body       []byte
}

func doJSON(t *testing.T, client *http.Client, method, url string, payload any, authCtx authContext, projectID string) httpResponse {
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if projectID != "" {
		req.Header.Set("X-Project-Id", projectID)
	}
	signRequest(t, req, authCtx)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	body := readBody(t, resp.Body)
	return httpResponse{StatusCode: resp.StatusCode, Body: body}
}

func doJSONWithIdempotency(t *testing.T, client *http.Client, method, url string, payload any, authCtx authContext, projectID, idempotencyKey string) httpResponse {
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if projectID != "" {
		req.Header.Set("X-Project-Id", projectID)
	}
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("idem-%d", time.Now().UnixNano())
	}
	req.Header.Set("Idempotency-Key", idempotencyKey)
	signRequest(t, req, authCtx)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	body := readBody(t, resp.Body)
	return httpResponse{StatusCode: resp.StatusCode, Body: body}
}

func signRequest(t *testing.T, req *http.Request, authCtx authContext) {
	roles := strings.Join(authCtx.roles, ",")
	reqID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	req.Header.Set("X-Request-Id", reqID)
	req.Header.Set(auth.HeaderSubject, authCtx.subject)
	req.Header.Set(auth.HeaderEmail, authCtx.email)
	req.Header.Set(auth.HeaderRoles, roles)
	req.Header.Set(auth.HeaderInternalAuthTimestamp, ts)

	sig, err := auth.ComputeInternalAuthSignature(authCtx.secret, ts, req.Method, req.URL.Path, reqID, authCtx.subject, authCtx.email, roles)
	if err != nil {
		t.Fatalf("compute signature: %v", err)
	}
	req.Header.Set(auth.HeaderInternalAuthSignature, sig)
}

func decodeJSONBody(t *testing.T, data []byte, out any) {
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func readBody(t *testing.T, r io.Reader) []byte {
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return data
}

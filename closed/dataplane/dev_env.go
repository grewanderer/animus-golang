package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/dataplane"
	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/platform/k8s"
)

const devEnvTTLMinimumSeconds int64 = 60

func (api *dataplaneAPI) handleCreateDevEnv(w http.ResponseWriter, r *http.Request) {
	devEnvID := strings.TrimSpace(r.PathValue("dev_env_id"))
	if devEnvID == "" {
		writeError(w, http.StatusBadRequest, "dev_env_id_required", r.Header.Get("X-Request-Id"))
		return
	}

	var req dataplane.DevEnvProvisionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", r.Header.Get("X-Request-Id"))
		return
	}
	if strings.TrimSpace(req.DevEnvID) == "" || strings.TrimSpace(req.ProjectID) == "" || strings.TrimSpace(req.TemplateRef) == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", r.Header.Get("X-Request-Id"))
		return
	}
	if req.DevEnvID != devEnvID {
		writeError(w, http.StatusBadRequest, "dev_env_id_mismatch", r.Header.Get("X-Request-Id"))
		return
	}
	if req.EmittedAt.IsZero() {
		writeError(w, http.StatusBadRequest, "emitted_at_required", r.Header.Get("X-Request-Id"))
		return
	}
	if strings.TrimSpace(req.ImageRef) == "" {
		writeError(w, http.StatusBadRequest, "image_ref_required", r.Header.Get("X-Request-Id"))
		return
	}
	if req.TTLSeconds < devEnvTTLMinimumSeconds {
		writeError(w, http.StatusBadRequest, "ttl_seconds_too_small", r.Header.Get("X-Request-Id"))
		return
	}

	if err := validateEgressPolicy(api.cfg.EgressMode, domain.EnvLock{NetworkClassRef: req.NetworkClassRef}); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "network_policy_required", r.Header.Get("X-Request-Id"))
		return
	}

	jobName := jobNameForDevEnv(devEnvID)
	namespace := devEnvNamespace(api.cfg, api.k8s)

	job, err := buildDevEnvJobSpec(req, jobName, namespace, api.cfg.DevEnvServiceAccount, api.cfg.DevEnvTTLAfterFinishedSeconds)
	if err != nil {
		writeError(w, http.StatusConflict, "devenv_job_build_failed", r.Header.Get("X-Request-Id"))
		return
	}

	if err := api.k8s.CreateJob(r.Context(), namespace, job); err != nil && !errors.Is(err, k8s.ErrAlreadyExists) {
		writeError(w, http.StatusBadGateway, "devenv_job_create_failed", r.Header.Get("X-Request-Id"))
		return
	}

	writeJSON(w, http.StatusOK, dataplane.DevEnvProvisionResponse{
		DevEnvID:  devEnvID,
		ProjectID: req.ProjectID,
		Accepted:  true,
		JobName:   jobName,
		Namespace: namespace,
	})
}

func (api *dataplaneAPI) handleDeleteDevEnv(w http.ResponseWriter, r *http.Request) {
	devEnvID := strings.TrimSpace(r.PathValue("dev_env_id"))
	if devEnvID == "" {
		writeError(w, http.StatusBadRequest, "dev_env_id_required", r.Header.Get("X-Request-Id"))
		return
	}

	var req dataplane.DevEnvDeleteRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", r.Header.Get("X-Request-Id"))
		return
	}
	if strings.TrimSpace(req.DevEnvID) == "" || strings.TrimSpace(req.ProjectID) == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", r.Header.Get("X-Request-Id"))
		return
	}
	if req.DevEnvID != devEnvID {
		writeError(w, http.StatusBadRequest, "dev_env_id_mismatch", r.Header.Get("X-Request-Id"))
		return
	}
	if req.EmittedAt.IsZero() {
		writeError(w, http.StatusBadRequest, "emitted_at_required", r.Header.Get("X-Request-Id"))
		return
	}

	jobName := jobNameForDevEnv(devEnvID)
	namespace := devEnvNamespace(api.cfg, api.k8s)
	err := api.k8s.DeleteJob(r.Context(), namespace, jobName)
	if err != nil && !errors.Is(err, k8s.ErrNotFound) {
		writeError(w, http.StatusBadGateway, "devenv_job_delete_failed", r.Header.Get("X-Request-Id"))
		return
	}

	writeJSON(w, http.StatusOK, dataplane.DevEnvDeleteResponse{
		DevEnvID:  devEnvID,
		ProjectID: req.ProjectID,
		Deleted:   true,
	})
}

func (api *dataplaneAPI) handleAccessDevEnv(w http.ResponseWriter, r *http.Request) {
	devEnvID := strings.TrimSpace(r.PathValue("dev_env_id"))
	if devEnvID == "" {
		writeError(w, http.StatusBadRequest, "dev_env_id_required", r.Header.Get("X-Request-Id"))
		return
	}

	var req dataplane.DevEnvAccessRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", r.Header.Get("X-Request-Id"))
		return
	}
	if strings.TrimSpace(req.DevEnvID) == "" || strings.TrimSpace(req.ProjectID) == "" || strings.TrimSpace(req.SessionID) == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", r.Header.Get("X-Request-Id"))
		return
	}
	if req.DevEnvID != devEnvID {
		writeError(w, http.StatusBadRequest, "dev_env_id_mismatch", r.Header.Get("X-Request-Id"))
		return
	}
	if req.EmittedAt.IsZero() {
		writeError(w, http.StatusBadRequest, "emitted_at_required", r.Header.Get("X-Request-Id"))
		return
	}

	jobName := jobNameForDevEnv(devEnvID)
	namespace := devEnvNamespace(api.cfg, api.k8s)

	status, err := inspectJob(r.Context(), api.k8s, namespace, jobName)
	if err != nil {
		if errors.Is(err, errJobNotFound) {
			writeError(w, http.StatusNotFound, "not_found", r.Header.Get("X-Request-Id"))
			return
		}
		writeError(w, http.StatusBadGateway, "devenv_status_unavailable", r.Header.Get("X-Request-Id"))
		return
	}

	ready := status.State == jobStateRunning
	message := status.Reason
	if message == "" {
		message = status.State
	}

	writeJSON(w, http.StatusOK, dataplane.DevEnvAccessResponse{
		DevEnvID:  devEnvID,
		ProjectID: req.ProjectID,
		Ready:     ready,
		JobName:   jobName,
		Namespace: namespace,
		Message:   message,
	})
}

func buildDevEnvJobSpec(req dataplane.DevEnvProvisionRequest, jobName, namespace, serviceAccount string, ttlAfterFinishedSeconds int32) (k8s.Job, error) {
	if strings.TrimSpace(req.ImageRef) == "" {
		return k8s.Job{}, errors.New("image ref is required")
	}
	if strings.TrimSpace(jobName) == "" {
		return k8s.Job{}, errors.New("job name is required")
	}
	if strings.TrimSpace(namespace) == "" {
		return k8s.Job{}, errors.New("namespace is required")
	}

	labels := map[string]string{
		"app.kubernetes.io/name":      "animus-dataplane",
		"app.kubernetes.io/component": "devenv",
		"animus.dev_env_id":           strings.TrimSpace(req.DevEnvID),
		"animus.project_id":           strings.TrimSpace(req.ProjectID),
		"animus.template_ref":         strings.TrimSpace(req.TemplateRef),
	}
	if strings.TrimSpace(req.NetworkClassRef) != "" {
		labels["animus.network_class_ref"] = strings.TrimSpace(req.NetworkClassRef)
	}
	if strings.TrimSpace(req.SecretAccessClassRef) != "" {
		labels["animus.secret_access_class_ref"] = strings.TrimSpace(req.SecretAccessClassRef)
	}
	labels = filterLabelLength(labels)

	container := k8s.Container{
		Name:      "devenv",
		Image:     strings.TrimSpace(req.ImageRef),
		Resources: buildResourceRequirements(req.ResourceDefaults, req.ResourceLimits),
		Env: []k8s.EnvVar{
			{Name: "ANIMUS_DEV_ENV_ID", Value: strings.TrimSpace(req.DevEnvID)},
			{Name: "ANIMUS_PROJECT_ID", Value: strings.TrimSpace(req.ProjectID)},
			{Name: "ANIMUS_TEMPLATE_REF", Value: strings.TrimSpace(req.TemplateRef)},
			{Name: "ANIMUS_NETWORK_CLASS_REF", Value: strings.TrimSpace(req.NetworkClassRef)},
			{Name: "ANIMUS_SECRET_ACCESS_CLASS_REF", Value: strings.TrimSpace(req.SecretAccessClassRef)},
			{Name: "ANIMUS_DEV_ENV_TTL_SECONDS", Value: strconv.FormatInt(req.TTLSeconds, 10)},
		},
	}

	podSpec := k8s.PodSpec{
		RestartPolicy: "Never",
		Containers:    []k8s.Container{container},
	}
	if strings.TrimSpace(serviceAccount) != "" {
		podSpec.ServiceAccountName = strings.TrimSpace(serviceAccount)
	}

	backoff := int32(0)
	var ttl *int32
	if ttlAfterFinishedSeconds > 0 {
		ttl = &ttlAfterFinishedSeconds
	}
	var activeDeadline *int64
	if req.TTLSeconds > 0 {
		ttlSeconds := req.TTLSeconds
		activeDeadline = &ttlSeconds
	}

	job := k8s.Job{
		Metadata: k8s.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: k8s.JobSpec{
			BackoffLimit:            &backoff,
			TTLSecondsAfterFinished: ttl,
			ActiveDeadlineSeconds:   activeDeadline,
			Template: k8s.PodTemplateSpec{
				Metadata: k8s.ObjectMeta{Labels: labels},
				Spec:     podSpec,
			},
		},
	}
	return job, nil
}

func devEnvNamespace(cfg dataplaneConfig, client *k8s.Client) string {
	if strings.TrimSpace(cfg.DevEnvNamespace) != "" {
		return strings.TrimSpace(cfg.DevEnvNamespace)
	}
	if strings.TrimSpace(cfg.Namespace) != "" {
		return strings.TrimSpace(cfg.Namespace)
	}
	if client == nil {
		return ""
	}
	return strings.TrimSpace(client.Namespace())
}

func jobNameForDevEnv(devEnvID string) string {
	base := "animus-devenv-" + sanitizeName(devEnvID)
	if len(base) <= 63 {
		return base
	}
	sum := sha256.Sum256([]byte(devEnvID))
	suffix := hex.EncodeToString(sum[:6])
	trim := 63 - len(suffix) - 1
	if trim < 1 {
		trim = 1
	}
	return base[:trim] + "-" + suffix
}

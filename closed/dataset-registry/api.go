package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/closed/internal/platform/auth"
	"github.com/animus-labs/animus-go/closed/internal/platform/lineageevent"
	"github.com/animus-labs/animus-go/closed/internal/platform/objectstore"
	"github.com/animus-labs/animus-go/closed/internal/repo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
)

type datasetRegistryAPI struct {
	logger         *slog.Logger
	db             *sql.DB
	store          *minio.Client
	storeCfg       objectstore.Config
	uploadMaxBytes int64
	uploadTimeout  time.Duration
	svc            *datasetService
}

func newDatasetRegistryAPI(logger *slog.Logger, db *sql.DB, store *minio.Client, storeCfg objectstore.Config, uploadMaxBytes int64, uploadTimeout time.Duration, svc *datasetService) *datasetRegistryAPI {
	if uploadMaxBytes <= 0 {
		uploadMaxBytes = int64(2) << 30 // 2 GiB
	}
	if uploadTimeout <= 0 {
		uploadTimeout = 30 * time.Minute
	}
	return &datasetRegistryAPI{
		logger:         logger,
		db:             db,
		store:          store,
		storeCfg:       storeCfg,
		uploadMaxBytes: uploadMaxBytes,
		uploadTimeout:  uploadTimeout,
		svc:            svc,
	}
}

func (api *datasetRegistryAPI) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /projects", api.handleCreateProject)
	mux.HandleFunc("GET /projects/{project_id}", api.handleGetProject)

	mux.HandleFunc("GET /datasets", api.handleListDatasets)
	mux.HandleFunc("POST /datasets", api.handleCreateDataset)
	mux.HandleFunc("GET /datasets/{dataset_id}", api.handleGetDataset)

	mux.HandleFunc("GET /datasets/{dataset_id}/versions", api.handleListDatasetVersions)
	mux.HandleFunc("POST /datasets/{dataset_id}/versions/upload", api.handleUploadDatasetVersion)

	mux.HandleFunc("GET /dataset-versions/{version_id}", api.handleGetDatasetVersion)
	mux.HandleFunc("GET /dataset-versions/{version_id}/download", api.handleDownloadDatasetVersion)
}

type dataset struct {
	DatasetID   string          `json:"dataset_id"`
	ProjectID   string          `json:"project_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   string          `json:"created_by"`
}

type datasetVersion struct {
	VersionID     string          `json:"version_id"`
	DatasetID     string          `json:"dataset_id"`
	ProjectID     string          `json:"project_id"`
	QualityRuleID string          `json:"quality_rule_id,omitempty"`
	Ordinal       int64           `json:"ordinal"`
	ContentSHA256 string          `json:"content_sha256"`
	ObjectKey     string          `json:"object_key"`
	SizeBytes     int64           `json:"size_bytes,omitempty"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
	CreatedBy     string          `json:"created_by"`
}

type project struct {
	ProjectID   string          `json:"project_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   string          `json:"created_by"`
}

type createDatasetRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type createProjectRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func (api *datasetRegistryAPI) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}

	var req createProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_json")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		api.writeError(w, r, http.StatusBadRequest, "name_required")
		return
	}

	projectID, _ := auth.ProjectIDFromContext(r.Context())
	proj, err := api.svc.CreateProject(r.Context(), projectID, name, strings.TrimSpace(req.Description), req.Metadata, buildAuditContext(r, identity))
	if err != nil {
		if isUniqueViolation(err) {
			api.writeError(w, r, http.StatusConflict, "project_name_exists")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	metaJSON, _ := json.Marshal(proj.Metadata)
	w.Header().Set("Location", "/projects/"+proj.ID)
	api.writeJSON(w, http.StatusCreated, project{
		ProjectID:   proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Metadata:    metaJSON,
		CreatedAt:   proj.CreatedAt,
		CreatedBy:   proj.CreatedBy,
	})
}

func (api *datasetRegistryAPI) handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.PathValue("project_id"))
	if projectID == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}

	proj, err := api.svc.GetProject(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	metaJSON, _ := json.Marshal(proj.Metadata)
	api.writeJSON(w, http.StatusOK, project{
		ProjectID:   proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Metadata:    metaJSON,
		CreatedAt:   proj.CreatedAt,
		CreatedBy:   proj.CreatedBy,
	})
}

func (api *datasetRegistryAPI) handleCreateDataset(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}

	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok || strings.TrimSpace(projectID) == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}

	var req createDatasetRequest
	if err := decodeJSON(r, &req); err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_json")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		api.writeError(w, r, http.StatusBadRequest, "name_required")
		return
	}
	description := strings.TrimSpace(req.Description)

	ds, err := api.svc.CreateDataset(r.Context(), projectID, name, description, req.Metadata, buildAuditContext(r, identity))
	if err != nil {
		if isUniqueViolation(err) {
			api.writeError(w, r, http.StatusConflict, "dataset_name_exists")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	metadataJSON, _ := json.Marshal(ds.Metadata)
	w.Header().Set("Location", "/datasets/"+ds.ID)
	api.writeJSON(w, http.StatusCreated, dataset{
		DatasetID:   ds.ID,
		ProjectID:   ds.ProjectID,
		Name:        ds.Name,
		Description: ds.Description,
		Metadata:    metadataJSON,
		CreatedAt:   ds.CreatedAt,
		CreatedBy:   ds.CreatedBy,
	})
}

func (api *datasetRegistryAPI) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	limit := clampInt(parseIntQuery(r, "limit", 100), 1, 500)
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok || strings.TrimSpace(projectID) == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}

	items, err := api.svc.ListDatasets(r.Context(), projectID, limit)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	out := make([]dataset, 0, len(items))
	for _, item := range items {
		metaJSON, _ := json.Marshal(item.Metadata)
		out = append(out, dataset{
			DatasetID:   item.ID,
			ProjectID:   item.ProjectID,
			Name:        item.Name,
			Description: item.Description,
			Metadata:    metaJSON,
			CreatedAt:   item.CreatedAt,
			CreatedBy:   item.CreatedBy,
		})
	}

	api.writeJSON(w, http.StatusOK, map[string]any{"datasets": out})
}

func (api *datasetRegistryAPI) handleGetDataset(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimSpace(r.PathValue("dataset_id"))
	if datasetID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_id_required")
		return
	}
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok || strings.TrimSpace(projectID) == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}

	item, err := api.svc.GetDataset(r.Context(), projectID, datasetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	metaJSON, _ := json.Marshal(item.Metadata)
	api.writeJSON(w, http.StatusOK, dataset{
		DatasetID:   item.ID,
		ProjectID:   item.ProjectID,
		Name:        item.Name,
		Description: item.Description,
		Metadata:    metaJSON,
		CreatedAt:   item.CreatedAt,
		CreatedBy:   item.CreatedBy,
	})
}

func (api *datasetRegistryAPI) handleListDatasetVersions(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimSpace(r.PathValue("dataset_id"))
	if datasetID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_id_required")
		return
	}

	limit := clampInt(parseIntQuery(r, "limit", 100), 1, 500)
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok || strings.TrimSpace(projectID) == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}

	versions, err := api.svc.ListDatasetVersions(r.Context(), projectID, datasetID, limit)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	out := make([]datasetVersion, 0, len(versions))
	for _, version := range versions {
		metaJSON, _ := json.Marshal(version.Metadata)
		out = append(out, datasetVersion{
			VersionID:     version.ID,
			DatasetID:     version.DatasetID,
			ProjectID:     version.ProjectID,
			QualityRuleID: strings.TrimSpace(version.QualityRuleID),
			Ordinal:       version.Ordinal,
			ContentSHA256: version.ContentSHA256,
			ObjectKey:     version.ObjectKey,
			SizeBytes:     version.SizeBytes,
			Metadata:      metaJSON,
			CreatedAt:     version.CreatedAt,
			CreatedBy:     version.CreatedBy,
		})
	}

	api.writeJSON(w, http.StatusOK, map[string]any{"versions": out})
}

func (api *datasetRegistryAPI) handleUploadDatasetVersion(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimSpace(r.PathValue("dataset_id"))
	if datasetID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_id_required")
		return
	}
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok || strings.TrimSpace(projectID) == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}

	if r.ContentLength > 0 && r.ContentLength > api.uploadMaxBytes {
		api.writeErrorWithDetails(w, r, http.StatusRequestEntityTooLarge, "upload_too_large", map[string]any{
			"max_bytes":       api.uploadMaxBytes,
			"content_length":  r.ContentLength,
			"max_mebibytes":   api.uploadMaxBytes >> 20,
			"mebibytes":       r.ContentLength >> 20,
			"advice":          "increase DATASET_REGISTRY_UPLOAD_MAX_MIB or upload a smaller archive",
			"request_headers": map[string]any{"content_type": r.Header.Get("Content-Type")},
		})
		return
	}

	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err := api.svc.GetDataset(r.Context(), projectID, datasetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().UTC()
	versionID := uuid.NewString()

	r.Body = http.MaxBytesReader(w, r.Body, api.uploadMaxBytes)
	mr, err := r.MultipartReader()
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			api.writeErrorWithDetails(w, r, http.StatusRequestEntityTooLarge, "upload_too_large", map[string]any{
				"max_bytes":     api.uploadMaxBytes,
				"max_mebibytes": api.uploadMaxBytes >> 20,
			})
			return
		}
		api.writeError(w, r, http.StatusBadRequest, "invalid_multipart")
		return
	}

	metadataMap := map[string]any{}
	var (
		uploadedObjectKey string
		contentSHA256     string
		sizeBytes         int64
		filename          string
		contentType       string
		qualityRuleID     string
	)

	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				api.writeErrorWithDetails(w, r, http.StatusRequestEntityTooLarge, "upload_too_large", map[string]any{
					"max_bytes":     api.uploadMaxBytes,
					"max_mebibytes": api.uploadMaxBytes >> 20,
				})
				return
			}
			api.writeError(w, r, http.StatusBadRequest, "invalid_multipart")
			return
		}
		formName := part.FormName()
		switch formName {
		case "metadata":
			raw, err := io.ReadAll(io.LimitReader(part, 1<<20))
			_ = part.Close()
			if err != nil {
				var maxErr *http.MaxBytesError
				if errors.As(err, &maxErr) {
					api.writeErrorWithDetails(w, r, http.StatusRequestEntityTooLarge, "upload_too_large", map[string]any{
						"max_bytes":     api.uploadMaxBytes,
						"max_mebibytes": api.uploadMaxBytes >> 20,
					})
					return
				}
				api.writeError(w, r, http.StatusBadRequest, "invalid_metadata")
				return
			}
			raw = bytesTrimSpace(raw)
			if len(raw) == 0 {
				continue
			}
			if err := json.Unmarshal(raw, &metadataMap); err != nil {
				api.writeError(w, r, http.StatusBadRequest, "invalid_metadata")
				return
			}
		case "quality_rule_id":
			raw, err := io.ReadAll(io.LimitReader(part, 4096))
			_ = part.Close()
			if err != nil {
				var maxErr *http.MaxBytesError
				if errors.As(err, &maxErr) {
					api.writeErrorWithDetails(w, r, http.StatusRequestEntityTooLarge, "upload_too_large", map[string]any{
						"max_bytes":     api.uploadMaxBytes,
						"max_mebibytes": api.uploadMaxBytes >> 20,
					})
					return
				}
				api.writeError(w, r, http.StatusBadRequest, "invalid_quality_rule_id")
				return
			}
			qualityRuleID = strings.TrimSpace(string(raw))
		case "file":
			if uploadedObjectKey != "" {
				_ = part.Close()
				api.writeError(w, r, http.StatusBadRequest, "multiple_files_not_supported")
				return
			}

			filename = sanitizeFilename(part.FileName())
			contentType = strings.TrimSpace(part.Header.Get("Content-Type"))
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			uploadedObjectKey = fmt.Sprintf("%s/%s/%s", datasetID, versionID, filename)
			hasher := sha256.New()
			counter := &countingWriter{}
			reader := io.TeeReader(part, io.MultiWriter(hasher, counter))

			uploadCtx, cancel := context.WithTimeout(r.Context(), api.uploadTimeout)
			_, putErr := api.store.PutObject(
				uploadCtx,
				api.storeCfg.BucketDatasets,
				uploadedObjectKey,
				reader,
				-1,
				minio.PutObjectOptions{ContentType: contentType},
			)
			cancel()
			_ = part.Close()
			if putErr != nil {
				var maxErr *http.MaxBytesError
				if errors.As(putErr, &maxErr) {
					api.writeErrorWithDetails(w, r, http.StatusRequestEntityTooLarge, "upload_too_large", map[string]any{
						"max_bytes":     api.uploadMaxBytes,
						"max_mebibytes": api.uploadMaxBytes >> 20,
					})
					return
				}
				api.writeError(w, r, http.StatusBadRequest, "upload_failed")
				return
			}
			contentSHA256 = hex.EncodeToString(hasher.Sum(nil))
			sizeBytes = counter.n
		default:
			_ = part.Close()
		}
	}

	if uploadedObjectKey == "" {
		api.writeError(w, r, http.StatusBadRequest, "file_required")
		return
	}

	if qualityRuleID != "" {
		var exists string
		if err := api.db.QueryRowContext(r.Context(), `SELECT rule_id FROM quality_rules WHERE rule_id = $1`, qualityRuleID).Scan(&exists); err != nil {
			_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
			if errors.Is(err, sql.ErrNoRows) {
				api.writeError(w, r, http.StatusNotFound, "quality_rule_not_found")
				return
			}
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
	}

	metadataMap["filename"] = filename
	metadataMap["content_type"] = contentType
	metadataMap["content_sha256"] = contentSHA256
	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusBadRequest, "invalid_metadata")
		return
	}

	ordinal, err := api.svc.NextDatasetVersionOrdinal(r.Context(), projectID, datasetID)
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	version, err := api.svc.CreateDatasetVersion(r.Context(), domain.DatasetVersion{
		ID:            versionID,
		ProjectID:     projectID,
		DatasetID:     datasetID,
		QualityRuleID: qualityRuleID,
		Ordinal:       ordinal,
		ContentSHA256: contentSHA256,
		ObjectKey:     uploadedObjectKey,
		SizeBytes:     sizeBytes,
	}, metadataMap, buildAuditContext(r, identity))
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		if isUniqueViolation(err) {
			api.writeError(w, r, http.StatusConflict, "duplicate_content")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err = lineageevent.Insert(r.Context(), api.db, lineageevent.Event{
		OccurredAt:  now,
		Actor:       identity.Subject,
		RequestID:   r.Header.Get("X-Request-Id"),
		SubjectType: "dataset",
		SubjectID:   datasetID,
		Predicate:   "has_version",
		ObjectType:  "dataset_version",
		ObjectID:    version.ID,
		Metadata: map[string]any{
			"ordinal":         version.Ordinal,
			"content_sha256":  version.ContentSHA256,
			"quality_rule_id": version.QualityRuleID,
			"size_bytes":      version.SizeBytes,
			"object_key":      version.ObjectKey,
		},
	})
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "lineage_write_failed")
		return
	}

	w.Header().Set("Location", "/dataset-versions/"+version.ID)
	api.writeJSON(w, http.StatusCreated, datasetVersion{
		VersionID:     version.ID,
		DatasetID:     version.DatasetID,
		ProjectID:     version.ProjectID,
		QualityRuleID: version.QualityRuleID,
		Ordinal:       version.Ordinal,
		ContentSHA256: version.ContentSHA256,
		ObjectKey:     version.ObjectKey,
		SizeBytes:     version.SizeBytes,
		Metadata:      metadataJSON,
		CreatedAt:     version.CreatedAt,
		CreatedBy:     version.CreatedBy,
	})
}

func (api *datasetRegistryAPI) handleGetDatasetVersion(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.PathValue("version_id"))
	if versionID == "" {
		api.writeError(w, r, http.StatusBadRequest, "version_id_required")
		return
	}
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok || strings.TrimSpace(projectID) == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}

	version, err := api.svc.GetDatasetVersion(r.Context(), projectID, versionID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	metaJSON, _ := json.Marshal(version.Metadata)
	api.writeJSON(w, http.StatusOK, datasetVersion{
		VersionID:     version.ID,
		DatasetID:     version.DatasetID,
		ProjectID:     version.ProjectID,
		QualityRuleID: strings.TrimSpace(version.QualityRuleID),
		Ordinal:       version.Ordinal,
		ContentSHA256: version.ContentSHA256,
		ObjectKey:     version.ObjectKey,
		SizeBytes:     version.SizeBytes,
		Metadata:      metaJSON,
		CreatedAt:     version.CreatedAt,
		CreatedBy:     version.CreatedBy,
	})
}

func (api *datasetRegistryAPI) handleDownloadDatasetVersion(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.PathValue("version_id"))
	if versionID == "" {
		api.writeError(w, r, http.StatusBadRequest, "version_id_required")
		return
	}
	if api.svc == nil {
		api.writeError(w, r, http.StatusInternalServerError, "service_unavailable")
		return
	}
	projectID, ok := auth.ProjectIDFromContext(r.Context())
	if !ok || strings.TrimSpace(projectID) == "" {
		api.writeError(w, r, http.StatusBadRequest, "project_id_required")
		return
	}

	version, err := api.svc.GetDatasetVersion(r.Context(), projectID, versionID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	metaJSON, _ := json.Marshal(version.Metadata)
	meta := normalizeJSON(metaJSON)
	filename := jsonFieldString(meta, "filename")
	if filename == "" {
		filename = "dataset.bin"
	}
	contentType := jsonFieldString(meta, "content_type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().UTC()
	ruleID := strings.TrimSpace(version.QualityRuleID)
	if ruleID == "" {
		_, _ = auditlog.Insert(r.Context(), api.db, auditlog.Event{
			OccurredAt:   now,
			Actor:        identity.Subject,
			Action:       "quality_gate.block",
			ResourceType: "dataset_version",
			ResourceID:   versionID,
			RequestID:    r.Header.Get("X-Request-Id"),
			IP:           requestIP(r.RemoteAddr),
			UserAgent:    r.UserAgent(),
			Payload: map[string]any{
				"service":            "dataset-registry",
				"dataset_id":         version.DatasetID,
				"dataset_version_id": versionID,
				"reason":             "no_rule",
			},
		})
		api.writeError(w, r, http.StatusConflict, "quality_rule_not_set")
		return
	}

	var (
		evalID     string
		evalStatus string
	)
	err = api.db.QueryRowContext(
		r.Context(),
		`SELECT evaluation_id, status
		 FROM quality_evaluations
		 WHERE dataset_version_id = $1 AND rule_id = $2
		 ORDER BY evaluated_at DESC
		 LIMIT 1`,
		versionID,
		ruleID,
	).Scan(&evalID, &evalStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, _ = auditlog.Insert(r.Context(), api.db, auditlog.Event{
				OccurredAt:   now,
				Actor:        identity.Subject,
				Action:       "quality_gate.block",
				ResourceType: "dataset_version",
				ResourceID:   versionID,
				RequestID:    r.Header.Get("X-Request-Id"),
				IP:           requestIP(r.RemoteAddr),
				UserAgent:    r.UserAgent(),
				Payload: map[string]any{
					"service":            "dataset-registry",
					"dataset_id":         version.DatasetID,
					"dataset_version_id": versionID,
					"rule_id":            ruleID,
					"reason":             "not_evaluated",
				},
			})
			api.writeError(w, r, http.StatusConflict, "quality_not_evaluated")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	if strings.ToLower(strings.TrimSpace(evalStatus)) != "pass" {
		_, _ = auditlog.Insert(r.Context(), api.db, auditlog.Event{
			OccurredAt:   now,
			Actor:        identity.Subject,
			Action:       "quality_gate.block",
			ResourceType: "dataset_version",
			ResourceID:   versionID,
			RequestID:    r.Header.Get("X-Request-Id"),
			IP:           requestIP(r.RemoteAddr),
			UserAgent:    r.UserAgent(),
			Payload: map[string]any{
				"service":            "dataset-registry",
				"dataset_id":         version.DatasetID,
				"dataset_version_id": versionID,
				"rule_id":            ruleID,
				"evaluation_id":      evalID,
				"status":             evalStatus,
				"reason":             "not_pass",
			},
		})
		api.writeError(w, r, http.StatusConflict, "quality_gate_failed")
		return
	}

	_, _ = auditlog.Insert(r.Context(), api.db, auditlog.Event{
		OccurredAt:   now,
		Actor:        identity.Subject,
		Action:       "quality_gate.allow",
		ResourceType: "dataset_version",
		ResourceID:   versionID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":            "dataset-registry",
			"dataset_id":         version.DatasetID,
			"dataset_version_id": versionID,
			"rule_id":            ruleID,
			"evaluation_id":      evalID,
			"status":             evalStatus,
		},
	})

	obj, err := api.store.GetObject(r.Context(), api.storeCfg.BucketDatasets, version.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		api.writeError(w, r, http.StatusBadGateway, "object_store_error")
		return
	}
	defer obj.Close()

	if _, err := obj.Stat(); err != nil {
		api.writeError(w, r, http.StatusNotFound, "not_found")
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if version.SizeBytes > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(version.SizeBytes, 10))
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, obj)
}

type countingWriter struct {
	n int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	w.n += int64(len(p))
	return len(p), nil
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("multiple JSON values")
	}
	return nil
}

func buildAuditContext(r *http.Request, identity auth.Identity) auditContext {
	return auditContext{
		Actor:     strings.TrimSpace(identity.Subject),
		RequestID: r.Header.Get("X-Request-Id"),
		IP:        requestIP(r.RemoteAddr),
		UserAgent: r.UserAgent(),
		Path:      r.URL.Path,
		Service:   "dataset-registry",
	}
}

func (api *datasetRegistryAPI) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(body)
}

func (api *datasetRegistryAPI) writeError(w http.ResponseWriter, r *http.Request, status int, code string) {
	api.writeJSON(w, status, map[string]any{
		"error":      code,
		"request_id": r.Header.Get("X-Request-Id"),
	})
}

func (api *datasetRegistryAPI) writeErrorWithDetails(w http.ResponseWriter, r *http.Request, status int, code string, details any) {
	api.writeJSON(w, status, map[string]any{
		"error":      code,
		"request_id": r.Header.Get("X-Request-Id"),
		"details":    details,
	})
}

func requestIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}

func jsonFieldString(raw json.RawMessage, key string) string {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	v, ok := obj[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func normalizeJSON(raw []byte) json.RawMessage {
	raw = bytesTrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("{}")
	}
	return raw
}

func bytesTrimSpace(in []byte) []byte {
	return []byte(strings.TrimSpace(string(in)))
}

func sanitizeFilename(name string) string {
	base := path.Base(strings.TrimSpace(name))
	if base == "" || base == "." || base == "/" {
		return "dataset.bin"
	}
	return base
}

func nullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func parseIntQuery(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return parsed
}

func clampInt(v int, min int, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

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

	"github.com/animus-labs/animus-go/internal/platform/auditlog"
	"github.com/animus-labs/animus-go/internal/platform/auth"
	"github.com/animus-labs/animus-go/internal/platform/lineageevent"
	"github.com/animus-labs/animus-go/internal/platform/objectstore"
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
}

func newDatasetRegistryAPI(logger *slog.Logger, db *sql.DB, store *minio.Client, storeCfg objectstore.Config) *datasetRegistryAPI {
	return &datasetRegistryAPI{
		logger:         logger,
		db:             db,
		store:          store,
		storeCfg:       storeCfg,
		uploadMaxBytes: 250 << 20, // 250 MiB
	}
}

func (api *datasetRegistryAPI) register(mux *http.ServeMux) {
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
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   string          `json:"created_by"`
}

type datasetVersion struct {
	VersionID     string          `json:"version_id"`
	DatasetID     string          `json:"dataset_id"`
	QualityRuleID string          `json:"quality_rule_id,omitempty"`
	Ordinal       int64           `json:"ordinal"`
	ContentSHA256 string          `json:"content_sha256"`
	ObjectKey     string          `json:"object_key"`
	SizeBytes     int64           `json:"size_bytes,omitempty"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
	CreatedBy     string          `json:"created_by"`
}

type createDatasetRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func (api *datasetRegistryAPI) handleCreateDataset(w http.ResponseWriter, r *http.Request) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
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

	metadataMap := req.Metadata
	if metadataMap == nil {
		metadataMap = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		api.writeError(w, r, http.StatusBadRequest, "invalid_metadata")
		return
	}

	now := time.Now().UTC()
	datasetID := uuid.NewString()

	type integrityInput struct {
		DatasetID   string          `json:"dataset_id"`
		Name        string          `json:"name"`
		Description string          `json:"description,omitempty"`
		Metadata    json.RawMessage `json:"metadata"`
		CreatedAt   time.Time       `json:"created_at"`
		CreatedBy   string          `json:"created_by"`
	}
	integrity, err := integritySHA256(integrityInput{
		DatasetID:   datasetID,
		Name:        name,
		Description: description,
		Metadata:    metadataJSON,
		CreatedAt:   now,
		CreatedBy:   identity.Subject,
	})
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	tx, err := api.db.BeginTx(r.Context(), nil)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(
		r.Context(),
		`INSERT INTO datasets (
			dataset_id,
			name,
			description,
			metadata,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		datasetID,
		name,
		nullString(description),
		metadataJSON,
		now,
		identity.Subject,
		integrity,
	)
	if err != nil {
		if isUniqueViolation(err) {
			api.writeError(w, r, http.StatusConflict, "dataset_name_exists")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err = auditlog.Insert(r.Context(), tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        identity.Subject,
		Action:       "dataset.create",
		ResourceType: "dataset",
		ResourceID:   datasetID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":      "dataset-registry",
			"dataset_id":   datasetID,
			"name":         name,
			"description":  description,
			"metadata":     metadataMap,
			"created_by":   identity.Subject,
			"request_path": r.URL.Path,
		},
	})
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
		return
	}

	if err := tx.Commit(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Location", "/datasets/"+datasetID)
	api.writeJSON(w, http.StatusCreated, dataset{
		DatasetID:   datasetID,
		Name:        name,
		Description: description,
		Metadata:    metadataJSON,
		CreatedAt:   now,
		CreatedBy:   identity.Subject,
	})
}

func (api *datasetRegistryAPI) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	limit := clampInt(parseIntQuery(r, "limit", 100), 1, 500)

	rows, err := api.db.QueryContext(
		r.Context(),
		`SELECT dataset_id, name, description, metadata, created_at, created_by
		 FROM datasets
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer rows.Close()

	out := make([]dataset, 0, limit)
	for rows.Next() {
		var (
			datasetID   string
			name        string
			description sql.NullString
			metadata    []byte
			createdAt   time.Time
			createdBy   string
		)
		if err := rows.Scan(&datasetID, &name, &description, &metadata, &createdAt, &createdBy); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		metadata = normalizeJSON(metadata)
		out = append(out, dataset{
			DatasetID:   datasetID,
			Name:        name,
			Description: description.String,
			Metadata:    metadata,
			CreatedAt:   createdAt,
			CreatedBy:   createdBy,
		})
	}
	if err := rows.Err(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]any{"datasets": out})
}

func (api *datasetRegistryAPI) handleGetDataset(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimSpace(r.PathValue("dataset_id"))
	if datasetID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_id_required")
		return
	}

	var (
		name        string
		description sql.NullString
		metadata    []byte
		createdAt   time.Time
		createdBy   string
	)
	err := api.db.QueryRowContext(
		r.Context(),
		`SELECT name, description, metadata, created_at, created_by
		 FROM datasets
		 WHERE dataset_id = $1`,
		datasetID,
	).Scan(&name, &description, &metadata, &createdAt, &createdBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, dataset{
		DatasetID:   datasetID,
		Name:        name,
		Description: description.String,
		Metadata:    normalizeJSON(metadata),
		CreatedAt:   createdAt,
		CreatedBy:   createdBy,
	})
}

func (api *datasetRegistryAPI) handleListDatasetVersions(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimSpace(r.PathValue("dataset_id"))
	if datasetID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_id_required")
		return
	}

	limit := clampInt(parseIntQuery(r, "limit", 100), 1, 500)

	rows, err := api.db.QueryContext(
		r.Context(),
		`SELECT version_id, quality_rule_id, ordinal, content_sha256, object_key, size_bytes, metadata, created_at, created_by
		 FROM dataset_versions
		 WHERE dataset_id = $1
		 ORDER BY ordinal DESC
		 LIMIT $2`,
		datasetID,
		limit,
	)
	if err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer rows.Close()

	out := make([]datasetVersion, 0, limit)
	for rows.Next() {
		var (
			versionID     string
			qualityRuleID sql.NullString
			ordinal       int64
			contentSHA256 string
			objectKey     string
			sizeBytes     sql.NullInt64
			metadata      []byte
			createdAt     time.Time
			createdBy     string
		)
		if err := rows.Scan(&versionID, &qualityRuleID, &ordinal, &contentSHA256, &objectKey, &sizeBytes, &metadata, &createdAt, &createdBy); err != nil {
			api.writeError(w, r, http.StatusInternalServerError, "internal_error")
			return
		}
		out = append(out, datasetVersion{
			VersionID:     versionID,
			DatasetID:     datasetID,
			QualityRuleID: strings.TrimSpace(qualityRuleID.String),
			Ordinal:       ordinal,
			ContentSHA256: contentSHA256,
			ObjectKey:     objectKey,
			SizeBytes:     sizeBytes.Int64,
			Metadata:      normalizeJSON(metadata),
			CreatedAt:     createdAt,
			CreatedBy:     createdBy,
		})
	}
	if err := rows.Err(); err != nil {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]any{"versions": out})
}

func (api *datasetRegistryAPI) handleUploadDatasetVersion(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimSpace(r.PathValue("dataset_id"))
	if datasetID == "" {
		api.writeError(w, r, http.StatusBadRequest, "dataset_id_required")
		return
	}

	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok || strings.TrimSpace(identity.Subject) == "" {
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	now := time.Now().UTC()
	versionID := uuid.NewString()

	r.Body = http.MaxBytesReader(w, r.Body, api.uploadMaxBytes)
	mr, err := r.MultipartReader()
	if err != nil {
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
			api.writeError(w, r, http.StatusBadRequest, "invalid_multipart")
			return
		}
		formName := part.FormName()
		switch formName {
		case "metadata":
			raw, err := io.ReadAll(io.LimitReader(part, 1<<20))
			_ = part.Close()
			if err != nil {
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

			uploadCtx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
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

	type integrityInput struct {
		VersionID     string          `json:"version_id"`
		DatasetID     string          `json:"dataset_id"`
		QualityRuleID string          `json:"quality_rule_id,omitempty"`
		Ordinal       int64           `json:"ordinal"`
		ContentSHA256 string          `json:"content_sha256"`
		ObjectKey     string          `json:"object_key"`
		SizeBytes     int64           `json:"size_bytes"`
		Metadata      json.RawMessage `json:"metadata"`
		CreatedAt     time.Time       `json:"created_at"`
		CreatedBy     string          `json:"created_by"`
	}

	tx, err := api.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	var locked string
	if err := tx.QueryRowContext(
		r.Context(),
		`SELECT dataset_id FROM datasets WHERE dataset_id = $1 FOR UPDATE`,
		datasetID,
	).Scan(&locked); err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	var ordinal int64
	if err := tx.QueryRowContext(
		r.Context(),
		`SELECT COALESCE(MAX(ordinal), 0) FROM dataset_versions WHERE dataset_id = $1`,
		datasetID,
	).Scan(&ordinal); err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}
	ordinal++

	integrity, err := integritySHA256(integrityInput{
		VersionID:     versionID,
		DatasetID:     datasetID,
		QualityRuleID: qualityRuleID,
		Ordinal:       ordinal,
		ContentSHA256: contentSHA256,
		ObjectKey:     uploadedObjectKey,
		SizeBytes:     sizeBytes,
		Metadata:      metadataJSON,
		CreatedAt:     now,
		CreatedBy:     identity.Subject,
	})
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err = tx.ExecContext(
		r.Context(),
		`INSERT INTO dataset_versions (
			version_id,
			dataset_id,
			quality_rule_id,
			ordinal,
			content_sha256,
			object_key,
			size_bytes,
			metadata,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		versionID,
		datasetID,
		nullString(qualityRuleID),
		ordinal,
		contentSHA256,
		uploadedObjectKey,
		sizeBytes,
		metadataJSON,
		now,
		identity.Subject,
		integrity,
	)
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		if isUniqueViolation(err) {
			api.writeError(w, r, http.StatusConflict, "duplicate_content")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	_, err = auditlog.Insert(r.Context(), tx, auditlog.Event{
		OccurredAt:   now,
		Actor:        identity.Subject,
		Action:       "dataset_version.create",
		ResourceType: "dataset_version",
		ResourceID:   versionID,
		RequestID:    r.Header.Get("X-Request-Id"),
		IP:           requestIP(r.RemoteAddr),
		UserAgent:    r.UserAgent(),
		Payload: map[string]any{
			"service":         "dataset-registry",
			"dataset_id":      datasetID,
			"version_id":      versionID,
			"quality_rule_id": qualityRuleID,
			"ordinal":         ordinal,
			"content_sha256":  contentSHA256,
			"size_bytes":      sizeBytes,
			"object_key":      uploadedObjectKey,
		},
	})
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "audit_failed")
		return
	}

	_, err = lineageevent.Insert(r.Context(), tx, lineageevent.Event{
		OccurredAt:  now,
		Actor:       identity.Subject,
		RequestID:   r.Header.Get("X-Request-Id"),
		SubjectType: "dataset",
		SubjectID:   datasetID,
		Predicate:   "has_version",
		ObjectType:  "dataset_version",
		ObjectID:    versionID,
		Metadata: map[string]any{
			"ordinal":         ordinal,
			"content_sha256":  contentSHA256,
			"quality_rule_id": qualityRuleID,
			"size_bytes":      sizeBytes,
			"object_key":      uploadedObjectKey,
		},
	})
	if err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "lineage_write_failed")
		return
	}

	if err := tx.Commit(); err != nil {
		_ = api.store.RemoveObject(r.Context(), api.storeCfg.BucketDatasets, uploadedObjectKey, minio.RemoveObjectOptions{})
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	w.Header().Set("Location", "/dataset-versions/"+versionID)
	api.writeJSON(w, http.StatusCreated, datasetVersion{
		VersionID:     versionID,
		DatasetID:     datasetID,
		QualityRuleID: qualityRuleID,
		Ordinal:       ordinal,
		ContentSHA256: contentSHA256,
		ObjectKey:     uploadedObjectKey,
		SizeBytes:     sizeBytes,
		Metadata:      metadataJSON,
		CreatedAt:     now,
		CreatedBy:     identity.Subject,
	})
}

func (api *datasetRegistryAPI) handleGetDatasetVersion(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.PathValue("version_id"))
	if versionID == "" {
		api.writeError(w, r, http.StatusBadRequest, "version_id_required")
		return
	}

	var (
		datasetID     string
		qualityRuleID sql.NullString
		ordinal       int64
		contentSHA256 string
		objectKey     string
		sizeBytes     sql.NullInt64
		metadata      []byte
		createdAt     time.Time
		createdBy     string
	)
	err := api.db.QueryRowContext(
		r.Context(),
		`SELECT dataset_id, quality_rule_id, ordinal, content_sha256, object_key, size_bytes, metadata, created_at, created_by
		 FROM dataset_versions
		 WHERE version_id = $1`,
		versionID,
	).Scan(&datasetID, &qualityRuleID, &ordinal, &contentSHA256, &objectKey, &sizeBytes, &metadata, &createdAt, &createdBy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	api.writeJSON(w, http.StatusOK, datasetVersion{
		VersionID:     versionID,
		DatasetID:     datasetID,
		QualityRuleID: strings.TrimSpace(qualityRuleID.String),
		Ordinal:       ordinal,
		ContentSHA256: contentSHA256,
		ObjectKey:     objectKey,
		SizeBytes:     sizeBytes.Int64,
		Metadata:      normalizeJSON(metadata),
		CreatedAt:     createdAt,
		CreatedBy:     createdBy,
	})
}

func (api *datasetRegistryAPI) handleDownloadDatasetVersion(w http.ResponseWriter, r *http.Request) {
	versionID := strings.TrimSpace(r.PathValue("version_id"))
	if versionID == "" {
		api.writeError(w, r, http.StatusBadRequest, "version_id_required")
		return
	}

	var (
		datasetID     string
		qualityRuleID sql.NullString
		objectKey     string
		sizeBytes     sql.NullInt64
		metadata      []byte
	)
	err := api.db.QueryRowContext(
		r.Context(),
		`SELECT dataset_id, quality_rule_id, object_key, size_bytes, metadata
		 FROM dataset_versions
		 WHERE version_id = $1`,
		versionID,
	).Scan(&datasetID, &qualityRuleID, &objectKey, &sizeBytes, &metadata)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			api.writeError(w, r, http.StatusNotFound, "not_found")
			return
		}
		api.writeError(w, r, http.StatusInternalServerError, "internal_error")
		return
	}

	meta := normalizeJSON(metadata)
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
	ruleID := strings.TrimSpace(qualityRuleID.String)
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
				"dataset_id":         datasetID,
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
					"dataset_id":         datasetID,
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
				"dataset_id":         datasetID,
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
			"dataset_id":         datasetID,
			"dataset_version_id": versionID,
			"rule_id":            ruleID,
			"evaluation_id":      evalID,
			"status":             evalStatus,
		},
	})

	obj, err := api.store.GetObject(r.Context(), api.storeCfg.BucketDatasets, objectKey, minio.GetObjectOptions{})
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
	if sizeBytes.Valid && sizeBytes.Int64 > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(sizeBytes.Int64, 10))
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

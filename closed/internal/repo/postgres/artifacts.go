package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

type ArtifactStore struct {
	db DB
}

func NewArtifactStore(db DB) *ArtifactStore {
	if db == nil {
		return nil
	}
	return &ArtifactStore{db: db}
}

func (s *ArtifactStore) CreateArtifact(ctx context.Context, artifact domain.Artifact) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("artifact store not initialized")
	}
	if err := artifact.Validate(); err != nil {
		return err
	}
	if err := requireIntegrity(artifact.IntegritySHA256); err != nil {
		return err
	}
	metadataJSON, err := encodeMetadata(artifact.Metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	createdAt := normalizeTime(artifact.CreatedAt)
	var retentionUntil sql.NullTime
	if artifact.RetentionUntil != nil {
		retentionUntil = sql.NullTime{Time: artifact.RetentionUntil.UTC(), Valid: true}
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO experiment_run_artifacts (
			artifact_id,
			run_id,
			project_id,
			kind,
			name,
			filename,
			content_type,
			object_key,
			sha256,
			size_bytes,
			metadata,
			retention_until,
			retention_policy,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		strings.TrimSpace(artifact.ID),
		strings.TrimSpace(artifact.RunID),
		strings.TrimSpace(artifact.ProjectID),
		strings.TrimSpace(artifact.Kind),
		strings.TrimSpace(artifact.Name),
		strings.TrimSpace(artifact.Filename),
		strings.TrimSpace(artifact.ContentType),
		strings.TrimSpace(artifact.ObjectKey),
		strings.TrimSpace(artifact.SHA256),
		artifact.SizeBytes,
		metadataJSON,
		retentionUntil,
		nullIfEmpty(artifact.RetentionPolicy),
		createdAt,
		strings.TrimSpace(artifact.CreatedBy),
		strings.TrimSpace(artifact.IntegritySHA256),
	)
	if err != nil {
		return fmt.Errorf("insert artifact: %w", err)
	}
	return nil
}

func (s *ArtifactStore) GetArtifact(ctx context.Context, projectID, id string) (domain.Artifact, error) {
	if s == nil || s.db == nil {
		return domain.Artifact{}, fmt.Errorf("artifact store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return domain.Artifact{}, fmt.Errorf("project id is required")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Artifact{}, fmt.Errorf("artifact id is required")
	}
	var artifact domain.Artifact
	var metadataJSON []byte
	var retentionUntil sql.NullTime
	var retentionPolicy sql.NullString
	row := s.db.QueryRowContext(
		ctx,
		`SELECT artifact_id, run_id, project_id, kind, name, filename, content_type, object_key, sha256, size_bytes, metadata, retention_until, retention_policy, created_at, created_by, integrity_sha256
		 FROM experiment_run_artifacts
		 WHERE project_id = $1 AND artifact_id = $2`,
		projectID,
		id,
	)
	if err := row.Scan(&artifact.ID, &artifact.RunID, &artifact.ProjectID, &artifact.Kind, &artifact.Name, &artifact.Filename, &artifact.ContentType, &artifact.ObjectKey, &artifact.SHA256, &artifact.SizeBytes, &metadataJSON, &retentionUntil, &retentionPolicy, &artifact.CreatedAt, &artifact.CreatedBy, &artifact.IntegritySHA256); err != nil {
		return domain.Artifact{}, handleNotFound(err)
	}
	if retentionUntil.Valid {
		retention := retentionUntil.Time.UTC()
		artifact.RetentionUntil = &retention
	}
	if retentionPolicy.Valid {
		artifact.RetentionPolicy = retentionPolicy.String
	}
	meta, err := decodeMetadata(metadataJSON)
	if err != nil {
		return domain.Artifact{}, fmt.Errorf("decode metadata: %w", err)
	}
	artifact.Metadata = meta
	return artifact, nil
}

func (s *ArtifactStore) ListArtifacts(ctx context.Context, filter repo.ArtifactFilter) ([]domain.Artifact, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("artifact store not initialized")
	}
	if strings.TrimSpace(filter.ProjectID) == "" {
		return nil, fmt.Errorf("project id is required")
	}
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if strings.TrimSpace(filter.ProjectID) != "" {
		args = append(args, strings.TrimSpace(filter.ProjectID))
		clauses = append(clauses, fmt.Sprintf("project_id = $%d", len(args)))
	}
	if strings.TrimSpace(filter.RunID) != "" {
		args = append(args, strings.TrimSpace(filter.RunID))
		clauses = append(clauses, fmt.Sprintf("run_id = $%d", len(args)))
	}
	if strings.TrimSpace(filter.Kind) != "" {
		args = append(args, strings.TrimSpace(filter.Kind))
		clauses = append(clauses, fmt.Sprintf("kind = $%d", len(args)))
	}

	query := `SELECT artifact_id, run_id, project_id, kind, name, filename, content_type, object_key, sha256, size_bytes, metadata, retention_until, retention_policy, created_at, created_by, integrity_sha256 FROM experiment_run_artifacts`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}
	defer rows.Close()

	artifacts := make([]domain.Artifact, 0)
	for rows.Next() {
		var artifact domain.Artifact
		var metadataJSON []byte
		var retentionUntil sql.NullTime
		var retentionPolicy sql.NullString
		if err := rows.Scan(&artifact.ID, &artifact.RunID, &artifact.ProjectID, &artifact.Kind, &artifact.Name, &artifact.Filename, &artifact.ContentType, &artifact.ObjectKey, &artifact.SHA256, &artifact.SizeBytes, &metadataJSON, &retentionUntil, &retentionPolicy, &artifact.CreatedAt, &artifact.CreatedBy, &artifact.IntegritySHA256); err != nil {
			return nil, fmt.Errorf("scan artifact: %w", err)
		}
		if retentionUntil.Valid {
			retention := retentionUntil.Time.UTC()
			artifact.RetentionUntil = &retention
		}
		if retentionPolicy.Valid {
			artifact.RetentionPolicy = retentionPolicy.String
		}
		meta, err := decodeMetadata(metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("decode metadata: %w", err)
		}
		artifact.Metadata = meta
		artifacts = append(artifacts, artifact)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list artifacts: %w", err)
	}
	return artifacts, nil
}

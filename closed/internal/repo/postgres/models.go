package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	"github.com/animus-labs/animus-go/closed/internal/repo"
)

type ModelStore struct {
	db DB
}

func NewModelStore(db DB) *ModelStore {
	if db == nil {
		return nil
	}
	return &ModelStore{db: db}
}

func (s *ModelStore) CreateModel(ctx context.Context, model domain.Model) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("model store not initialized")
	}
	if err := model.Validate(); err != nil {
		return err
	}
	if err := requireIntegrity(model.IntegritySHA256); err != nil {
		return err
	}
	metadataJSON, err := encodeMetadata(model.Metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	createdAt := normalizeTime(model.CreatedAt)
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO models (
			model_id,
			project_id,
			name,
			version,
			status,
			artifact_id,
			metadata,
			created_at,
			created_by,
			integrity_sha256
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		strings.TrimSpace(model.ID),
		strings.TrimSpace(model.ProjectID),
		strings.TrimSpace(model.Name),
		strings.TrimSpace(model.Version),
		string(model.Status),
		nullIfEmpty(model.ArtifactID),
		metadataJSON,
		createdAt,
		strings.TrimSpace(model.CreatedBy),
		strings.TrimSpace(model.IntegritySHA256),
	)
	if err != nil {
		return fmt.Errorf("insert model: %w", err)
	}
	return nil
}

func (s *ModelStore) GetModel(ctx context.Context, projectID, id string) (domain.Model, error) {
	if s == nil || s.db == nil {
		return domain.Model{}, fmt.Errorf("model store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return domain.Model{}, fmt.Errorf("project id is required")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.Model{}, fmt.Errorf("model id is required")
	}
	var model domain.Model
	var metadataJSON []byte
	var artifactID sql.NullString
	row := s.db.QueryRowContext(
		ctx,
		`SELECT model_id, project_id, name, version, status, artifact_id, metadata, created_at, created_by, integrity_sha256
		 FROM models
		 WHERE project_id = $1 AND model_id = $2`,
		projectID,
		id,
	)
	if err := row.Scan(&model.ID, &model.ProjectID, &model.Name, &model.Version, &model.Status, &artifactID, &metadataJSON, &model.CreatedAt, &model.CreatedBy, &model.IntegritySHA256); err != nil {
		return domain.Model{}, handleNotFound(err)
	}
	if artifactID.Valid {
		model.ArtifactID = artifactID.String
	}
	meta, err := decodeMetadata(metadataJSON)
	if err != nil {
		return domain.Model{}, fmt.Errorf("decode metadata: %w", err)
	}
	model.Metadata = meta
	return model, nil
}

func (s *ModelStore) ListModels(ctx context.Context, filter repo.ModelFilter) ([]domain.Model, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("model store not initialized")
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
	if strings.TrimSpace(filter.Name) != "" {
		args = append(args, strings.TrimSpace(filter.Name))
		clauses = append(clauses, fmt.Sprintf("name = $%d", len(args)))
	}
	if strings.TrimSpace(string(filter.Status)) != "" {
		args = append(args, string(filter.Status))
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}

	query := `SELECT model_id, project_id, name, version, status, artifact_id, metadata, created_at, created_by, integrity_sha256 FROM models`
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
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()

	models := make([]domain.Model, 0)
	for rows.Next() {
		var model domain.Model
		var metadataJSON []byte
		var artifactID sql.NullString
		if err := rows.Scan(&model.ID, &model.ProjectID, &model.Name, &model.Version, &model.Status, &artifactID, &metadataJSON, &model.CreatedAt, &model.CreatedBy, &model.IntegritySHA256); err != nil {
			return nil, fmt.Errorf("scan model: %w", err)
		}
		if artifactID.Valid {
			model.ArtifactID = artifactID.String
		}
		meta, err := decodeMetadata(metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("decode metadata: %w", err)
		}
		model.Metadata = meta
		models = append(models, model)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return models, nil
}

func (s *ModelStore) UpdateModelStatus(ctx context.Context, projectID, id string, status domain.ModelStatus) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("model store not initialized")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return fmt.Errorf("project id is required")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("model id is required")
	}
	if !status.Valid() {
		return fmt.Errorf("invalid model status")
	}
	res, err := s.db.ExecContext(ctx, `UPDATE models SET status = $1 WHERE project_id = $2 AND model_id = $3`, string(status), projectID, id)
	if err != nil {
		return fmt.Errorf("update model status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update model status: %w", err)
	}
	if rows == 0 {
		return repo.ErrNotFound
	}
	return nil
}

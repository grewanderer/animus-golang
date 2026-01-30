package artifacts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/animus-labs/animus-go/closed/internal/domain"
	store "github.com/animus-labs/animus-go/closed/internal/storage/objectstore"
)

// Store handles artifact persistence into object storage.
type Store struct {
	bucket string
	store  store.Store
	now    func() time.Time
}

// Upload describes an artifact upload request.
type Upload struct {
	ProjectID       string
	RunID           string
	ArtifactID      string
	Kind            string
	Name            string
	Filename        string
	ContentType     string
	CreatedBy       string
	Metadata        domain.Metadata
	RetentionUntil  *time.Time
	RetentionPolicy string
	Body            []byte
}

func NewStore(objectStore store.Store, bucket string) (*Store, error) {
	if objectStore == nil {
		return nil, errors.New("object store is required")
	}
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return nil, errors.New("bucket is required")
	}
	return &Store{bucket: bucket, store: objectStore, now: time.Now}, nil
}

func (s *Store) PutArtifact(ctx context.Context, upload Upload) (domain.Artifact, error) {
	if s == nil || s.store == nil {
		return domain.Artifact{}, errors.New("artifact store not initialized")
	}
	if strings.TrimSpace(upload.ArtifactID) == "" {
		return domain.Artifact{}, errors.New("artifact id is required")
	}
	if strings.TrimSpace(upload.RunID) == "" {
		return domain.Artifact{}, errors.New("run id is required")
	}
	if strings.TrimSpace(upload.ProjectID) == "" {
		return domain.Artifact{}, errors.New("project id is required")
	}
	if strings.TrimSpace(upload.Kind) == "" {
		return domain.Artifact{}, errors.New("artifact kind is required")
	}

	body := upload.Body
	size := int64(len(body))
	sha := sha256.Sum256(body)
	sum := hex.EncodeToString(sha[:])

	objectKey := fmt.Sprintf("runs/%s/%s", upload.RunID, upload.ArtifactID)
	err := s.store.Put(ctx, s.bucket, objectKey, bytes.NewReader(body), size, strings.TrimSpace(upload.ContentType))
	if err != nil {
		return domain.Artifact{}, err
	}

	createdAt := s.now().UTC()
	artifact := domain.Artifact{
		ID:              strings.TrimSpace(upload.ArtifactID),
		ProjectID:       strings.TrimSpace(upload.ProjectID),
		RunID:           strings.TrimSpace(upload.RunID),
		Kind:            strings.TrimSpace(upload.Kind),
		Name:            strings.TrimSpace(upload.Name),
		Filename:        strings.TrimSpace(upload.Filename),
		ContentType:     strings.TrimSpace(upload.ContentType),
		ObjectKey:       objectKey,
		SHA256:          sum,
		SizeBytes:       size,
		Metadata:        upload.Metadata,
		RetentionUntil:  upload.RetentionUntil,
		RetentionPolicy: strings.TrimSpace(upload.RetentionPolicy),
		CreatedAt:       createdAt,
		CreatedBy:       strings.TrimSpace(upload.CreatedBy),
		IntegritySHA256: sum,
	}
	return artifact, nil
}

// GetArtifact loads an artifact payload from object storage.
func (s *Store) GetArtifact(ctx context.Context, artifact domain.Artifact) (io.ReadCloser, error) {
	if s == nil || s.store == nil {
		return nil, errors.New("artifact store not initialized")
	}
	if strings.TrimSpace(artifact.ObjectKey) == "" {
		return nil, errors.New("object key is required")
	}
	reader, _, err := s.store.Get(ctx, s.bucket, artifact.ObjectKey)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

package document

import (
	"context"
	"errors"
	"testing"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/document"
)

type fakeRepo struct {
	createID  int64
	createErr error
}

func (f *fakeRepo) Create(ctx context.Context, doc document.Document) (int64, error) {
	return f.createID, f.createErr
}

func (f *fakeRepo) GetByID(ctx context.Context, id int64) (document.Document, error) {
	return document.Document{}, errors.New("not implemented")
}

func (f *fakeRepo) ListByCluster(ctx context.Context, clusterID int64, limit, offset int) ([]document.Document, error) {
	return nil, errors.New("not implemented")
}

func TestDocumentServiceCreate(t *testing.T) {
	svc := NewService(nil, logger.Nop())
	if _, err := svc.Create(context.Background(), document.Document{}); err == nil {
		t.Fatalf("expected error with nil repo")
	}

	repo := &fakeRepo{createID: 42}
	svc = NewService(repo, logger.Nop())
	id, err := svc.Create(context.Background(), document.Document{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected id=42, got %d", id)
	}
}

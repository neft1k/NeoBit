package document

import (
	"context"
	"fmt"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/document"
)

type DocumentService struct {
	repo Repository
	log  logger.Logger
}

func NewService(repo Repository, log logger.Logger) *DocumentService {
	if log == nil {
		log = logger.Nop()
	}
	return &DocumentService{repo: repo, log: log}
}

func (s *DocumentService) Create(ctx context.Context, doc document.Document) (int64, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("document service: repo is nil")
	}
	return s.repo.Create(ctx, doc)
}

func (s *DocumentService) GetByID(ctx context.Context, id int64) (document.Document, error) {
	if s.repo == nil {
		return document.Document{}, fmt.Errorf("document service: repo is nil")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *DocumentService) ListByCluster(ctx context.Context, clusterID int64, limit, offset int) ([]document.Document, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("document service: repo is nil")
	}
	return s.repo.ListByCluster(ctx, clusterID, limit, offset)
}

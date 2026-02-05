package document

import (
	"NeoBIT/internal/models/document"
	"context"
)

type Service interface {
	Create(ctx context.Context, doc document.Document) (int64, error)
	GetByID(ctx context.Context, id int64) (document.Document, error)
	ListByCluster(ctx context.Context, clusterID int64, limit, offset int) ([]document.Document, error)
}

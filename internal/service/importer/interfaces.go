package importer

import (
	"context"

	"NeoBIT/internal/models/document"
)

type DocumentRepository interface {
	CreateBatch(ctx context.Context, docs []document.Document) (int64, error)
	Count(ctx context.Context) (int64, error)
}

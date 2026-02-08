package document

import (
	"context"
	"fmt"

	"NeoBIT/internal/models/document"
	sq "github.com/Masterminds/squirrel"
	"github.com/pgvector/pgvector-go"
)

func (r *DocumentRepo) CreateBatch(ctx context.Context, docs []document.Document) (int64, error) {
	if r.pool == nil {
		return 0, fmt.Errorf("document repo: pool is nil")
	}
	if len(docs) == 0 {
		return 0, nil
	}

	builder := sq.
		Insert("documents").
		Columns("hn_id", "title", "url", "by", "score", "time", "text", "embedding", "cluster_id").
		PlaceholderFormat(sq.Dollar)

	for _, doc := range docs {
		builder = builder.Values(
			doc.HNID,
			doc.Title,
			doc.URL,
			doc.By,
			doc.Score,
			doc.Time,
			doc.Text,
			pgvector.NewVector(doc.Embedding),
			doc.ClusterID,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build insert batch documents: %w", err)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("insert batch documents: %w", err)
	}

	return tag.RowsAffected(), nil
}

func (r *DocumentRepo) Count(ctx context.Context) (int64, error) {
	if r.pool == nil {
		return 0, fmt.Errorf("document repo: pool is nil")
	}

	query, args, err := sq.
		Select("COUNT(*)").
		From("documents").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build count documents: %w", err)
	}

	var count int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return count, nil
}

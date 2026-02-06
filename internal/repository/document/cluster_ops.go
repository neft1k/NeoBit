package document

import (
	"context"
	"fmt"

	"NeoBIT/internal/models/document"
	sq "github.com/Masterminds/squirrel"
	"github.com/pgvector/pgvector-go"
)

func (r *DocumentRepo) ListUnclustered(ctx context.Context, limit int) ([]document.Document, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("document repo: pool is nil")
	}
	if limit <= 0 {
		limit = 1000
	}

	query, args, err := sq.
		Select(
			"id",
			"COALESCE(hn_id, 0) AS hn_id",
			"COALESCE(title, '') AS title",
			"COALESCE(url, '') AS url",
			"COALESCE(by, '') AS by",
			"COALESCE(score, 0) AS score",
			"COALESCE(time, now()) AS time",
			"COALESCE(text, '') AS text",
			"embedding",
			"cluster_id",
			"created_at",
			"updated_at",
		).
		From("documents").
		Where("cluster_id IS NULL").
		OrderBy("id ASC").
		Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list unclustered: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list unclustered documents: %w", err)
	}
	defer rows.Close()

	var out []document.Document
	for rows.Next() {
		var doc document.Document
		var embedding pgvector.Vector
		if err := rows.Scan(
			&doc.ID,
			&doc.HNID,
			&doc.Title,
			&doc.URL,
			&doc.By,
			&doc.Score,
			&doc.Time,
			&doc.Text,
			&embedding,
			&doc.ClusterID,
			&doc.CreatedAt,
			&doc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan unclustered document: %w", err)
		}
		doc.Embedding = embedding.Slice()
		out = append(out, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unclustered documents: %w", err)
	}
	return out, nil
}

func (r *DocumentRepo) UpdateClusterIDs(ctx context.Context, ids []int64, clusterID int64) error {
	if r.pool == nil {
		return fmt.Errorf("document repo: pool is nil")
	}
	if len(ids) == 0 {
		return nil
	}

	query, args, err := sq.
		Update("documents").
		Set("cluster_id", clusterID).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": ids}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update cluster ids: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update cluster ids: %w", err)
	}
	return nil
}

func (r *DocumentRepo) PctClustered(ctx context.Context) (float64, error) {
	if r.pool == nil {
		return 0, fmt.Errorf("document repo: pool is nil")
	}

	var pct float64
	query, args, err := sq.
		Select("COALESCE(100.0 * COUNT(*) FILTER (WHERE cluster_id IS NOT NULL) / NULLIF(COUNT(*), 0), 0) AS pct").
		From("documents").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("document repo: build pct clustered: %w", err)
	}

	err = r.pool.QueryRow(ctx, query, args...).Scan(&pct)
	if err != nil {
		return 0, fmt.Errorf("document repo: pct clustered: %w", err)
	}
	return pct, nil
}

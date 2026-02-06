package document

import (
	"context"
	"fmt"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/document"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type DocumentRepo struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

func NewDocumentRepo(pool *pgxpool.Pool, log logger.Logger) *DocumentRepo {
	if log == nil {
		log = logger.Nop()
	}
	return &DocumentRepo{pool: pool, log: log}
}

func (r *DocumentRepo) Create(ctx context.Context, doc document.Document) (int64, error) {
	if r.pool == nil {
		return 0, fmt.Errorf("document repo: pool is nil")
	}
	embedding := pgvector.NewVector(doc.Embedding)

	query, args, err := sq.
		Insert("documents").
		Columns("hn_id", "title", "url", "by", "score", "time", "text", "embedding", "cluster_id").
		Values(
			doc.HNID,
			doc.Title,
			doc.URL,
			doc.By,
			doc.Score,
			doc.Time,
			doc.Text,
			embedding,
			doc.ClusterID,
		).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build insert document: %w", err)
	}

	var id int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert document: %w", err)
	}
	return id, nil
}

func (r *DocumentRepo) GetByID(ctx context.Context, id int64) (document.Document, error) {
	if r.pool == nil {
		return document.Document{}, fmt.Errorf("document repo: pool is nil")
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
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return document.Document{}, fmt.Errorf("build get document: %w", err)
	}

	var doc document.Document
	var embedding pgvector.Vector
	row := r.pool.QueryRow(ctx, query, args...)

	if err := row.Scan(
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
		return document.Document{}, fmt.Errorf("get document: %w", err)
	}

	doc.Embedding = embedding.Slice()
	return doc, nil
}

func (r *DocumentRepo) ListByCluster(ctx context.Context, clusterID int64, limit, offset int) ([]document.Document, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("document repo: pool is nil")
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
		Where(sq.Eq{"cluster_id": clusterID}).
		OrderBy("id ASC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list documents: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list cluster documents: %w", err)
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
			return nil, fmt.Errorf("scan cluster document: %w", err)
		}

		doc.Embedding = embedding.Slice()
		out = append(out, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cluster documents: %w", err)
	}
	return out, nil
}

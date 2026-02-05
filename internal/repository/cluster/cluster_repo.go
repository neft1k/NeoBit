package cluster

import (
	"context"
	"fmt"

	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/cluster"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type ClusterRepo struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

func NewClusterRepo(pool *pgxpool.Pool, log logger.Logger) *ClusterRepo {
	if log == nil {
		log = logger.Nop()
	}
	return &ClusterRepo{pool: pool, log: log}
}

func (r *ClusterRepo) Create(ctx context.Context, cluster cluster.Cluster) (int64, error) {
	if r.pool == nil {
		return 0, fmt.Errorf("cluster repo: pool is nil")
	}

	query, args, err := sq.
		Insert("clusters").
		Columns("algorithm", "k", "centroid").
		Values(cluster.Algorithm, cluster.K, pgvector.NewVector(cluster.Centroid)).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build insert cluster: %w", err)
	}

	var id int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return 0, fmt.Errorf("insert cluster: %w", err)
	}
	return id, nil
}

func (r *ClusterRepo) List(ctx context.Context, limit, offset int) ([]cluster.Cluster, error) {
	if r.pool == nil {
		return nil, fmt.Errorf("cluster repo: pool is nil")
	}

	query, args, err := sq.
		Select("c.id", "c.algorithm", "c.k", "c.centroid", "c.created_at", "c.updated_at", "COALESCE(COUNT(d.id), 0)").
		From("clusters c").
		LeftJoin("documents d ON d.cluster_id = c.id").
		GroupBy("c.id").
		OrderBy("c.id").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list clusters: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	defer rows.Close()

	var out []cluster.Cluster
	for rows.Next() {
		var cluster cluster.Cluster
		var centroid pgvector.Vector

		if err := rows.Scan(
			&cluster.ID,
			&cluster.Algorithm,
			&cluster.K,
			&centroid,
			&cluster.CreatedAt,
			&cluster.UpdatedAt,
			&cluster.Size,
		); err != nil {
			return nil, fmt.Errorf("scan cluster: %w", err)
		}

		cluster.Centroid = centroid.Slice()
		out = append(out, cluster)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clusters: %w", err)
	}
	return out, nil
}

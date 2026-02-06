package cluster

import (
	"NeoBIT/internal/models/cluster"
	"NeoBIT/internal/models/document"
	"context"
)

type ClusterRepository interface {
	Create(ctx context.Context, cluster cluster.Cluster) (int64, error)
	List(ctx context.Context, limit, offset int) ([]cluster.Cluster, error)
	SizeStats(ctx context.Context) (min float64, max float64, avg float64, err error)
}

type DocumentRepository interface {
	ListUnclustered(ctx context.Context, limit int) ([]document.Document, error)
	UpdateClusterIDs(ctx context.Context, ids []int64, clusterID int64) error
	PctClustered(ctx context.Context) (float64, error)
}

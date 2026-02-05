package cluster

import (
	"NeoBIT/internal/models/cluster"
	"context"
)

type Service interface {
	List(ctx context.Context, limit, offset int) ([]cluster.Cluster, error)
}

package cluster

import (
	"context"

	"NeoBIT/internal/models/cluster"
)

type Service interface {
	List(ctx context.Context, limit, offset int) ([]cluster.Cluster, error)
}

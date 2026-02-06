package cluster

import (
	"context"
	"testing"

	"NeoBIT/internal/config"
	"NeoBIT/internal/logger"
	"NeoBIT/internal/models/cluster"
	"NeoBIT/internal/models/document"
)

type fakeClusterRepo struct{}

func (f *fakeClusterRepo) Create(ctx context.Context, c cluster.Cluster) (int64, error) {
	return 1, nil
}

func (f *fakeClusterRepo) List(ctx context.Context, limit, offset int) ([]cluster.Cluster, error) {
	return []cluster.Cluster{{ID: 1}}, nil
}

func (f *fakeClusterRepo) SizeStats(ctx context.Context) (float64, float64, float64, error) {
	return 0, 0, 0, nil
}

type fakeDocRepo struct{}

func (f *fakeDocRepo) ListUnclustered(ctx context.Context, limit int) ([]document.Document, error) {
	return nil, nil
}

func (f *fakeDocRepo) UpdateClusterIDs(ctx context.Context, ids []int64, clusterID int64) error {
	return nil
}

func (f *fakeDocRepo) PctClustered(ctx context.Context) (float64, error) {
	return 0, nil
}

func TestClusterServiceListNilRepo(t *testing.T) {
	svc := NewService(nil, nil, config.DefaultClusterConfig(), logger.Nop())
	if _, err := svc.List(context.Background(), 10, 0); err == nil {
		t.Fatalf("expected error with nil repo")
	}
}

func TestClusterServiceList(t *testing.T) {
	svc := NewService(&fakeClusterRepo{}, &fakeDocRepo{}, config.DefaultClusterConfig(), logger.Nop())
	res, err := svc.List(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result")
	}
}

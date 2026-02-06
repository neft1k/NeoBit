package cluster

import (
	"context"
	"fmt"

	"NeoBIT/internal/config"
	"NeoBIT/internal/logger"
	"NeoBIT/internal/metrics"
	"NeoBIT/internal/models/cluster"
)

type ClusterService struct {
	clusterRepo ClusterRepository
	docRepo     DocumentRepository
	cfg         config.ClusterConfig
	log         logger.Logger
}

func NewService(clusterRepo ClusterRepository, docRepo DocumentRepository, cfg config.ClusterConfig, log logger.Logger) *ClusterService {
	if log == nil {
		log = logger.Nop()
	}
	return &ClusterService{clusterRepo: clusterRepo, docRepo: docRepo, cfg: cfg, log: log}
}

func (s *ClusterService) List(ctx context.Context, limit, offset int) ([]cluster.Cluster, error) {
	if s.clusterRepo == nil {
		return nil, fmt.Errorf("cluster service: cluster repo is nil")
	}
	return s.clusterRepo.List(ctx, limit, offset)
}

func (s *ClusterService) processBatch(ctx context.Context) {
	batchSize := s.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	docs, err := s.docRepo.ListUnclustered(ctx, batchSize)
	if err != nil {
		s.log.Error(ctx, "cluster worker: failed to list unclustered", logger.FieldAny("error", err))
		return
	}
	if len(docs) == 0 {
		s.log.Info(ctx, "cluster worker: no unclustered documents")
		return
	}

	points := make([][]float32, 0, len(docs))
	ids := make([]int64, 0, len(docs))
	for _, doc := range docs {
		if len(doc.Embedding) == 0 {
			continue
		}
		points = append(points, doc.Embedding)
		ids = append(ids, doc.ID)
	}
	if len(points) == 0 {
		s.log.Warn(ctx, "cluster worker: no embeddings in batch")
		return
	}

	k := s.cfg.K
	if k <= 0 {
		k = 10
	}
	if k > len(points) {
		k = len(points)
	}

	s.log.Info(ctx, "cluster worker: processing batch", logger.FieldAny("size", len(points)), logger.FieldAny("k", k))
	assignments, centroids := assignOnePass(points, k)

	clusterIDs := make([]int64, len(centroids))
	for i, centroid := range centroids {
		id, err := s.clusterRepo.Create(ctx, cluster.Cluster{
			Algorithm: "simple",
			K:         k,
			Centroid:  centroid,
		})
		if err != nil {
			s.log.Error(ctx, "cluster worker: create cluster failed", logger.FieldAny("error", err))
			return
		}
		clusterIDs[i] = id
	}

	buckets := make(map[int64][]int64, len(clusterIDs))
	for i, clusterIdx := range assignments {
		clusterID := clusterIDs[clusterIdx]
		buckets[clusterID] = append(buckets[clusterID], ids[i])
	}

	for clusterID, docIDs := range buckets {
		if err := s.docRepo.UpdateClusterIDs(ctx, docIDs, clusterID); err != nil {
			s.log.Error(ctx, "cluster worker: update cluster ids failed", logger.FieldAny("error", err))
			continue
		}
	}
	s.log.Info(ctx, "cluster worker: updated docs", logger.FieldAny("docs", len(points)), logger.FieldAny("clusters", len(clusterIDs)))

	minSize, maxSize, avgSize, err := s.clusterRepo.SizeStats(ctx)
	if err != nil {
		s.log.Error(ctx, "cluster worker: size stats failed", logger.FieldAny("error", err))
	} else {
		metrics.SetClusterSizeStats(minSize, maxSize, avgSize)
	}

	pct, err := s.docRepo.PctClustered(ctx)
	if err != nil {
		s.log.Error(ctx, "cluster worker: pct clustered failed", logger.FieldAny("error", err))
	} else {
		metrics.SetPctClustered(pct)
	}
}

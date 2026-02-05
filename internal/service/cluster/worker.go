package cluster

import (
	"context"
	"time"
)

func StartWorker(ctx context.Context, svc *ClusterService) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if svc == nil || svc.clusterRepo == nil || svc.docRepo == nil {
			return
		}
		interval := svc.cfg.Interval
		if interval <= 0 {
			interval = 5 * time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				svc.processBatch(ctx)
			}
		}
	}()
	return done
}

package importer

import (
	"context"
	"time"

	"NeoBIT/internal/logger"
)

func StartWorker(ctx context.Context, svc *ImportService) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		if svc == nil {
			return
		}

		runCtx, runCancel := context.WithCancel(context.WithoutCancel(ctx))
		defer runCancel()

		runDone := make(chan error, 1)
		go func() {
			runDone <- svc.Run(runCtx)
		}()

		select {
		case err := <-runDone:
			if err != nil {
				svc.log.Error(ctx, "dataset import failed", logger.FieldAny("error", err))
			}
		case <-ctx.Done():
			timeout := svc.cfg.ShutdownTimeout
			if timeout <= 0 {
				timeout = 30 * time.Second
			}

			svc.log.Info(ctx, "import worker: shutdown requested, waiting current work", logger.FieldAny("timeout", timeout))

			timer := time.NewTimer(timeout)
			defer timer.Stop()

			select {
			case err := <-runDone:
				if err != nil {
					svc.log.Error(ctx, "dataset import stopped with error", logger.FieldAny("error", err))
				}
			case <-timer.C:
				svc.log.Warn(ctx, "import worker: graceful wait timeout, canceling run context")
				runCancel()
				select {
				case err := <-runDone:
					if err != nil {
						svc.log.Warn(ctx, "dataset import canceled after timeout", logger.FieldAny("error", err))
					}
				case <-time.After(2 * time.Second):
					svc.log.Warn(ctx, "import worker: run did not stop after cancel")
				}
			}
		}
	}()
	return done
}

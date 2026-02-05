package server

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"NeoBIT/internal/config"
	"NeoBIT/internal/db"
	"NeoBIT/internal/logger"
	clusterrepo "NeoBIT/internal/repository/cluster"
	documentrepo "NeoBIT/internal/repository/document"
	clusterservice "NeoBIT/internal/service/cluster"
	documentservice "NeoBIT/internal/service/document"
	clusterhandler "NeoBIT/internal/transport/http/handler/cluster"
	documenthandler "NeoBIT/internal/transport/http/handler/document"

	"github.com/go-chi/chi/v5"
)

func Start(cfg config.Config, log logger.Logger) error {
	if log == nil {
		log = logger.Nop()
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.InitPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	docRepo := documentrepo.NewDocumentRepo(pool, log)
	docSvc := documentservice.NewService(docRepo, log)
	docHandler := documenthandler.NewHandler(docSvc, log)

	clusterRepo := clusterrepo.NewClusterRepo(pool, log)
	clusterSvc := clusterservice.NewService(clusterRepo, docRepo, config.DefaultClusterConfig(), log)
	clusterHandler := clusterhandler.NewHandler(clusterSvc, log)

	workerDone := clusterservice.StartWorker(ctx, clusterSvc)

	r := chi.NewRouter()

	r.Route("/documents", func(r chi.Router) {
		r.Post("/", docHandler.Create)
		r.Get("/{id}", docHandler.GetByID)
	})

	r.Route("/clusters", func(r chi.Router) {
		r.Get("/", clusterHandler.List)
		r.Get("/{id}/documents", docHandler.ListByCluster)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	log.Info(ctx, "starting server", logger.FieldAny("port", cfg.Port))

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error(ctx, "server error", logger.FieldAny("error", err))
		}
	}()

	<-ctx.Done()

	log.Info(ctx, "shutting down server")

	workersStopped := make(chan struct{})
	go func() {
		if workerDone != nil {
			<-workerDone
		}
		close(workersStopped)
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdownErrCh := make(chan error, 1)
	go func() {
		shutdownErrCh <- srv.Shutdown(shutdownCtx)
	}()

	select {
	case <-workersStopped:
		log.Info(ctx, "background workers stopped")
	case <-shutdownCtx.Done():
		log.Warn(ctx, "background workers shutdown timeout")
	}

	err = <-shutdownErrCh
	if err != nil {
		log.Error(ctx, "server shutdown failed", logger.FieldAny("error", err))
	} else {
		log.Info(ctx, "server shutdown successfully")
	}
	return err
}

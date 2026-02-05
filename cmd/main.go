package main

import (
	"context"
	"log"

	"NeoBIT/internal/config"
	"NeoBIT/internal/logger"
	"NeoBIT/internal/server"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logCfg := config.GetLoggerConfig()
	appLogger, err := logger.New(logCfg)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer func() {
		if err := appLogger.Sync(); err != nil {
			appLogger.Warn(context.Background(), "logger sync failed", logger.FieldAny("error", err))
		}
	}()

	if err := server.Start(cfg, appLogger); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}
}

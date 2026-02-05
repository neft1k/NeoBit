package cluster

import "NeoBIT/internal/logger"

type Handler struct {
	svc Service
	log logger.Logger
}

func NewHandler(svc Service, log logger.Logger) *Handler {
	if log == nil {
		log = logger.Nop()
	}
	return &Handler{svc: svc, log: log}
}

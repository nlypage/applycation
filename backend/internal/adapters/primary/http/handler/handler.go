package handler

import primaryports "github.com/nlypage/applycation/backend/internal/ports/primary"

type Handler struct {
	healthService primaryports.HealthService
}

func New(healthService primaryports.HealthService) *Handler {
	return &Handler{healthService: healthService}
}

package handler

import (
	"context"

	"github.com/nlypage/applycation/backend/internal/adapters/primary/http/openapi"
)

func (h *Handler) GetHealth(ctx context.Context, _ openapi.GetHealthRequestObject) (openapi.GetHealthResponseObject, error) {
	health := h.healthService.Health(ctx)
	return openapi.GetHealth200JSONResponse{Status: health.Status}, nil
}

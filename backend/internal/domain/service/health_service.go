package service

import (
	"context"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
)

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Health(_ context.Context) entity.Health {
	return entity.Health{Status: "ok"}
}

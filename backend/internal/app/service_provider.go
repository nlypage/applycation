package app

import (
	"github.com/nlypage/applycation/backend/internal/domain/service"
	primaryports "github.com/nlypage/applycation/backend/internal/ports/primary"
)

// ServiceProvider wires domain services to primary ports.
type ServiceProvider struct {
	HealthService primaryports.HealthService
}

func NewServiceProvider() *ServiceProvider {
	return &ServiceProvider{
		HealthService: service.NewHealthService(),
	}
}

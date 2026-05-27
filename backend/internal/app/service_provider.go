package app

import (
	"github.com/nlypage/applycation/backend/internal/domain/service"
	primaryports "github.com/nlypage/applycation/backend/internal/ports/primary"
	"github.com/nlypage/applycation/backend/pkg/closer"
)

// ServiceProvider wires domain services to primary ports.
type ServiceProvider struct {
	HealthService primaryports.HealthService

	closer *closer.Closer
}

func NewServiceProvider(closeManager *closer.Closer) *ServiceProvider {
	if closeManager == nil {
		closeManager = closer.New()
	}

	return &ServiceProvider{
		HealthService: service.NewHealthService(),
		closer:        closeManager,
	}
}

func (p *ServiceProvider) AddCloser(f ...func() error) {
	p.closer.Add(f...)
}

func (p *ServiceProvider) CloseAll() {
	p.closer.CloseAll()
}

func (p *ServiceProvider) Wait() {
	p.closer.Wait()
}

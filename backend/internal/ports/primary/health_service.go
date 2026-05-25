package primary

import (
	"context"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
)

// HealthService is an inbound use-case port.
type HealthService interface {
	Health(ctx context.Context) entity.Health
}

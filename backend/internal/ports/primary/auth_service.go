package primary

import (
	"context"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
)

// SetupInput contains data for first local owner setup.
type SetupInput struct {
	Password string
}

// AuthService is an inbound local authentication use-case port.
type AuthService interface {
	Setup(ctx context.Context, input SetupInput) (entity.Owner, error)
}

package primary

import (
	"context"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
)

// SetupInput contains data for first local owner setup.
type SetupInput struct {
	Password string
}

// LoginInput contains data for local owner login.
type LoginInput struct {
	Password  string
	UserAgent string
	IPAddress string
}

// LoginResult contains authenticated owner session data.
type LoginResult struct {
	Owner        entity.Owner
	Session      entity.OwnerSession
	SessionToken string
}

// AuthService is an inbound local authentication use-case port.
type AuthService interface {
	Setup(ctx context.Context, input SetupInput) (entity.Owner, error)
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
}

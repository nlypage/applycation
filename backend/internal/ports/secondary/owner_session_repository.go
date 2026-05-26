package secondary

import (
	"context"
	"time"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
)

type CreateOwnerSessionParams struct {
	OwnerID          string
	SessionTokenHash string
	UserAgent        *string
	IPAddress        *string
	ExpiresAt        time.Time
}

type OwnerSessionRepository interface {
	Create(ctx context.Context, params CreateOwnerSessionParams) (entity.OwnerSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (entity.OwnerSession, error)
	Touch(ctx context.Context, tokenHash string) (entity.OwnerSession, error)
	Revoke(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
}

package secondary

import (
	"context"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
)

type OwnerRepository interface {
	Create(ctx context.Context, passwordHash string) (entity.Owner, error)
	GetSingle(ctx context.Context) (entity.Owner, error)
	GetByID(ctx context.Context, id string) (entity.Owner, error)
	UpdatePassword(ctx context.Context, id string, passwordHash string) (entity.Owner, error)
}

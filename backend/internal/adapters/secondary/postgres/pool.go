package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nlypage/applycation/backend/internal/adapters/config"
)

func NewPool(ctx context.Context, cfg config.Postgres) (*pgxpool.Pool, error) {
	if cfg.DatabaseURL == "" {
		return nil, configErr("DATABASE_URL is empty")
	}

	return pgxpool.New(ctx, cfg.DatabaseURL)
}

type configErr string

func (e configErr) Error() string { return string(e) }

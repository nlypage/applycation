package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/nlypage/applycation/backend/internal/adapters/secondary/postgres/sqlc"
	"github.com/nlypage/applycation/backend/internal/domain/entity"
	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

type OwnerRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

var _ secondaryports.OwnerRepository = (*OwnerRepository)(nil)

func NewOwnerRepository(pool *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{pool: pool, q: db.New()}
}

func (r *OwnerRepository) Create(ctx context.Context, passwordHash string) (entity.Owner, error) {
	owner, err := r.q.CreateOwner(ctx, dbtx(ctx, r.pool), passwordHash)
	if err != nil {
		return entity.Owner{}, fmt.Errorf("create owner: %w", err)
	}

	return mapOwner(owner), nil
}

func (r *OwnerRepository) GetSingle(ctx context.Context) (entity.Owner, error) {
	owner, err := r.q.GetSingleOwner(ctx, dbtx(ctx, r.pool))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Owner{}, secondaryports.ErrNotFound
		}
		return entity.Owner{}, fmt.Errorf("get single owner: %w", err)
	}

	return mapOwner(owner), nil
}

func (r *OwnerRepository) GetByID(ctx context.Context, id string) (entity.Owner, error) {
	pgID, err := toPgUUID(id)
	if err != nil {
		return entity.Owner{}, err
	}

	owner, err := r.q.GetOwnerByID(ctx, dbtx(ctx, r.pool), pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Owner{}, secondaryports.ErrNotFound
		}
		return entity.Owner{}, fmt.Errorf("get owner by id: %w", err)
	}

	return mapOwner(owner), nil
}

func (r *OwnerRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) (entity.Owner, error) {
	pgID, err := toPgUUID(id)
	if err != nil {
		return entity.Owner{}, err
	}

	owner, err := r.q.UpdateOwnerPassword(ctx, dbtx(ctx, r.pool), db.UpdateOwnerPasswordParams{
		ID:           pgID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Owner{}, secondaryports.ErrNotFound
		}
		return entity.Owner{}, fmt.Errorf("update owner password: %w", err)
	}

	return mapOwner(owner), nil
}

func mapOwner(owner db.Owner) entity.Owner {
	return entity.Owner{
		ID:                owner.ID.String(),
		PasswordHash:      owner.PasswordHash,
		PasswordChangedAt: timestamptzToTime(owner.PasswordChangedAt),
		CreatedAt:         timestamptzToTime(owner.CreatedAt),
		UpdatedAt:         timestamptzToTime(owner.UpdatedAt),
	}
}

func toPgUUID(id string) (pgtype.UUID, error) {
	var pgID pgtype.UUID
	if err := pgID.Scan(id); err != nil {
		return pgtype.UUID{}, fmt.Errorf("parse uuid: %w", err)
	}
	return pgID, nil
}

func timestamptzToTime(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}

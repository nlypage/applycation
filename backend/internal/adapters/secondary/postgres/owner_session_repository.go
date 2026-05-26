package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/nlypage/applycation/backend/internal/adapters/secondary/postgres/sqlc"
	"github.com/nlypage/applycation/backend/internal/domain/entity"
	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

type OwnerSessionRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

var _ secondaryports.OwnerSessionRepository = (*OwnerSessionRepository)(nil)

func NewOwnerSessionRepository(pool *pgxpool.Pool) *OwnerSessionRepository {
	return &OwnerSessionRepository{pool: pool, q: db.New()}
}

func (r *OwnerSessionRepository) Create(ctx context.Context, params secondaryports.CreateOwnerSessionParams) (entity.OwnerSession, error) {
	ownerID, err := toPgUUID(params.OwnerID)
	if err != nil {
		return entity.OwnerSession{}, err
	}

	ipAddress, err := netipAddrFromStringPtr(params.IPAddress)
	if err != nil {
		return entity.OwnerSession{}, err
	}

	session, err := r.q.CreateOwnerSession(ctx, dbtx(ctx, r.pool), db.CreateOwnerSessionParams{
		OwnerID:          ownerID,
		SessionTokenHash: params.SessionTokenHash,
		UserAgent:        params.UserAgent,
		IpAddress:        ipAddress,
		ExpiresAt:        pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
	})
	if err != nil {
		return entity.OwnerSession{}, fmt.Errorf("create owner session: %w", err)
	}

	return mapOwnerSession(session), nil
}

func (r *OwnerSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (entity.OwnerSession, error) {
	session, err := r.q.GetOwnerSessionByTokenHash(ctx, dbtx(ctx, r.pool), tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.OwnerSession{}, secondaryports.ErrNotFound
		}
		return entity.OwnerSession{}, fmt.Errorf("get owner session by token hash: %w", err)
	}

	return mapOwnerSession(session), nil
}

func (r *OwnerSessionRepository) Touch(ctx context.Context, tokenHash string) (entity.OwnerSession, error) {
	session, err := r.q.TouchOwnerSession(ctx, dbtx(ctx, r.pool), tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.OwnerSession{}, secondaryports.ErrNotFound
		}
		return entity.OwnerSession{}, fmt.Errorf("touch owner session: %w", err)
	}

	return mapOwnerSession(session), nil
}

func (r *OwnerSessionRepository) Revoke(ctx context.Context, tokenHash string) error {
	if err := r.q.RevokeOwnerSession(ctx, dbtx(ctx, r.pool), tokenHash); err != nil {
		return fmt.Errorf("revoke owner session: %w", err)
	}
	return nil
}

func (r *OwnerSessionRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	rows, err := r.q.DeleteExpiredOwnerSessions(ctx, dbtx(ctx, r.pool), pgtype.Timestamptz{Time: now, Valid: true})
	if err != nil {
		return 0, fmt.Errorf("delete expired owner sessions: %w", err)
	}
	return rows, nil
}

func mapOwnerSession(session db.OwnerSession) entity.OwnerSession {
	return entity.OwnerSession{
		ID:               session.ID.String(),
		OwnerID:          session.OwnerID.String(),
		SessionTokenHash: session.SessionTokenHash,
		UserAgent:        session.UserAgent,
		IPAddress:        netipAddrToStringPtr(session.IpAddress),
		ExpiresAt:        timestamptzToTime(session.ExpiresAt),
		RevokedAt:        timestamptzToPtr(session.RevokedAt),
		LastSeenAt:       timestamptzToTime(session.LastSeenAt),
		CreatedAt:        timestamptzToTime(session.CreatedAt),
		UpdatedAt:        timestamptzToTime(session.UpdatedAt),
	}
}

func netipAddrFromStringPtr(v *string) (*netip.Addr, error) {
	if v == nil || *v == "" {
		return nil, nil
	}
	addr, err := netip.ParseAddr(*v)
	if err != nil {
		return nil, fmt.Errorf("parse ip address: %w", err)
	}
	return &addr, nil
}

func netipAddrToStringPtr(v *netip.Addr) *string {
	if v == nil {
		return nil
	}
	return new(v.String())
}

func timestamptzToPtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	return new(v.Time)
}

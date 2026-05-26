package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/nlypage/applycation/backend/internal/adapters/secondary/postgres"
	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
	"github.com/nlypage/applycation/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOwnerSessionRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)
	ownerRepo := postgres.NewOwnerRepository(db.Pool)
	sessionRepo := postgres.NewOwnerSessionRepository(db.Pool)

	// Create test owner first
	ctx := context.Background()
	owner, err := ownerRepo.Create(ctx, "test_password_hash")
	require.NoError(t, err)

	t.Run("Create and GetByTokenHash", func(t *testing.T) {
		tokenHash := "test_token_hash_123"
		userAgent := "Mozilla/5.0 Test Browser"
		ipAddress := "192.168.1.100"
		expiresAt := time.Now().Add(24 * time.Hour)

		params := secondaryports.CreateOwnerSessionParams{
			OwnerID:          owner.ID,
			SessionTokenHash: tokenHash,
			UserAgent:        &userAgent,
			IPAddress:        &ipAddress,
			ExpiresAt:        expiresAt,
		}

		// Create session
		session, err := sessionRepo.Create(ctx, params)
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, owner.ID, session.OwnerID)
		assert.Equal(t, tokenHash, session.SessionTokenHash)
		assert.Equal(t, userAgent, *session.UserAgent)
		assert.Equal(t, ipAddress, *session.IPAddress)
		assert.WithinDuration(t, expiresAt, session.ExpiresAt, time.Second)
		assert.Nil(t, session.RevokedAt)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.UpdatedAt.IsZero())

		// Get by token hash
		retrieved, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.OwnerID, retrieved.OwnerID)
		assert.Equal(t, session.SessionTokenHash, retrieved.SessionTokenHash)
	})

	t.Run("Create with minimal params", func(t *testing.T) {
		tokenHash := "minimal_token_hash"
		expiresAt := time.Now().Add(12 * time.Hour)

		params := secondaryports.CreateOwnerSessionParams{
			OwnerID:          owner.ID,
			SessionTokenHash: tokenHash,
			UserAgent:        nil,
			IPAddress:        nil,
			ExpiresAt:        expiresAt,
		}

		session, err := sessionRepo.Create(ctx, params)
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, owner.ID, session.OwnerID)
		assert.Equal(t, tokenHash, session.SessionTokenHash)
		assert.Nil(t, session.UserAgent)
		assert.Nil(t, session.IPAddress)
	})

	t.Run("Touch session", func(t *testing.T) {
		tokenHash := "touch_test_token"
		expiresAt := time.Now().Add(24 * time.Hour)

		params := secondaryports.CreateOwnerSessionParams{
			OwnerID:          owner.ID,
			SessionTokenHash: tokenHash,
			ExpiresAt:        expiresAt,
		}

		// Create session
		session, err := sessionRepo.Create(ctx, params)
		require.NoError(t, err)
		originalLastSeen := session.LastSeenAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Touch session
		touched, err := sessionRepo.Touch(ctx, tokenHash)
		require.NoError(t, err)
		assert.Equal(t, session.ID, touched.ID)
		assert.True(t, touched.LastSeenAt.After(originalLastSeen))
	})

	t.Run("Revoke session", func(t *testing.T) {
		tokenHash := "revoke_test_token"
		expiresAt := time.Now().Add(24 * time.Hour)

		params := secondaryports.CreateOwnerSessionParams{
			OwnerID:          owner.ID,
			SessionTokenHash: tokenHash,
			ExpiresAt:        expiresAt,
		}

		// Create session
		session, err := sessionRepo.Create(ctx, params)
		require.NoError(t, err)
		assert.Nil(t, session.RevokedAt)

		// Revoke session
		err = sessionRepo.Revoke(ctx, tokenHash)
		require.NoError(t, err)

		// Verify session is revoked
		retrieved, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.RevokedAt)
		assert.False(t, retrieved.RevokedAt.IsZero())
	})

	t.Run("DeleteExpired", func(t *testing.T) {
		// Create expired session
		expiredTokenHash := "expired_token"
		expiredTime := time.Now().Add(-1 * time.Hour) // Already expired

		expiredParams := secondaryports.CreateOwnerSessionParams{
			OwnerID:          owner.ID,
			SessionTokenHash: expiredTokenHash,
			ExpiresAt:        expiredTime,
		}

		_, err := sessionRepo.Create(ctx, expiredParams)
		require.NoError(t, err)

		// Create valid session
		validTokenHash := "valid_token"
		validTime := time.Now().Add(1 * time.Hour) // Not expired

		validParams := secondaryports.CreateOwnerSessionParams{
			OwnerID:          owner.ID,
			SessionTokenHash: validTokenHash,
			ExpiresAt:        validTime,
		}

		_, err = sessionRepo.Create(ctx, validParams)
		require.NoError(t, err)

		// Delete expired sessions
		deletedCount, err := sessionRepo.DeleteExpired(ctx, time.Now())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deletedCount, int64(1))

		// Expired session should be gone
		_, err = sessionRepo.GetByTokenHash(ctx, expiredTokenHash)
		assert.Error(t, err) // Should be ErrNotFound

		// Valid session should still exist
		_, err = sessionRepo.GetByTokenHash(ctx, validTokenHash)
		require.NoError(t, err)
	})

	t.Run("GetByTokenHash_NotFound", func(t *testing.T) {
		_, err := sessionRepo.GetByTokenHash(ctx, "non_existent_token")
		assert.Error(t, err)
		// Should return ErrNotFound from secondary ports
	})

	t.Run("Touch_NotFound", func(t *testing.T) {
		_, err := sessionRepo.Touch(ctx, "non_existent_token")
		assert.Error(t, err)
		// Should return ErrNotFound from secondary ports
	})
}

func TestOwnerSessionRepository_WithTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)
	ownerRepo := postgres.NewOwnerRepository(db.Pool)
	sessionRepo := postgres.NewOwnerSessionRepository(db.Pool)

	// Create test owner
	ctx := context.Background()
	owner, err := ownerRepo.Create(ctx, "tx_test_password")
	require.NoError(t, err)

	t.Run("Transaction rollback", func(t *testing.T) {
		tokenHash := "tx_rollback_token"
		expiresAt := time.Now().Add(24 * time.Hour)

		err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
			params := secondaryports.CreateOwnerSessionParams{
				OwnerID:          owner.ID,
				SessionTokenHash: tokenHash,
				ExpiresAt:        expiresAt,
			}

			// Create session within transaction
			session, err := sessionRepo.Create(txCtx, params)
			require.NoError(t, err)
			assert.NotEmpty(t, session.ID)

			// Verify session exists within transaction
			retrieved, err := sessionRepo.GetByTokenHash(txCtx, tokenHash)
			require.NoError(t, err)
			assert.Equal(t, session.ID, retrieved.ID)

			// Return error to trigger rollback
			return assert.AnError
		})

		// Transaction should have failed
		require.Error(t, err)

		// Session should not exist after rollback
		_, err = sessionRepo.GetByTokenHash(ctx, tokenHash)
		assert.Error(t, err) // Should be ErrNotFound
	})

	t.Run("Transaction commit", func(t *testing.T) {
		tokenHash := "tx_commit_token"
		expiresAt := time.Now().Add(24 * time.Hour)

		err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
			params := secondaryports.CreateOwnerSessionParams{
				OwnerID:          owner.ID,
				SessionTokenHash: tokenHash,
				ExpiresAt:        expiresAt,
			}

			_, err := sessionRepo.Create(txCtx, params)
			require.NoError(t, err)
			return nil
		})

		require.NoError(t, err)

		// Session should exist after commit
		retrieved, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
		require.NoError(t, err)
		assert.Equal(t, tokenHash, retrieved.SessionTokenHash)
	})
}

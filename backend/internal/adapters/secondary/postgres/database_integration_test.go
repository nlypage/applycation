package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nlypage/applycation/backend/internal/adapters/secondary/postgres"
	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
	"github.com/nlypage/applycation/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)

	t.Run("Database connection and migrations", func(t *testing.T) {
		ctx := context.Background()

		// Test basic connection
		err := db.Pool.Ping(ctx)
		require.NoError(t, err)

		// Verify tables exist by querying them
		var count int
		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM owners").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count) // Should be empty initially

		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM owner_sessions").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		err = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM credentials").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Cross-repository operations", func(t *testing.T) {
		ctx := context.Background()
		ownerRepo := postgres.NewOwnerRepository(db.Pool)
		sessionRepo := postgres.NewOwnerSessionRepository(db.Pool)

		// Create owner
		owner, err := ownerRepo.Create(ctx, "integration_test_hash")
		require.NoError(t, err)

		// Create multiple sessions for the owner
		sessions := make([]string, 3)
		for i := 0; i < 3; i++ {
			tokenHash := fmt.Sprintf("token_hash_%d", i)
			sessions[i] = tokenHash

			params := secondaryports.CreateOwnerSessionParams{
				OwnerID:          owner.ID,
				SessionTokenHash: tokenHash,
				ExpiresAt:        time.Now().Add(24 * time.Hour),
			}

			session, err := sessionRepo.Create(ctx, params)
			require.NoError(t, err)
			assert.Equal(t, owner.ID, session.OwnerID)
		}

		// Verify all sessions exist
		for _, tokenHash := range sessions {
			session, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
			require.NoError(t, err)
			assert.Equal(t, owner.ID, session.OwnerID)
		}

		// Update owner password
		newPasswordHash := "updated_password_hash"
		updatedOwner, err := ownerRepo.UpdatePassword(ctx, owner.ID, newPasswordHash)
		require.NoError(t, err)
		assert.Equal(t, newPasswordHash, updatedOwner.PasswordHash)

		// Sessions should still be valid after password update
		for _, tokenHash := range sessions {
			session, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
			require.NoError(t, err)
			assert.Equal(t, owner.ID, session.OwnerID)
		}
	})

	t.Run("Complex transaction scenario", func(t *testing.T) {
		ctx := context.Background()
		ownerRepo := postgres.NewOwnerRepository(db.Pool)
		sessionRepo := postgres.NewOwnerSessionRepository(db.Pool)

		// Test complex transaction with multiple operations
		var ownerID string
		var sessionTokens []string

		err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
			// Create owner
			owner, err := ownerRepo.Create(txCtx, "complex_tx_hash")
			if err != nil {
				return err
			}
			ownerID = owner.ID

			// Create multiple sessions
			for i := 0; i < 2; i++ {
				tokenHash := fmt.Sprintf("complex_token_%d", i)
				sessionTokens = append(sessionTokens, tokenHash)

				params := secondaryports.CreateOwnerSessionParams{
					OwnerID:          owner.ID,
					SessionTokenHash: tokenHash,
					ExpiresAt:        time.Now().Add(24 * time.Hour),
				}

				_, err := sessionRepo.Create(txCtx, params)
				if err != nil {
					return err
				}
			}

			// Update owner password within same transaction
			_, err = ownerRepo.UpdatePassword(txCtx, owner.ID, "updated_in_tx_hash")
			return err
		})

		require.NoError(t, err)

		// Verify all operations committed successfully
		retrievedOwner, err := ownerRepo.GetByID(ctx, ownerID)
		require.NoError(t, err)
		assert.Equal(t, "updated_in_tx_hash", retrievedOwner.PasswordHash)

		for _, tokenHash := range sessionTokens {
			session, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
			require.NoError(t, err)
			assert.Equal(t, ownerID, session.OwnerID)
		}
	})

	t.Run("Concurrent operations", func(t *testing.T) {
		ctx := context.Background()
		ownerRepo := postgres.NewOwnerRepository(db.Pool)

		// Create owner for concurrent test
		owner, err := ownerRepo.Create(ctx, "concurrent_test_hash")
		require.NoError(t, err)

		// Test concurrent password updates
		done := make(chan error, 2)

		go func() {
			_, err := ownerRepo.UpdatePassword(ctx, owner.ID, "concurrent_hash_1")
			done <- err
		}()

		go func() {
			_, err := ownerRepo.UpdatePassword(ctx, owner.ID, "concurrent_hash_2")
			done <- err
		}()

		// Both operations should complete without deadlock
		err1 := <-done
		err2 := <-done

		// At least one should succeed (both might succeed due to serialization)
		assert.True(t, err1 == nil || err2 == nil, "at least one concurrent operation should succeed")

		// Final state should be consistent
		finalOwner, err := ownerRepo.GetByID(ctx, owner.ID)
		require.NoError(t, err)
		assert.True(t,
			finalOwner.PasswordHash == "concurrent_hash_1" ||
				finalOwner.PasswordHash == "concurrent_hash_2",
			"final password should be one of the concurrent updates")
	})
}

func TestTxManager_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)

	t.Run("Nested transaction unsupported", func(t *testing.T) {
		ctx := context.Background()
		ownerRepo := postgres.NewOwnerRepository(db.Pool)

		// Current TxManager starts a new transaction each call,
		// so nested RunInTx does not share outer tx context.
		err := db.TxManager.RunInTx(ctx, func(outerCtx context.Context) error {
			owner, err := ownerRepo.Create(outerCtx, "outer_tx_hash")
			if err != nil {
				return err
			}
			return db.TxManager.RunInTx(outerCtx, func(innerCtx context.Context) error {
				_, err := ownerRepo.UpdatePassword(innerCtx, owner.ID, "inner_tx_hash")
				return err
			})
		})

		require.Error(t, err)
	})

	t.Run("Transaction timeout behavior", func(t *testing.T) {
		// Create context with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		ownerRepo := postgres.NewOwnerRepository(db.Pool)

		err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
			// Create owner
			_, err := ownerRepo.Create(txCtx, "timeout_test_hash")
			if err != nil {
				return err
			}

			// Simulate long operation
			time.Sleep(200 * time.Millisecond)
			return nil
		})

		// Should fail due to context timeout
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context")

		// Verify no data was committed
		_, err = ownerRepo.GetSingle(context.Background())
		assert.Error(t, err) // Should be ErrNotFound
	})
}

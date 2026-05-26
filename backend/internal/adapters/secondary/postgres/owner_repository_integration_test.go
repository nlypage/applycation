package postgres_test

import (
	"context"
	"testing"

	"github.com/nlypage/applycation/backend/internal/adapters/secondary/postgres"
	"github.com/nlypage/applycation/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOwnerRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)
	repo := postgres.NewOwnerRepository(db.Pool)

	t.Run("Create and GetSingle", func(t *testing.T) {
		ctx := context.Background()
		passwordHash := "hashed_password_123"

		// Create owner
		owner, err := repo.Create(ctx, passwordHash)
		require.NoError(t, err)
		assert.NotEmpty(t, owner.ID)
		assert.Equal(t, passwordHash, owner.PasswordHash)
		assert.False(t, owner.CreatedAt.IsZero())
		assert.False(t, owner.UpdatedAt.IsZero())

		// Get single owner
		retrieved, err := repo.GetSingle(ctx)
		require.NoError(t, err)
		assert.Equal(t, owner.ID, retrieved.ID)
		assert.Equal(t, owner.PasswordHash, retrieved.PasswordHash)
	})

	t.Run("GetByID", func(t *testing.T) {
		ctx := context.Background()
		passwordHash := "another_hash_456"

		// Create owner
		owner, err := repo.Create(ctx, passwordHash)
		require.NoError(t, err)

		// Get by ID
		retrieved, err := repo.GetByID(ctx, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, owner.ID, retrieved.ID)
		assert.Equal(t, owner.PasswordHash, retrieved.PasswordHash)
	})

	t.Run("UpdatePassword", func(t *testing.T) {
		ctx := context.Background()
		originalHash := "original_hash"
		newHash := "new_hash_789"

		// Create owner
		owner, err := repo.Create(ctx, originalHash)
		require.NoError(t, err)

		// Update password
		updated, err := repo.UpdatePassword(ctx, owner.ID, newHash)
		require.NoError(t, err)
		assert.Equal(t, owner.ID, updated.ID)
		assert.Equal(t, newHash, updated.PasswordHash)
		assert.True(t, updated.PasswordChangedAt.After(owner.PasswordChangedAt))

		// Verify update persisted
		retrieved, err := repo.GetByID(ctx, owner.ID)
		require.NoError(t, err)
		assert.Equal(t, newHash, retrieved.PasswordHash)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		ctx := context.Background()
		nonExistentID := "550e8400-e29b-41d4-a716-446655440000"

		_, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		// Should return ErrNotFound from secondary ports
	})

	t.Run("UpdatePassword_NotFound", func(t *testing.T) {
		ctx := context.Background()
		nonExistentID := "550e8400-e29b-41d4-a716-446655440000"

		_, err := repo.UpdatePassword(ctx, nonExistentID, "new_hash")
		assert.Error(t, err)
		// Should return ErrNotFound from secondary ports
	})
}

func TestOwnerRepository_WithTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)
	repo := postgres.NewOwnerRepository(db.Pool)

	t.Run("Transaction rollback", func(t *testing.T) {
		ctx := context.Background()
		passwordHash := "tx_test_hash"

		// Test that transaction rollback works
		err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
			// Create owner within transaction
			owner, err := repo.Create(txCtx, passwordHash)
			require.NoError(t, err)
			assert.NotEmpty(t, owner.ID)

			// Verify owner exists within transaction
			retrieved, err := repo.GetByID(txCtx, owner.ID)
			require.NoError(t, err)
			assert.Equal(t, owner.ID, retrieved.ID)

			// Return error to trigger rollback
			return assert.AnError
		})

		// Transaction should have failed
		require.Error(t, err)

		// Owner should not exist after rollback
		_, err = repo.GetSingle(ctx)
		assert.Error(t, err) // Should be ErrNotFound
	})

	t.Run("Transaction commit", func(t *testing.T) {
		ctx := context.Background()
		passwordHash := "tx_commit_hash"

		var ownerID string

		// Test successful transaction
		err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
			owner, err := repo.Create(txCtx, passwordHash)
			require.NoError(t, err)
			ownerID = owner.ID
			return nil
		})

		require.NoError(t, err)

		// Owner should exist after commit
		retrieved, err := repo.GetByID(ctx, ownerID)
		require.NoError(t, err)
		assert.Equal(t, passwordHash, retrieved.PasswordHash)
	})
}

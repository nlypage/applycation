package testutil_test

import (
	"context"
	"testing"

	"github.com/nlypage/applycation/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTestDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testcontainers test")
	}

	db := testutil.SetupTestDatabase(t)

	t.Run("Database connection works", func(t *testing.T) {
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

	t.Run("Transaction manager works", func(t *testing.T) {
		ctx := context.Background()

		err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
			// Simple transaction test
			_, err := db.Pool.Exec(txCtx, "SELECT 1")
			return err
		})

		require.NoError(t, err)
	})
}

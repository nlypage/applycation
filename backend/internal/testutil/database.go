package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	postgresadapter "github.com/nlypage/applycation/backend/internal/adapters/secondary/postgres"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDatabase provides a test database instance for integration tests.
type TestDatabase struct {
	Pool      *pgxpool.Pool
	TxManager *postgresadapter.TxManager
	URL       string
}

// SetupTestDatabase creates a test database connection using testcontainers.
// If TEST_DATABASE_URL is provided, uses external database instead.
func SetupTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dbURL := os.Getenv("TEST_DATABASE_URL"); dbURL != "" {
		return setupExternalDatabase(t, dbURL)
	}

	container, err := postgrescontainer.Run(ctx,
		"postgres:16-alpine",
		postgrescontainer.WithDatabase("applycation_test"),
		postgrescontainer.WithUsername("postgres"),
		postgrescontainer.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	db := newTestDatabase(t, connStr)

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		db.Pool.Close()
		if err := container.Terminate(cleanupCtx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	return db
}

func setupExternalDatabase(t *testing.T, dbURL string) *TestDatabase {
	t.Helper()

	db := newTestDatabase(t, dbURL)
	// Внешнюю БД чистим между тестами, контейнер не трогаем.
	t.Cleanup(func() {
		cleanupDatabase(t, db.Pool)
		db.Pool.Close()
	})
	return db
}

func newTestDatabase(t *testing.T, connStr string) *TestDatabase {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create database pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping database: %v", err)
	}
	if err := runMigrations(connStr); err != nil {
		pool.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &TestDatabase{
		Pool:      pool,
		TxManager: postgresadapter.NewTxManager(pool),
		URL:       connStr,
	}
}

// runMigrations applies db/migrations using goose library.
func runMigrations(connStr string) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}
	migrationsDir := filepath.Join(projectRoot, "db", "migrations")

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// cleanupDatabase truncates known tables for test isolation.
func cleanupDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename NOT LIKE 'pg_%'
	`)
	if err != nil {
		t.Logf("failed to get table names for cleanup: %v", err)
		return
	}
	defer rows.Close()

	allowed := map[string]bool{
		"goose_db_version": true,
		"owners":           true,
		"owner_sessions":   true,
		"credentials":      true,
	}

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Logf("failed to scan table name: %v", err)
			continue
		}
		if !allowed[table] {
			continue
		}
		if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+table+" CASCADE"); err != nil {
			t.Logf("failed to truncate table %s: %v", table, err)
		}
	}
}

// findProjectRoot finds repository root that contains db/migrations.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "db", "migrations")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("project root with db/migrations not found")
}

// WithTestTx runs fn within transaction and rolls back at the end.
func (db *TestDatabase) WithTestTx(t *testing.T, fn func(ctx context.Context)) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rollback := errors.New("rollback test transaction")
	err := db.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
		fn(txCtx)
		return rollback
	})
	if err == nil || !errors.Is(err, rollback) {
		t.Fatalf("failed to rollback test transaction: %v", err)
	}
}

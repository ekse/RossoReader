package pgstoretest

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ekse/rossoreader/internal/bootstrap"
	"github.com/ekse/rossoreader/internal/db"
	"github.com/ekse/rossoreader/internal/domain"
	"github.com/ekse/rossoreader/internal/store/pgstore"
)

func runMigrationsRetry(t *testing.T, connStr string) {
	t.Helper()
	var lastErr error
	for i := 0; i < 10; i++ {
		if err := db.RunMigrations(connStr); err == nil {
			return
		} else {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
		}
	}
	require.NoError(t, lastErr)
}

func SetupTestStore(t *testing.T) (*pgstore.PGStore, *pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("rssreader"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	runMigrationsRetry(t, connStr)
	if t.Failed() {
		pgContainer.Terminate(ctx)
		t.FailNow()
	}

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	store := pgstore.New(pool)

	if err := bootstrap.Run(ctx, pool, store); err != nil {
		pool.Close()
		pgContainer.Terminate(ctx)
		t.Fatalf("bootstrap failed: %v", err)
	}

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return store, pool, cleanup
}

func CreateTestUser(t *testing.T, ctx context.Context, store *pgstore.PGStore, username string) domain.User {
	t.Helper()
	u, err := store.CreateUser(ctx, username, "test-hash-not-secret", false)
	require.NoError(t, err)
	return u
}

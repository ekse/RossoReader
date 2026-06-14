package pgstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/store/pgstore"
)

func setupTestStore(t *testing.T) (*pgstore.PGStore, func()) {
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

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Run migrations
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS feeds (
			id              SERIAL PRIMARY KEY,
			url             TEXT NOT NULL UNIQUE,
			title           TEXT NOT NULL,
			description     TEXT,
			site_link       TEXT,
			last_fetched_at TIMESTAMPTZ,
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS items (
			id           SERIAL PRIMARY KEY,
			feed_id      INTEGER NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
			guid         TEXT NOT NULL,
			title        TEXT NOT NULL,
			url          TEXT NOT NULL,
			content      TEXT,
			description  TEXT,
			author       TEXT,
			published_at TIMESTAMPTZ,
			fetched_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			read         BOOLEAN NOT NULL DEFAULT FALSE,
			starred      BOOLEAN NOT NULL DEFAULT FALSE,
			UNIQUE(feed_id, guid)
		);
		CREATE INDEX IF NOT EXISTS idx_items_feed_id ON items(feed_id);
		CREATE INDEX IF NOT EXISTS idx_items_published_at ON items(published_at DESC);
		CREATE INDEX IF NOT EXISTS idx_items_read ON items(read);
		CREATE INDEX IF NOT EXISTS idx_items_starred ON items(starred);
		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	store := pgstore.New(pool)

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return store, cleanup
}

func TestPGStore_CreateAndGetFeed(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	feed, err := store.CreateFeed(ctx, "https://example.com/rss", "Example Blog", "An example blog", "https://example.com", "")
	require.NoError(t, err)
	assert.Equal(t, "Example Blog", feed.Title)
	assert.Equal(t, "https://example.com/rss", feed.URL)
	assert.NotZero(t, feed.ID)

	got, err := store.GetFeed(ctx, feed.ID)
	require.NoError(t, err)
	assert.Equal(t, feed.Title, got.Title)
	assert.Equal(t, feed.URL, got.URL)
}

func TestPGStore_GetFeeds(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	_, err := store.CreateFeed(ctx, "https://a.com/rss", "A", "", "https://a.com", "")
	require.NoError(t, err)
	_, err = store.CreateFeed(ctx, "https://b.com/rss", "B", "", "https://b.com", "")
	require.NoError(t, err)

	feeds, err := store.GetFeeds(ctx)
	require.NoError(t, err)
	assert.Len(t, feeds, 2)
}

func TestPGStore_DeleteFeed(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	feed, err := store.CreateFeed(ctx, "https://example.com/rss", "Example", "", "", "")
	require.NoError(t, err)

	err = store.DeleteFeed(ctx, feed.ID)
	require.NoError(t, err)

	_, err = store.GetFeed(ctx, feed.ID)
	assert.Error(t, err)
}

func TestPGStore_UpsertItem(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	feed, err := store.CreateFeed(ctx, "https://example.com/rss", "Example", "", "", "")
	require.NoError(t, err)

	now := time.Now()
	item := domain.Item{
		FeedID:      feed.ID,
		GUID:        "post-1",
		Title:       "First Post",
		URL:         "https://example.com/post-1",
		Content:     strPtr("Full content"),
		Description: strPtr("A description"),
		Author:      strPtr("Author"),
		PublishedAt: &now,
	}

	created, err := store.UpsertItem(ctx, item)
	require.NoError(t, err)
	assert.Equal(t, "First Post", created.Title)
	assert.NotZero(t, created.ID)

	// Upsert again (same GUID), should update
	item.Title = "Updated Post"
	updated, err := store.UpsertItem(ctx, item)
	require.NoError(t, err)
	assert.Equal(t, "Updated Post", updated.Title)
	assert.Equal(t, created.ID, updated.ID)
}

func TestPGStore_GetItemsWithFilters(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	feed, err := store.CreateFeed(ctx, "https://example.com/rss", "Example", "", "", "")
	require.NoError(t, err)

	now := time.Now()
	store.UpsertItem(ctx, domain.Item{FeedID: feed.ID, GUID: "1", Title: "Post 1", URL: "https://ex.com/1", PublishedAt: &now})
	store.UpsertItem(ctx, domain.Item{FeedID: feed.ID, GUID: "2", Title: "Post 2", URL: "https://ex.com/2", PublishedAt: &now})

	items, total, err := store.GetItems(ctx, domain.ItemsQuery{PerPage: 10, Page: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
}

func TestPGStore_Settings(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	err := store.UpsertSetting(ctx, "fetch_interval", "60")
	require.NoError(t, err)

	val, err := store.GetSetting(ctx, "fetch_interval")
	require.NoError(t, err)
	assert.Equal(t, "60", val)

	settings, err := store.GetSettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, "60", settings["fetch_interval"])
}

func strPtr(s string) *string {
	return &s
}

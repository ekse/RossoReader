package pgstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/ekse/rssreader/internal/bootstrap"
	"github.com/ekse/rssreader/internal/db"
	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/store/pgstore"
)

// runMigrationsRetry applies migrations with retries to accommodate the
// postgres container briefly rejecting connections right after start.
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

func setupTestStore(t *testing.T) (*pgstore.PGStore, *pgxpool.Pool, func()) {
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

	// Run bootstrap to finalize schema (drop legacy constraints/columns).
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

func createTestUser(t *testing.T, ctx context.Context, store *pgstore.PGStore, username string) domain.User {
	t.Helper()
	u, err := store.CreateUser(ctx, username, "test-hash-not-secret", false)
	require.NoError(t, err)
	return u
}

func TestPGStore_CreateAndGetFeed(t *testing.T) {
	store, _, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, ctx, store, "alice")

	feed, err := store.CreateFeed(ctx, user.ID, "https://example.com/rss", "Example Blog", "An example blog", "https://example.com", "", "")
	require.NoError(t, err)
	assert.Equal(t, "Example Blog", feed.Title)
	assert.Equal(t, "https://example.com/rss", feed.URL)
	assert.Equal(t, user.ID, feed.UserID)
	assert.NotZero(t, feed.ID)

	got, err := store.GetFeed(ctx, user.ID, feed.ID)
	require.NoError(t, err)
	assert.Equal(t, feed.Title, got.Title)
	assert.Equal(t, feed.URL, got.URL)
}

func TestPGStore_GetFeeds(t *testing.T) {
	store, _, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	alice := createTestUser(t, ctx, store, "alice")
	bob := createTestUser(t, ctx, store, "bob")

	_, err := store.CreateFeed(ctx, alice.ID, "https://a.com/rss", "A", "", "https://a.com", "", "")
	require.NoError(t, err)
	_, err = store.CreateFeed(ctx, alice.ID, "https://b.com/rss", "B", "", "https://b.com", "", "")
	require.NoError(t, err)
	_, err = store.CreateFeed(ctx, bob.ID, "https://c.com/rss", "C", "", "https://c.com", "", "")
	require.NoError(t, err)

	feeds, err := store.GetFeeds(ctx, alice.ID)
	require.NoError(t, err)
	assert.Len(t, feeds, 2)

	allFeeds, err := store.GetAllFeeds(ctx)
	require.NoError(t, err)
	assert.Len(t, allFeeds, 3)
}

func TestPGStore_DeleteFeed(t *testing.T) {
	store, _, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	alice := createTestUser(t, ctx, store, "alice")
	bob := createTestUser(t, ctx, store, "bob")

	feed, err := store.CreateFeed(ctx, alice.ID, "https://example.com/rss", "Example", "", "", "", "")
	require.NoError(t, err)

	// bob's delete attempt does nothing (no rows match user_id filter)
	_ = store.DeleteFeed(ctx, bob.ID, feed.ID)

	// alice's feed is still present
	_, err = store.GetFeed(ctx, alice.ID, feed.ID)
	require.NoError(t, err)

	// alice deletes her own feed
	err = store.DeleteFeed(ctx, alice.ID, feed.ID)
	require.NoError(t, err)

	_, err = store.GetFeed(ctx, alice.ID, feed.ID)
	assert.Error(t, err)
}

func TestPGStore_UpsertItem(t *testing.T) {
	store, _, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, ctx, store, "alice")

	feed, err := store.CreateFeed(ctx, user.ID, "https://example.com/rss", "Example", "", "", "", "")
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

	item.Title = "Updated Post"
	updated, err := store.UpsertItem(ctx, item)
	require.NoError(t, err)
	assert.Equal(t, "Updated Post", updated.Title)
	assert.Equal(t, created.ID, updated.ID)
}

func TestPGStore_GetItemsWithFilters(t *testing.T) {
	store, _, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, ctx, store, "alice")

	feed, err := store.CreateFeed(ctx, user.ID, "https://example.com/rss", "Example", "", "", "", "")
	require.NoError(t, err)

	now := time.Now()
	store.UpsertItem(ctx, domain.Item{FeedID: feed.ID, GUID: "1", Title: "Post 1", URL: "https://ex.com/1", PublishedAt: &now})
	item2, _ := store.UpsertItem(ctx, domain.Item{FeedID: feed.ID, GUID: "2", Title: "Post 2", URL: "https://ex.com/2", PublishedAt: &now})

	// Mark item 2 as read for the user.
	require.NoError(t, store.MarkItemRead(ctx, user.ID, item2.ID, true))

	items, total, err := store.GetItems(ctx, domain.ItemsQuery{UserID: user.ID, PerPage: 10, Page: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// Filter unread only.
	unreadFilter := false
	items, total, err = store.GetItems(ctx, domain.ItemsQuery{UserID: user.ID, PerPage: 10, Page: 1, Read: &unreadFilter})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, "Post 1", items[0].Title)
	assert.False(t, items[0].Read)
	assert.False(t, items[0].Starred)

	// Other user sees no items.
	bob := createTestUser(t, ctx, store, "bob")
	bobItems, _, err := store.GetItems(ctx, domain.ItemsQuery{UserID: bob.ID, PerPage: 10, Page: 1})
	require.NoError(t, err)
	assert.Empty(t, bobItems)
}

func TestPGStore_Settings(t *testing.T) {
	store, _, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	user := createTestUser(t, ctx, store, "alice")
	other := createTestUser(t, ctx, store, "bob")

	err := store.UpsertSetting(ctx, user.ID, "fetch_interval", "60")
	require.NoError(t, err)

	val, err := store.GetSetting(ctx, user.ID, "fetch_interval")
	require.NoError(t, err)
	assert.Equal(t, "60", val)

	settings, err := store.GetSettings(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "60", settings["fetch_interval"])

	// Bob cannot see alice's setting.
	bobSettings, _ := store.GetSettings(ctx, other.ID)
	assert.Empty(t, bobSettings)

	// Bob can have their own.
	store.UpsertSetting(ctx, other.ID, "fetch_interval", "15")
	val2, _ := store.GetSetting(ctx, other.ID, "fetch_interval")
	assert.Equal(t, "15", val2)
}

func strPtr(s string) *string {
	return &s
}

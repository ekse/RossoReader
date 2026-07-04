package scheduler_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/fetcher"
	"github.com/ekse/rssreader/internal/scheduler"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

type mockFetcher struct {
	items       []domain.Item
	title       string
	description string
	siteLink    string
	etag        string
	notModified bool
	err         error
}

func (m *mockFetcher) Fetch(_ context.Context, _, _ string) (*fetcher.FetchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.notModified {
		return &fetcher.FetchResult{NotModified: true}, nil
	}
	return &fetcher.FetchResult{
		Items:       m.items,
		Title:       m.title,
		Description: m.description,
		SiteLink:    m.siteLink,
		Etag:        m.etag,
	}, nil
}

func TestScheduler_FetchAll(t *testing.T) {
	store := mockstore.New()
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
	})

	now := time.Now()
	mf := &mockFetcher{
		items: []domain.Item{
			{GUID: "1", Title: "Post 1", URL: "https://example.com/1", PublishedAt: &now},
			{GUID: "2", Title: "Post 2", URL: "https://example.com/2", PublishedAt: &now},
		},
		title:       "Example Blog",
		description: "A blog",
		siteLink:    "https://example.com",
	}

	s := scheduler.New(store, mf)
	err := s.FetchAll(context.Background())
	require.NoError(t, err)

	assert.Len(t, store.Items, 2)
	assert.Equal(t, "Post 1", store.Items[0].Title)
	assert.Equal(t, "Post 2", store.Items[1].Title)
	assert.NotNil(t, store.Feeds[0].LastFetchedAt)
}

func TestScheduler_FetchFeed_UpdatesMetadata(t *testing.T) {
	store := mockstore.New()
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Old Title",
	})

	mf := &mockFetcher{
		items:       []domain.Item{},
		title:       "New Title",
		description: "New description",
		siteLink:    "https://example.com",
	}

	s := scheduler.New(store, mf)
	err := s.FetchFeed(context.Background(), store.Feeds[0])
	require.NoError(t, err)

	assert.Equal(t, "New Title", store.Feeds[0].Title)
}

func TestScheduler_FetchFeed_NotModified(t *testing.T) {
	store := mockstore.New()
	now := time.Now()
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
		LastFetchedAt: &now,
	})

	mf := &mockFetcher{notModified: true}

	s := scheduler.New(store, mf)
	err := s.FetchFeed(context.Background(), store.Feeds[0])
	require.NoError(t, err)

	assert.Len(t, store.Items, 0)
	assert.NotNil(t, store.Feeds[0].LastFetchedAt)
	assert.True(t, store.Feeds[0].LastFetchedAt.After(now))
}

func TestScheduler_FetchFeed_FetcherError(t *testing.T) {
	store := mockstore.New()
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
	})

	mf := &mockFetcher{err: assert.AnError}

	s := scheduler.New(store, mf)
	err := s.FetchFeed(context.Background(), store.Feeds[0])
	assert.Error(t, err)
	assert.NotNil(t, store.Feeds[0].LastFetchError)
	assert.Contains(t, *store.Feeds[0].LastFetchError, assert.AnError.Error())
}

func TestScheduler_FetchFeed_ClearsErrorOnSuccess(t *testing.T) {
	store := mockstore.New()
	now := time.Now()
	errStr := "previous error"
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
		LastFetchedAt: &now, LastFetchError: &errStr,
	})

	mf := &mockFetcher{
		items: []domain.Item{
			{GUID: "1", Title: "Post 1", URL: "https://example.com/1", PublishedAt: &now},
		},
	}

	s := scheduler.New(store, mf)
	err := s.FetchFeed(context.Background(), store.Feeds[0])
	require.NoError(t, err)
	assert.Nil(t, store.Feeds[0].LastFetchError)
}

func TestScheduler_FetchFeed_ClearsErrorOnNotModified(t *testing.T) {
	store := mockstore.New()
	now := time.Now()
	errStr := "previous error"
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
		LastFetchedAt: &now, LastFetchError: &errStr,
	})

	mf := &mockFetcher{notModified: true}

	s := scheduler.New(store, mf)
	err := s.FetchFeed(context.Background(), store.Feeds[0])
	require.NoError(t, err)
	assert.Nil(t, store.Feeds[0].LastFetchError)
}

func TestScheduler_PurgeAll_DeletesExcess(t *testing.T) {
	store := mockstore.New()
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
	})
	store.SetItemsLimit(context.Background(), 3)

	now := time.Now()
	for i := 0; i < 5; i++ {
		store.Items = append(store.Items, domain.Item{
			ID: int64(i + 1), FeedID: 1, GUID: fmt.Sprintf("g%d", i),
			Title: fmt.Sprintf("Item %d", i), URL: fmt.Sprintf("https://ex.com/%d", i),
			PublishedAt: &now,
		})
	}

	mf := &mockFetcher{}
	s := scheduler.New(store, mf)
	err := s.PurgeAll(context.Background())
	require.NoError(t, err)

	assert.Len(t, store.Items, 3)
}

func TestScheduler_PurgeAll_KeepsStarred(t *testing.T) {
	store := mockstore.New()
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
	})
	store.SetItemsLimit(context.Background(), 2)

	now := time.Now()
	store.Items = append(store.Items, domain.Item{
		ID: 1, FeedID: 1, GUID: "1", Title: "Newest",
		URL: "https://ex.com/1", PublishedAt: &now,
	})
	past := now.Add(-time.Hour)
	store.Items = append(store.Items, domain.Item{
		ID: 2, FeedID: 1, GUID: "2", Title: "Old",
		URL: "https://ex.com/2", PublishedAt: &past,
	})
	farPast := now.Add(-2 * time.Hour)
	store.Items = append(store.Items, domain.Item{
		ID: 3, FeedID: 1, GUID: "3", Title: "Oldest",
		URL: "https://ex.com/3", PublishedAt: &farPast,
	})

	// User 1 stars the oldest item — it should be kept even though it's beyond the limit.
	store.MarkItemStarred(context.Background(), 1, 3, true)

	mf := &mockFetcher{}
	s := scheduler.New(store, mf)
	err := s.PurgeAll(context.Background())
	require.NoError(t, err)

	// 2 newest + 1 starred (oldest) = 3 items kept.
	assert.Len(t, store.Items, 3)
}

func TestScheduler_PurgeAll_UnderLimit(t *testing.T) {
	store := mockstore.New()
	store.Feeds = append(store.Feeds, domain.Feed{
		ID: 1, UserID: 1, URL: "https://example.com/rss", Title: "Example",
	})
	store.SetItemsLimit(context.Background(), 10)

	now := time.Now()
	for i := 0; i < 5; i++ {
		store.Items = append(store.Items, domain.Item{
			ID: int64(i + 1), FeedID: 1, GUID: fmt.Sprintf("g%d", i),
			Title: fmt.Sprintf("Item %d", i), URL: fmt.Sprintf("https://ex.com/%d", i),
			PublishedAt: &now,
		})
	}

	mf := &mockFetcher{}
	s := scheduler.New(store, mf)
	err := s.PurgeAll(context.Background())
	require.NoError(t, err)

	assert.Len(t, store.Items, 5)
}

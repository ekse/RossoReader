package store

import (
	"context"

	"github.com/ekse/rssreader/internal/domain"
)

type Store interface {
	// Feeds
	GetFeeds(ctx context.Context) ([]domain.Feed, error)
	GetFeed(ctx context.Context, id int64) (domain.Feed, error)
	CreateFeed(ctx context.Context, url, title, description, siteLink, iconURL string) (domain.Feed, error)
	DeleteFeed(ctx context.Context, id int64) error
	UpdateFeedLastFetched(ctx context.Context, id int64) error
	UpdateFeedMetadata(ctx context.Context, id int64, title, description, siteLink, iconURL string) error

	// Items
	GetItems(ctx context.Context, q domain.ItemsQuery) ([]domain.Item, int64, error)
	GetItem(ctx context.Context, id int64) (domain.Item, error)
	UpsertItem(ctx context.Context, item domain.Item) (domain.Item, error)
	MarkItemRead(ctx context.Context, id int64, read bool) error
	MarkItemStarred(ctx context.Context, id int64, starred bool) error
	MarkAllFeedItemsRead(ctx context.Context, feedID int64) error
	MarkAllItemsRead(ctx context.Context) error
	GetUnreadCountByFeed(ctx context.Context) (map[int64]int, error)

	// Settings
	GetSettings(ctx context.Context) (map[string]string, error)
	GetSetting(ctx context.Context, key string) (string, error)
	UpsertSetting(ctx context.Context, key, value string) error
	DeleteSetting(ctx context.Context, key string) error
}

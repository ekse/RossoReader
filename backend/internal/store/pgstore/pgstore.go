package pgstore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ekse/rssreader/internal/db/generated"
	"github.com/ekse/rssreader/internal/domain"
)

type PGStore struct {
	q *generated.Queries
}

func New(pool *pgxpool.Pool) *PGStore {
	return &PGStore{q: generated.New(pool)}
}

func (s *PGStore) GetFeeds(ctx context.Context) ([]domain.Feed, error) {
	rows, err := s.q.GetFeeds(ctx)
	if err != nil {
		return nil, fmt.Errorf("get feeds: %w", err)
	}
	feeds := make([]domain.Feed, 0, len(rows))
	for _, r := range rows {
		feeds = append(feeds, toDomainFeed(r))
	}
	return feeds, nil
}

func (s *PGStore) GetFeed(ctx context.Context, id int64) (domain.Feed, error) {
	r, err := s.q.GetFeedByID(ctx, int32(id))
	if err != nil {
		return domain.Feed{}, fmt.Errorf("get feed %d: %w", id, err)
	}
	return toDomainFeed(r), nil
}

func (s *PGStore) CreateFeed(ctx context.Context, url, title, description, siteLink string) (domain.Feed, error) {
	r, err := s.q.CreateFeed(ctx, generated.CreateFeedParams{
		Url:         url,
		Title:       title,
		Description: nullableString(description),
		SiteLink:    nullableString(siteLink),
	})
	if err != nil {
		return domain.Feed{}, fmt.Errorf("create feed: %w", err)
	}
	return toDomainFeed(r), nil
}

func (s *PGStore) DeleteFeed(ctx context.Context, id int64) error {
	err := s.q.DeleteFeed(ctx, int32(id))
	if err != nil {
		return fmt.Errorf("delete feed %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) UpdateFeedLastFetched(ctx context.Context, id int64) error {
	err := s.q.UpdateFeedLastFetched(ctx, int32(id))
	if err != nil {
		return fmt.Errorf("update feed last fetched %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) UpdateFeedMetadata(ctx context.Context, id int64, title, description, siteLink string) error {
	err := s.q.UpdateFeedMetadata(ctx, generated.UpdateFeedMetadataParams{
		ID:          int32(id),
		Title:       title,
		Description: nullableString(description),
		SiteLink:    nullableString(siteLink),
	})
	if err != nil {
		return fmt.Errorf("update feed metadata %d: %w", id, err)
	}
	return nil
}

func (s *PGStore) GetItems(ctx context.Context, q domain.ItemsQuery) ([]domain.Item, int64, error) {
	limit := int32(q.PerPage)
	if limit <= 0 {
		limit = 20
	}
	offset := int32((q.Page - 1) * q.PerPage)
	if offset < 0 {
		offset = 0
	}
	var feedID *int32
	if q.FeedID != nil {
		v := int32(*q.FeedID)
		feedID = &v
	}

	params := generated.GetItemsParams{
		Limit:  limit,
		Offset: offset,
		FeedID: feedID,
		Read:   q.Read,
		Starred: q.Starred,
	}

	rows, err := s.q.GetItems(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("get items: %w", err)
	}

	countParams := generated.CountItemsParams{
		FeedID:  feedID,
		Read:    q.Read,
		Starred: q.Starred,
	}
	total, err := s.q.CountItems(ctx, countParams)
	if err != nil {
		return nil, 0, fmt.Errorf("count items: %w", err)
	}

	items := make([]domain.Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, toDomainItem(r))
	}
	return items, total, nil
}

func (s *PGStore) GetItem(ctx context.Context, id int64) (domain.Item, error) {
	r, err := s.q.GetItemByID(ctx, int32(id))
	if err != nil {
		return domain.Item{}, fmt.Errorf("get item %d: %w", id, err)
	}
	return toDomainItem(r), nil
}

func (s *PGStore) UpsertItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	r, err := s.q.UpsertItem(ctx, generated.UpsertItemParams{
		FeedID:      int32(item.FeedID),
		Guid:        item.GUID,
		Title:       item.Title,
		Url:         item.URL,
		Content:     item.Content,
		Description: item.Description,
		Author:      item.Author,
		PublishedAt: toTimestamptz(item.PublishedAt),
	})
	if err != nil {
		return domain.Item{}, fmt.Errorf("upsert item: %w", err)
	}
	return toDomainItem(r), nil
}

func (s *PGStore) MarkItemRead(ctx context.Context, id int64, read bool) error {
	err := s.q.MarkItemRead(ctx, generated.MarkItemReadParams{
		ID:   int32(id),
		Read: read,
	})
	if err != nil {
		return fmt.Errorf("mark item %d read: %w", id, err)
	}
	return nil
}

func (s *PGStore) MarkItemStarred(ctx context.Context, id int64, starred bool) error {
	err := s.q.MarkItemStarred(ctx, generated.MarkItemStarredParams{
		ID:      int32(id),
		Starred: starred,
	})
	if err != nil {
		return fmt.Errorf("mark item %d starred: %w", id, err)
	}
	return nil
}

func (s *PGStore) MarkAllFeedItemsRead(ctx context.Context, feedID int64) error {
	err := s.q.MarkFeedItemsRead(ctx, int32(feedID))
	if err != nil {
		return fmt.Errorf("mark all feed items read %d: %w", feedID, err)
	}
	return nil
}

func (s *PGStore) GetUnreadCountByFeed(ctx context.Context) (map[int64]int, error) {
	rows, err := s.q.GetUnreadCountByFeed(ctx)
	if err != nil {
		return nil, fmt.Errorf("get unread count by feed: %w", err)
	}
	result := make(map[int64]int, len(rows))
	for _, r := range rows {
		result[int64(r.FeedID)] = int(r.Count)
	}
	return result, nil
}

func (s *PGStore) GetSettings(ctx context.Context) (map[string]string, error) {
	rows, err := s.q.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	result := make(map[string]string, len(rows))
	for _, r := range rows {
		result[r.Key] = r.Value
	}
	return result, nil
}

func (s *PGStore) GetSetting(ctx context.Context, key string) (string, error) {
	r, err := s.q.GetSetting(ctx, key)
	if err != nil {
		return "", fmt.Errorf("get setting %q: %w", key, err)
	}
	return r.Value, nil
}

func (s *PGStore) UpsertSetting(ctx context.Context, key, value string) error {
	err := s.q.UpsertSetting(ctx, generated.UpsertSettingParams{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("upsert setting %q: %w", key, err)
	}
	return nil
}

func (s *PGStore) DeleteSetting(ctx context.Context, key string) error {
	err := s.q.DeleteSetting(ctx, key)
	if err != nil {
		return fmt.Errorf("delete setting %q: %w", key, err)
	}
	return nil
}

func toDomainFeed(r generated.Feed) domain.Feed {
	return domain.Feed{
		ID:            int64(r.ID),
		URL:           r.Url,
		Title:         r.Title,
		Description:   r.Description,
		SiteLink:      r.SiteLink,
		LastFetchedAt: fromTimestamptz(r.LastFetchedAt),
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
	}
}

func toDomainItem(r generated.Item) domain.Item {
	return domain.Item{
		ID:          int64(r.ID),
		FeedID:      int64(r.FeedID),
		GUID:        r.Guid,
		Title:       r.Title,
		URL:         r.Url,
		Content:     r.Content,
		Description: r.Description,
		Author:      r.Author,
		PublishedAt: fromTimestamptz(r.PublishedAt),
		FetchedAt:   r.FetchedAt.Time,
		Read:        r.Read,
		Starred:     r.Starred,
	}
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func fromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

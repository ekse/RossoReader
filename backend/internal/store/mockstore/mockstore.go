package mockstore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ekse/rssreader/internal/domain"
)

type MockStore struct {
	mu       sync.Mutex
	Feeds    []domain.Feed
	Items    []domain.Item
	Settings map[string]string
	NextFeedID int64
	NextItemID int64
}

func New() *MockStore {
	return &MockStore{
		Feeds:      []domain.Feed{},
		Items:      []domain.Item{},
		Settings:   make(map[string]string),
		NextFeedID: 1,
		NextItemID: 1,
	}
}

func (m *MockStore) GetFeeds(_ context.Context) ([]domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]domain.Feed, len(m.Feeds))
	copy(result, m.Feeds)
	return result, nil
}

func (m *MockStore) GetFeed(_ context.Context, id int64) (domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range m.Feeds {
		if f.ID == id {
			return f, nil
		}
	}
	return domain.Feed{}, fmt.Errorf("feed %d not found", id)
}

func (m *MockStore) CreateFeed(_ context.Context, url, title, description, siteLink, iconURL string) (domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	f := domain.Feed{
		ID:          m.NextFeedID,
		URL:         url,
		Title:       title,
		Description: strPtr(description),
		SiteLink:    strPtr(siteLink),
		IconURL:     strPtr(iconURL),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.NextFeedID++
	m.Feeds = append(m.Feeds, f)
	return f, nil
}

func (m *MockStore) DeleteFeed(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, f := range m.Feeds {
		if f.ID == id {
			m.Feeds = append(m.Feeds[:i], m.Feeds[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("feed %d not found", id)
}

func (m *MockStore) UpdateFeedLastFetched(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for i, f := range m.Feeds {
		if f.ID == id {
			m.Feeds[i].LastFetchedAt = &now
			return nil
		}
	}
	return fmt.Errorf("feed %d not found", id)
}

func (m *MockStore) UpdateFeedMetadata(_ context.Context, id int64, title, description, siteLink, iconURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, f := range m.Feeds {
		if f.ID == id {
			m.Feeds[i].Title = title
			m.Feeds[i].Description = strPtr(description)
			m.Feeds[i].SiteLink = strPtr(siteLink)
			m.Feeds[i].IconURL = strPtr(iconURL)
			return nil
		}
	}
	return fmt.Errorf("feed %d not found", id)
}

func (m *MockStore) GetItems(_ context.Context, q domain.ItemsQuery) ([]domain.Item, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var filtered []domain.Item
	for _, item := range m.Items {
		if q.FeedID != nil && item.FeedID != *q.FeedID {
			continue
		}
		if q.Read != nil && item.Read != *q.Read {
			continue
		}
		if q.Starred != nil && item.Starred != *q.Starred {
			continue
		}
		filtered = append(filtered, item)
	}
	total := int64(len(filtered))

	start := (q.Page - 1) * q.PerPage
	if start < 0 {
		start = 0
	}
	if start >= len(filtered) {
		return nil, total, nil
	}
	end := start + q.PerPage
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

func (m *MockStore) GetItem(_ context.Context, id int64) (domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.Item{}, fmt.Errorf("item %d not found", id)
}

func (m *MockStore) UpsertItem(_ context.Context, item domain.Item) (domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, existing := range m.Items {
		if existing.FeedID == item.FeedID && existing.GUID == item.GUID {
			m.Items[i].Title = item.Title
			m.Items[i].URL = item.URL
			m.Items[i].Content = item.Content
			m.Items[i].Description = item.Description
			m.Items[i].Author = item.Author
			m.Items[i].PublishedAt = item.PublishedAt
			return m.Items[i], nil
		}
	}
	item.ID = m.NextItemID
	m.NextItemID++
	m.Items = append(m.Items, item)
	return item, nil
}

func (m *MockStore) MarkItemRead(_ context.Context, id int64, read bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i].Read = read
			return nil
		}
	}
	return fmt.Errorf("item %d not found", id)
}

func (m *MockStore) MarkItemStarred(_ context.Context, id int64, starred bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i].Starred = starred
			return nil
		}
	}
	return fmt.Errorf("item %d not found", id)
}

func (m *MockStore) MarkAllFeedItemsRead(_ context.Context, feedID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, item := range m.Items {
		if item.FeedID == feedID {
			m.Items[i].Read = true
		}
	}
	return nil
}

func (m *MockStore) MarkAllItemsRead(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.Items {
		m.Items[i].Read = true
	}
	return nil
}

func (m *MockStore) GetUnreadCountByFeed(_ context.Context) (map[int64]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[int64]int)
	for _, item := range m.Items {
		if !item.Read {
			result[item.FeedID]++
		}
	}
	return result, nil
}

func (m *MockStore) GetSettings(_ context.Context) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]string, len(m.Settings))
	for k, v := range m.Settings {
		result[k] = v
	}
	return result, nil
}

func (m *MockStore) GetSetting(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.Settings[key]
	if !ok {
		return "", fmt.Errorf("setting %q not found", key)
	}
	return v, nil
}

func (m *MockStore) UpsertSetting(_ context.Context, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Settings[key] = value
	return nil
}

func (m *MockStore) DeleteSetting(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Settings, key)
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

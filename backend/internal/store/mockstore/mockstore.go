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
	Settings map[int64]map[string]string
	Users    []domain.User
	Sessions map[[16]byte]domain.Session
	// PasswdHash indexed by user ID
	Passwords map[int64]string

	NextFeedID int64
	NextItemID int64
	NextUserID int64
}

func New() *MockStore {
	return &MockStore{
		Feeds:      []domain.Feed{},
		Items:      []domain.Item{},
		Settings:   make(map[int64]map[string]string),
		Users:      []domain.User{},
		Sessions:   make(map[[16]byte]domain.Session),
		Passwords:  make(map[int64]string),
		NextFeedID: 1,
		NextItemID: 1,
		NextUserID: 1,
	}
}

// Feeds

func (m *MockStore) GetFeeds(_ context.Context, userID int64) ([]domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]domain.Feed, 0, len(m.Feeds))
	for _, f := range m.Feeds {
		if f.UserID == userID {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *MockStore) GetAllFeeds(_ context.Context) ([]domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]domain.Feed, len(m.Feeds))
	copy(result, m.Feeds)
	return result, nil
}

func (m *MockStore) GetFeed(_ context.Context, userID, id int64) (domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range m.Feeds {
		if f.ID == id && f.UserID == userID {
			return f, nil
		}
	}
	return domain.Feed{}, fmt.Errorf("feed %d not found", id)
}

func (m *MockStore) GetFeedByIDAny(_ context.Context, id int64) (domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, f := range m.Feeds {
		if f.ID == id {
			return f, nil
		}
	}
	return domain.Feed{}, fmt.Errorf("feed %d not found", id)
}

func (m *MockStore) CreateFeed(_ context.Context, userID int64, url, title, description, siteLink, iconURL string) (domain.Feed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	f := domain.Feed{
		ID:          m.NextFeedID,
		UserID:      userID,
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

func (m *MockStore) DeleteFeed(_ context.Context, userID, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, f := range m.Feeds {
		if f.ID == id && f.UserID == userID {
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

// Items

func (m *MockStore) GetItems(_ context.Context, q domain.ItemsQuery) ([]domain.Item, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var filtered []domain.Item
	for _, item := range m.Items {
		if !m.userOwnsFeed(q.UserID, item.FeedID) {
			continue
		}
		if q.FeedID != nil && item.FeedID != *q.FeedID {
			continue
		}
		state := m.getItemState(q.UserID, item.ID)
		if q.Read != nil && state.Read != *q.Read {
			continue
		}
		if q.Starred != nil && state.Starred != *q.Starred {
			continue
		}
		it := item
		it.Read = state.Read
		it.Starred = state.Starred
		filtered = append(filtered, it)
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

func (m *MockStore) GetItem(_ context.Context, userID, id int64) (domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.Items {
		if item.ID == id {
			if !m.userOwnsFeed(userID, item.FeedID) {
				return domain.Item{}, fmt.Errorf("item %d not found", id)
			}
			state := m.getItemState(userID, item.ID)
			it := item
			it.Read = state.Read
			it.Starred = state.Starred
			return it, nil
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
	item.Read = false
	item.Starred = false
	m.Items = append(m.Items, item)
	return item, nil
}

func (m *MockStore) MarkItemRead(_ context.Context, userID, id int64, read bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.Items {
		if item.ID == id && m.userOwnsFeed(userID, item.FeedID) {
			m.setItemState(userID, id, read, func(s itemState) itemState {
				s.Read = read
				return s
			})
			return nil
		}
	}
	return fmt.Errorf("item %d not found", id)
}

func (m *MockStore) MarkItemStarred(_ context.Context, userID, id int64, starred bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.Items {
		if item.ID == id && m.userOwnsFeed(userID, item.FeedID) {
			m.setItemState(userID, id, starred, func(s itemState) itemState {
				s.Starred = starred
				return s
			})
			return nil
		}
	}
	return fmt.Errorf("item %d not found", id)
}

func (m *MockStore) MarkAllFeedItemsRead(_ context.Context, userID, feedID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.Items {
		if item.FeedID == feedID && m.userOwnsFeed(userID, feedID) {
			m.setItemStateRead(userID, item.ID, true)
		}
	}
	return nil
}

func (m *MockStore) MarkAllItemsRead(_ context.Context, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, item := range m.Items {
		if m.userOwnsFeed(userID, item.FeedID) {
			m.setItemStateRead(userID, item.ID, true)
		}
	}
	return nil
}

func (m *MockStore) GetUnreadCountByFeed(_ context.Context, userID int64) (map[int64]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[int64]int)
	for _, item := range m.Items {
		if !m.userOwnsFeed(userID, item.FeedID) {
			continue
		}
		if !m.getItemState(userID, item.ID).Read {
			result[item.FeedID]++
		}
	}
	return result, nil
}

// Settings

func (m *MockStore) GetSettings(_ context.Context, userID int64) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]string)
	for k, v := range m.Settings[userID] {
		result[k] = v
	}
	return result, nil
}

func (m *MockStore) GetSetting(_ context.Context, userID int64, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.Settings[userID]; ok {
		if v, ok := s[key]; ok {
			return v, nil
		}
	}
	return "", fmt.Errorf("setting %q not found", key)
}

func (m *MockStore) UpsertSetting(_ context.Context, userID int64, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.Settings[userID]; !ok {
		m.Settings[userID] = make(map[string]string)
	}
	m.Settings[userID][key] = value
	return nil
}

func (m *MockStore) DeleteSetting(_ context.Context, userID int64, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.Settings[userID]; ok {
		delete(s, key)
	}
	return nil
}

// Users

func (m *MockStore) CreateUser(_ context.Context, username, passwordHash string, isAdmin bool) (domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	u := domain.User{
		ID:        m.NextUserID,
		Username:  username,
		IsAdmin:   isAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.NextUserID++
	m.Users = append(m.Users, u)
	m.Passwords[u.ID] = passwordHash
	return u, nil
}

func (m *MockStore) GetUserByUsername(_ context.Context, username string) (domain.User, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.Users {
		if u.Username == username {
			return u, m.Passwords[u.ID], nil
		}
	}
	return domain.User{}, "", fmt.Errorf("user %q not found", username)
}

func (m *MockStore) GetUserByID(_ context.Context, id int64) (domain.User, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.Users {
		if u.ID == id {
			return u, m.Passwords[u.ID], nil
		}
	}
	return domain.User{}, "", fmt.Errorf("user %d not found", id)
}

func (m *MockStore) ListUsers(_ context.Context) ([]domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]domain.User, len(m.Users))
	copy(result, m.Users)
	return result, nil
}

func (m *MockStore) UpdateUserPassword(_ context.Context, id int64, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Passwords[id] = passwordHash
	return nil
}

func (m *MockStore) DeleteUser(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, u := range m.Users {
		if u.ID == id {
			m.Users = append(m.Users[:i], m.Users[i+1:]...)
			delete(m.Passwords, id)
			return nil
		}
	}
	return fmt.Errorf("user %d not found", id)
}

func (m *MockStore) CountUsers(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.Users)), nil
}

// Sessions

func (m *MockStore) CreateSession(_ context.Context, id [16]byte, userID int64, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, _, err := m.getUserByIDLocked(userID)
	if err != nil {
		return err
	}
	m.Sessions[id] = domain.Session{
		ID:        id,
		User:      u,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (m *MockStore) GetSession(_ context.Context, id [16]byte) (domain.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.Sessions[id]
	if !ok {
		return domain.Session{}, fmt.Errorf("session not found")
	}
	if s.ExpiresAt.Before(time.Now()) {
		return domain.Session{}, fmt.Errorf("session expired")
	}
	return s, nil
}

func (m *MockStore) DeleteSession(_ context.Context, id [16]byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Sessions, id)
	return nil
}

func (m *MockStore) DeleteExpiredSessions(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for id, s := range m.Sessions {
		if s.ExpiresAt.Before(now) {
			delete(m.Sessions, id)
		}
	}
	return nil
}

// Helpers

type itemState struct {
	Read    bool
	Starred bool
}

type itemStateKey struct {
	UserID int64
	ItemID int64
}

var itemStates = make(map[itemStateKey]itemState)

func (m *MockStore) userOwnsFeed(userID, feedID int64) bool {
	for _, f := range m.Feeds {
		if f.ID == feedID && f.UserID == userID {
			return true
		}
	}
	return false
}

func (m *MockStore) getItemState(userID, itemID int64) itemState {
	return itemStates[itemStateKey{userID, itemID}]
}

func (m *MockStore) setItemState(userID, itemID int64, _ bool, mutate func(s itemState) itemState) {
	k := itemStateKey{userID, itemID}
	itemStates[k] = mutate(itemStates[k])
}

func (m *MockStore) setItemStateRead(userID, itemID int64, read bool) {
	k := itemStateKey{userID, itemID}
	s := itemStates[k]
	s.Read = read
	itemStates[k] = s
}

func (m *MockStore) getUserByIDLocked(id int64) (domain.User, string, error) {
	for _, u := range m.Users {
		if u.ID == id {
			return u, m.Passwords[u.ID], nil
		}
	}
	return domain.User{}, "", fmt.Errorf("user %d not found", id)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
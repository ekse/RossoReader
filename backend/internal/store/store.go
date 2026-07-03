package store

import (
	"context"
	"time"

	"github.com/ekse/rssreader/internal/domain"
)

type Store interface {
	// Feeds
	GetFeeds(ctx context.Context, userID int64) ([]domain.Feed, error)
	GetAllFeeds(ctx context.Context) ([]domain.Feed, error)
	GetFeed(ctx context.Context, userID, id int64) (domain.Feed, error)
	GetFeedByIDAny(ctx context.Context, id int64) (domain.Feed, error)
	CreateFeed(ctx context.Context, userID int64, url, title, description, siteLink, iconURL, etag string) (domain.Feed, error)
	UpdateFeedEtag(ctx context.Context, id int64, etag string) error
	DeleteFeed(ctx context.Context, userID, id int64) error
	UpdateFeedLastFetched(ctx context.Context, id int64) error
	UpdateFeedMetadata(ctx context.Context, id int64, title, description, siteLink, iconURL string) error

	// Items
	GetItems(ctx context.Context, q domain.ItemsQuery) ([]domain.Item, int64, error)
	GetItem(ctx context.Context, userID, id int64) (domain.Item, error)
	UpsertItem(ctx context.Context, item domain.Item) (domain.Item, error)
	MarkItemRead(ctx context.Context, userID, id int64, read bool) error
	MarkItemStarred(ctx context.Context, userID, id int64, starred bool) error
	MarkAllFeedItemsRead(ctx context.Context, userID, feedID int64) error
	MarkAllItemsRead(ctx context.Context, userID int64) error
	GetUnreadCountByFeed(ctx context.Context, userID int64) (map[int64]int, error)

	// Settings
	GetSettings(ctx context.Context, userID int64) (map[string]string, error)
	GetSetting(ctx context.Context, userID int64, key string) (string, error)
	UpsertSetting(ctx context.Context, userID int64, key, value string) error
	DeleteSetting(ctx context.Context, userID int64, key string) error

	// Users
	CreateUser(ctx context.Context, username, passwordHash string, isAdmin bool) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, string, error)
	GetUserByID(ctx context.Context, id int64) (domain.User, string, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	UpdateUserPassword(ctx context.Context, id int64, passwordHash string) error
	DeleteUser(ctx context.Context, id int64) error
	CountUsers(ctx context.Context) (int64, error)

	// Sessions
	CreateSession(ctx context.Context, id [16]byte, userID int64, expiresAt time.Time) error
	GetSession(ctx context.Context, id [16]byte) (domain.Session, error)
	DeleteSession(ctx context.Context, id [16]byte) error
	DeleteExpiredSessions(ctx context.Context) error

	// Passkeys
	CreatePasskey(ctx context.Context, userID int64, name string, credentialID, publicKey []byte, attestationType string, transports []string, signCount int64, backupEligible, backupState bool, aaguid []byte) (domain.Passkey, error)
	GetPasskeysByUserID(ctx context.Context, userID int64) ([]domain.Passkey, error)
	GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (domain.Passkey, error)
	UpdatePasskeySignCount(ctx context.Context, id int64, signCount int64) error
	DeletePasskey(ctx context.Context, userID, id int64) error

	// Passkey Auth State
	SaveAuthState(ctx context.Context, id [16]byte, stateType string, stateData []byte, expiresAt time.Time) error
	GetAuthState(ctx context.Context, id [16]byte) (string, []byte, error)
	DeleteAuthState(ctx context.Context, id [16]byte) error
	DeleteExpiredAuthStates(ctx context.Context) error
}

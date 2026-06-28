package domain

import (
	"errors"
	"time"
)

var ErrFeedAlreadyExists = errors.New("You are already subscribed to this feed.")
var ErrUserAlreadyExists = errors.New("This username is already taken.")

type Feed struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	URL           string     `json:"url"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	SiteLink      *string    `json:"site_link,omitempty"`
	IconURL       *string    `json:"icon_url,omitempty"`
	LastFetchedAt *time.Time `json:"last_fetched_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Item struct {
	ID          int64      `json:"id"`
	FeedID      int64      `json:"feed_id"`
	GUID        string     `json:"guid"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	Content     *string    `json:"content,omitempty"`
	Description *string    `json:"description,omitempty"`
	Author      *string    `json:"author,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	FetchedAt   time.Time  `json:"fetched_at"`
	Read        bool       `json:"read"`
	Starred     bool       `json:"starred"`
}

type ItemsQuery struct {
	Page    int
	PerPage int
	UserID  int64
	FeedID  *int64
	Read    *bool
	Starred *bool
}

type Settings struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DiscoveredFeed struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID        [16]byte  `json:"-"`
	User      User      `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Passkey struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"-"`
	Name            string    `json:"name"`
	CredentialID    []byte    `json:"-"`
	PublicKey       []byte    `json:"-"`
	AttestationType string    `json:"-"`
	Transports      []string  `json:"transports"`
	SignCount       int64     `json:"-"`
	BackupEligible  bool      `json:"backup_eligible"`
	BackupState     bool      `json:"backup_state"`
	AAGUID          []byte    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"-"`
}

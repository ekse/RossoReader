package domain

import "time"

type Feed struct {
	ID            int64      `json:"id"`
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
	FeedID  *int64
	Read    *bool
	Starred *bool
}

type Settings struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

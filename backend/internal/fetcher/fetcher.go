package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/ekse/rssreader/internal/domain"
)

type Fetcher interface {
	Fetch(ctx context.Context, feedURL string) ([]domain.Item, string, string, string, error)
}

type HTTPFetcher struct {
	client *gofeed.Parser
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		client: gofeed.NewParser(),
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, feedURL string) ([]domain.Item, string, string, string, error) {
	feed, err := f.client.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("parse feed %q: %w", feedURL, err)
	}

	title := feed.Title
	description := ""
	if feed.Description != "" {
		description = feed.Description
	}
	siteLink := ""
	if feed.Link != "" {
		siteLink = feed.Link
	}

	items := make([]domain.Item, 0, len(feed.Items))
	for _, gi := range feed.Items {
		guid := gi.GUID
		if guid == "" {
			guid = gi.Link
		}
		if guid == "" {
			continue
		}

		item := domain.Item{
			GUID:    guid,
			Title:   gi.Title,
			URL:     gi.Link,
			Content: nullString(gi.Content),
			Description: func() *string {
				if gi.Description != "" {
					return &gi.Description
				}
				return nil
			}(),
			Author: func() *string {
				if gi.Author != nil && gi.Author.Name != "" {
					return &gi.Author.Name
				}
				return nil
			}(),
			PublishedAt: parsePublished(gi.PublishedParsed, gi.UpdatedParsed),
		}

		items = append(items, item)
	}

	return items, title, description, siteLink, nil
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parsePublished(published, updated *time.Time) *time.Time {
	if published != nil {
		return published
	}
	return updated
}

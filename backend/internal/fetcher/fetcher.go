package fetcher

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"

	"github.com/ekse/rssreader/internal/domain"
)

type FetchResult struct {
	Items       []domain.Item
	Title       string
	Description string
	SiteLink    string
	Etag        string
	NotModified bool
}

type Fetcher interface {
	Fetch(ctx context.Context, feedURL, etag string) (*FetchResult, error)
}

type Discoverer interface {
	Discover(ctx context.Context, rawURL string) ([]domain.DiscoveredFeed, error)
}

type HTTPFetcher struct {
	parser *gofeed.Parser
	http   *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		parser: gofeed.NewParser(),
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, feedURL, etag string) (*FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "RossoRSSReader/1.0")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := f.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %q: %w", feedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &FetchResult{NotModified: true}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching %q", resp.StatusCode, feedURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	feed, err := f.parser.ParseString(string(body))
	if err != nil {
		return nil, fmt.Errorf("parse feed %q: %w", feedURL, err)
	}

	respEtag := resp.Header.Get("Etag")

	title := html.UnescapeString(feed.Title)
	description := ""
	if feed.Description != "" {
		description = html.UnescapeString(feed.Description)
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
			Title:   html.UnescapeString(gi.Title),
			URL:     gi.Link,
			Content: nullString(gi.Content),
			Description: func() *string {
				if gi.Description != "" {
					s := html.UnescapeString(gi.Description)
					return &s
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

	return &FetchResult{
		Items:       items,
		Title:       title,
		Description: description,
		SiteLink:    siteLink,
		Etag:        respEtag,
	}, nil
}

func (f *HTTPFetcher) Discover(ctx context.Context, rawURL string) ([]domain.DiscoveredFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "RossoRSSReader/1.0")

	resp, err := f.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %q: %w", rawURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	content := string(body)

	if parsed, ok := tryParseRSS(content); ok {
		return []domain.DiscoveredFeed{{URL: rawURL, Title: parsed}}, nil
	}

	return discoverFromHTML(content, rawURL), nil
}

func tryParseRSS(content string) (string, bool) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(content)
	if err != nil {
		return "", false
	}
	return feed.Title, true
}

func discoverFromHTML(content, baseURL string) []domain.DiscoveredFeed {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil
	}

	var feeds []domain.DiscoveredFeed
	doc.Find(`link[rel="alternate"]`).Each(func(_ int, s *goquery.Selection) {
		linkType, _ := s.Attr("type")
		href, _ := s.Attr("href")
		if href == "" {
			return
		}
		if isFeedType(linkType) {
			title, _ := s.Attr("title")
			feeds = append(feeds, domain.DiscoveredFeed{
				URL:   resolveURL(baseURL, href),
				Title: html.UnescapeString(title),
			})
		}
	})
	return feeds
}

func isFeedType(t string) bool {
	t = strings.ToLower(strings.TrimSpace(t))
	return t == "application/rss+xml" || t == "application/atom+xml" || t == "application/rss"
}

func resolveURL(base, ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return baseURL.ResolveReference(refURL).String()
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

package scheduler

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/robfig/cron/v3"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/fetcher"
	"github.com/ekse/rssreader/internal/store"
)

type Scheduler struct {
	store   store.Store
	fetcher fetcher.Fetcher
	cron    *cron.Cron
}

func New(s store.Store, f fetcher.Fetcher) *Scheduler {
	return &Scheduler{
		store:   s,
		fetcher: f,
	}
}

func fetchIntervalMinutes() int {
	v := os.Getenv("FETCH_INTERVAL_MINUTES")
	if v == "" {
		return 30
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 30
	}
	return n
}

func (s *Scheduler) Start() {
	minutes := fetchIntervalMinutes()
	s.cron = cron.New()
	expr := fmt.Sprintf("@every %dm", minutes)
	s.cron.AddFunc(expr, func() {
		log.Printf("scheduler: starting feed fetch cycle")
		ctx := context.Background()
		if err := s.FetchAll(ctx); err != nil {
			log.Printf("scheduler: fetch cycle error: %v", err)
		}
		log.Printf("scheduler: feed fetch cycle complete")
	})
	s.cron.AddFunc("@every 30m", func() {
		ctx := context.Background()
		if err := s.store.DeleteExpiredAuthStates(ctx); err != nil {
			log.Printf("scheduler: auth state cleanup error: %v", err)
		}
	})
	s.cron.Start()
	log.Printf("scheduler: started, fetching every %d minutes", minutes)
}

func (s *Scheduler) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
	}
	log.Printf("scheduler: stopped")
}

func (s *Scheduler) FetchAll(ctx context.Context) error {
	feeds, err := s.store.GetAllFeeds(ctx)
	if err != nil {
		return fmt.Errorf("get feeds: %w", err)
	}

	for _, feed := range feeds {
		if err := s.FetchFeed(ctx, feed); err != nil {
			log.Printf("scheduler: error fetching feed %d (%s): %v", feed.ID, feed.URL, err)
		}
	}

	return nil
}

func (s *Scheduler) FetchFeed(ctx context.Context, feed domain.Feed) error {
	items, title, description, siteLink, err := s.fetcher.Fetch(ctx, feed.URL)
	if err != nil {
		return fmt.Errorf("fetch feed %q: %w", feed.URL, err)
	}

	iconURL := faviconURL(siteLink)
	needsUpdate := title != "" && (title != feed.Title || !ptrEqual(siteLink, feed.SiteLink) ||
		!ptrEqual(description, feed.Description) || !ptrEqual(iconURL, feed.IconURL))
	if needsUpdate {
		if err := s.store.UpdateFeedMetadata(ctx, feed.ID, title, description, siteLink, iconURL); err != nil {
			return fmt.Errorf("update feed metadata: %w", err)
		}
	}

	for _, item := range items {
		item.FeedID = feed.ID
		if _, err := s.store.UpsertItem(ctx, item); err != nil {
			return fmt.Errorf("upsert item %q: %w", item.GUID, err)
		}
	}

	if err := s.store.UpdateFeedLastFetched(ctx, feed.ID); err != nil {
		return fmt.Errorf("update feed last fetched: %w", err)
	}

	return nil
}

func (s *Scheduler) FetchFeedByID(ctx context.Context, feedID int64) error {
	feed, err := s.store.GetFeedByIDAny(ctx, feedID)
	if err != nil {
		return fmt.Errorf("get feed %d: %w", feedID, err)
	}
	return s.FetchFeed(ctx, feed)
}

func faviconURL(siteLink string) string {
	if siteLink == "" {
		return ""
	}
	u, err := url.Parse(siteLink)
	if err != nil || u.Host == "" {
		return ""
	}
	return "https://www.google.com/s2/favicons?domain=" + u.Host + "&sz=32"
}

func ptrEqual(a string, b *string) bool {
	if b == nil {
		return a == ""
	}
	return a == *b
}

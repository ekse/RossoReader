package fetcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/fetcher"
)

const testRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
  <title>Test Blog</title>
  <link>https://example.com</link>
  <description>A test blog</description>
  <item>
    <guid>post-1</guid>
    <title>First Post</title>
    <link>https://example.com/post-1</link>
    <description>This is the first post</description>
    <author>Author One</author>
    <pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
  </item>
  <item>
    <guid>post-2</guid>
    <title>Second Post</title>
    <link>https://example.com/post-2</link>
    <description>This is the second post</description>
    <pubDate>Tue, 02 Jan 2024 00:00:00 GMT</pubDate>
  </item>
</channel>
</rss>`

func TestHTTPFetcher_Fetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(testRSS))
	}))
	defer srv.Close()

	f := fetcher.NewHTTPFetcher()
	items, title, description, siteLink, err := f.Fetch(context.Background(), srv.URL)
	require.NoError(t, err)

	assert.Equal(t, "Test Blog", title)
	assert.Equal(t, "A test blog", description)
	assert.Equal(t, "https://example.com", siteLink)
	assert.Len(t, items, 2)
	assert.Equal(t, "First Post", items[0].Title)
	assert.Equal(t, "post-1", items[0].GUID)
	assert.Equal(t, "Author One", *items[0].Author)
	assert.Equal(t, "Second Post", items[1].Title)
	assert.Equal(t, "post-2", items[1].GUID)
}

func TestHTTPFetcher_Fetch_Error(t *testing.T) {
	f := fetcher.NewHTTPFetcher()
	_, _, _, _, err := f.Fetch(context.Background(), "http://invalid.local/feed.xml")
	assert.Error(t, err)
}

func TestHTTPFetcher_Fetch_EmptyFeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>Empty</title><link>https://example.com</link></channel></rss>`))
	}))
	defer srv.Close()

	f := fetcher.NewHTTPFetcher()
	items, title, _, _, err := f.Fetch(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "Empty", title)
	assert.Empty(t, items)
}

func TestHTTPFetcher_Discover_RSS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(testRSS))
	}))
	defer srv.Close()

	f := fetcher.NewHTTPFetcher()
	feeds, err := f.Discover(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Len(t, feeds, 1)
	assert.Equal(t, srv.URL, feeds[0].URL)
	assert.Equal(t, "Test Blog", feeds[0].Title)
}

func TestHTTPFetcher_Discover_HTML(t *testing.T) {
	htmlPage := `<html><head>
<link rel="alternate" type="application/rss+xml" title="Blog Feed" href="/feed/" />
<link rel="alternate" type="application/rss+xml" title="Comments Feed" href="https://example.com/comments/feed/" />
<link rel="alternate" type="application/json+oembed" href="/oembed" />
</head><body>hello</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlPage))
	}))
	defer srv.Close()

	f := fetcher.NewHTTPFetcher()
	feeds, err := f.Discover(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Len(t, feeds, 2)
	assert.Equal(t, srv.URL+"/feed/", feeds[0].URL)
	assert.Equal(t, "Blog Feed", feeds[0].Title)
	assert.Equal(t, "https://example.com/comments/feed/", feeds[1].URL)
	assert.Equal(t, "Comments Feed", feeds[1].Title)
}

func TestHTTPFetcher_Discover_NoFeeds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>no feeds here</body></html>`))
	}))
	defer srv.Close()

	f := fetcher.NewHTTPFetcher()
	feeds, err := f.Discover(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Empty(t, feeds)
}

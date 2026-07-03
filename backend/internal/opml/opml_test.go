package opml

import (
	"strings"
	"testing"
	"time"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOPML_Valid(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <head><title>Feeds</title></head>
  <body>
    <outline text="Feed One" type="rss" xmlUrl="https://a.com/feed"/>
    <outline text="Feed Two" type="rss" xmlUrl="https://b.com/rss"/>
  </body>
</opml>`
	feeds, err := ParseOPML(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, feeds, 2)
	assert.Equal(t, "Feed One", feeds[0].Title)
	assert.Equal(t, "https://a.com/feed", feeds[0].URL)
	assert.Equal(t, "Feed Two", feeds[1].Title)
	assert.Equal(t, "https://b.com/rss", feeds[1].URL)
}

func TestParseOPML_Empty(t *testing.T) {
	input := `<?xml version="1.0"?>
<opml version="1.0">
  <head><title>Empty</title></head>
  <body>
  </body>
</opml>`
	feeds, err := ParseOPML(strings.NewReader(input))
	require.NoError(t, err)
	assert.Empty(t, feeds)
}

func TestParseOPML_NoXmlUrl(t *testing.T) {
	input := `<?xml version="1.0"?>
<opml version="1.0">
  <body>
    <outline text="Folder">
      <outline text="Child Feed" type="rss" xmlUrl="https://c.com/feed"/>
    </outline>
    <outline text="No URL Node"/>
  </body>
</opml>`
	feeds, err := ParseOPML(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, feeds, 1)
	assert.Equal(t, "Child Feed", feeds[0].Title)
	assert.Equal(t, "https://c.com/feed", feeds[0].URL)
}

func TestParseOPML_InvalidXML(t *testing.T) {
	_, err := ParseOPML(strings.NewReader("not xml"))
	assert.Error(t, err)
}

func TestParseOPML_NestedOutlines(t *testing.T) {
	input := `<?xml version="1.0"?>
<opml version="2.0">
  <body>
    <outline text="Tech">
      <outline text="Blog A" type="rss" xmlUrl="https://a.com/rss"/>
      <outline text="Blog B" type="rss" xmlUrl="https://b.com/rss"/>
    </outline>
    <outline text="News">
      <outline text="News C" type="rss" xmlUrl="https://c.com/rss"/>
    </outline>
  </body>
</opml>`
	feeds, err := ParseOPML(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, feeds, 3)
	assert.Equal(t, "Blog A", feeds[0].Title)
	assert.Equal(t, "Blog B", feeds[1].Title)
	assert.Equal(t, "News C", feeds[2].Title)
}

func TestParseOPML_TitleFallback(t *testing.T) {
	input := `<?xml version="1.0"?>
<opml version="1.0">
  <body>
    <outline title="Fallback Title" type="rss" xmlUrl="https://d.com/feed"/>
  </body>
</opml>`
	feeds, err := ParseOPML(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, feeds, 1)
	assert.Equal(t, "Fallback Title", feeds[0].Title)
}

func TestParseOPML_XMLEntities(t *testing.T) {
	input := `<?xml version="1.0"?>
<opml version="1.0">
  <body>
    <outline text="Test &amp; Co" type="rss" xmlUrl="https://e.com/feed?x=1&amp;y=2"/>
  </body>
</opml>`
	feeds, err := ParseOPML(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, feeds, 1)
	assert.Equal(t, "Test & Co", feeds[0].Title)
	assert.Equal(t, "https://e.com/feed?x=1&y=2", feeds[0].URL)
}

func TestGenerateOPML_Basic(t *testing.T) {
	now := time.Now()
	feeds := []domain.Feed{
		{ID: 1, UserID: 1, URL: "https://a.com/rss", Title: "Feed A", CreatedAt: now, UpdatedAt: now},
		{ID: 2, UserID: 1, URL: "https://b.com/rss", Title: "Feed B", CreatedAt: now, UpdatedAt: now},
	}
	data, err := GenerateOPML(feeds)
	require.NoError(t, err)

	result, err := ParseOPML(strings.NewReader(string(data)))
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "Feed A", result[0].Title)
	assert.Equal(t, "https://a.com/rss", result[0].URL)
	assert.Equal(t, "Feed B", result[1].Title)
	assert.Equal(t, "https://b.com/rss", result[1].URL)
}

func TestGenerateOPML_Empty(t *testing.T) {
	data, err := GenerateOPML(nil)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<body>")
	assert.Contains(t, string(data), "</body>")
}

func TestGenerateOPML_RoundTrip(t *testing.T) {
	now := time.Now()
	feeds := []domain.Feed{
		{ID: 1, UserID: 1, URL: "https://a.com/rss", Title: "Alpha Feed", CreatedAt: now, UpdatedAt: now},
		{ID: 2, UserID: 1, URL: "https://b.com/feed.xml", Title: "Beta & Co", CreatedAt: now, UpdatedAt: now},
	}
	data, err := GenerateOPML(feeds)
	require.NoError(t, err)

	parsed, err := ParseOPML(strings.NewReader(string(data)))
	require.NoError(t, err)
	require.Len(t, parsed, 2)
	assert.Equal(t, "Alpha Feed", parsed[0].Title)
	assert.Equal(t, "https://a.com/rss", parsed[0].URL)
	assert.Equal(t, "Beta & Co", parsed[1].Title)
	assert.Equal(t, "https://b.com/feed.xml", parsed[1].URL)
}

func TestParseOPML_CaseInsensitiveRoot(t *testing.T) {
	input := `<?xml version="1.0"?>
<OPML version="1.0">
  <body>
    <outline text="Upper" type="rss" xmlUrl="https://u.com/rss"/>
  </body>
</OPML>`
	feeds, err := ParseOPML(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, feeds, 1)
	assert.Equal(t, "Upper", feeds[0].Title)
}

func TestGenerateOPML_XMLEscape(t *testing.T) {
	now := time.Now()
	feeds := []domain.Feed{
		{ID: 1, UserID: 1, URL: "https://x.com/f?q=a&b", Title: `Title "with" <quotes>`, CreatedAt: now, UpdatedAt: now},
	}
	data, err := GenerateOPML(feeds)
	require.NoError(t, err)
	output := string(data)
	assert.NotContains(t, output, `"with"`)
	assert.NotContains(t, output, `<quotes>`)
	assert.Contains(t, output, "&amp;")

	parsed, err := ParseOPML(strings.NewReader(output))
	require.NoError(t, err)
	require.Len(t, parsed, 1)
	assert.Equal(t, `Title "with" <quotes>`, parsed[0].Title)
	assert.Equal(t, "https://x.com/f?q=a&b", parsed[0].URL)
}

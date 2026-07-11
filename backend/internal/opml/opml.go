package opml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ekse/rossoreader/internal/domain"
)

type OpmlFeed struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type opmlDocument struct {
	XMLName xml.Name
	Head    opmlHead `xml:"head"`
	Body    opmlBody `xml:"body"`
}

type opmlHead struct {
	Title string `xml:"title"`
}

type opmlBody struct {
	Outlines []rawOutline `xml:"outline"`
}

type rawOutline struct {
	Text     string       `xml:"text,attr"`
	Title    string       `xml:"title,attr"`
	Type     string       `xml:"type,attr"`
	XMLURL   string       `xml:"xmlUrl,attr"`
	Outlines []rawOutline `xml:"outline"`
}

func ParseOPML(r io.Reader) ([]OpmlFeed, error) {
	var doc opmlDocument
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parse opml: %w", err)
	}

	var feeds []OpmlFeed
	for _, o := range doc.Body.Outlines {
		collectFeeds(o, &feeds)
	}
	if feeds == nil {
		feeds = []OpmlFeed{}
	}
	return feeds, nil
}

func collectFeeds(o rawOutline, feeds *[]OpmlFeed) {
	if o.XMLURL != "" {
		title := o.Text
		if title == "" {
			title = o.Title
		}
		*feeds = append(*feeds, OpmlFeed{Title: title, URL: o.XMLURL})
	}
	for _, child := range o.Outlines {
		collectFeeds(child, feeds)
	}
}

func GenerateOPML(feeds []domain.Feed) ([]byte, error) {
	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<opml version="2.0">`)
	b.WriteByte('\n')

	b.WriteString("  <head>\n")
	b.WriteString("    <title>Feed Subscriptions</title>\n")
	b.WriteString("  </head>\n")

	b.WriteString("  <body>\n")
	for _, f := range feeds {
		title := f.Title
		if title == "" {
			title = f.URL
		}
		b.WriteString(fmt.Sprintf(
			`    <outline text="%s" type="rss" xmlUrl="%s"/>`,
			xmlEscape(title), xmlEscape(f.URL),
		))
		b.WriteByte('\n')
	}
	b.WriteString("  </body>\n")

	b.WriteString("</opml>\n")
	return []byte(b.String()), nil
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

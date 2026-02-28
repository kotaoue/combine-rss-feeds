package feed

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Item is a normalised feed entry shared across RSS and Atom sources.
type Item struct {
	Title       string
	Link        string
	PubDate     time.Time
	Description string
}

// RSS 2.0 input structures

type rssInput struct {
	XMLName xml.Name     `xml:"rss"`
	Channel rssInChannel `xml:"channel"`
}

type rssInChannel struct {
	Title string      `xml:"title"`
	Items []rssInItem `xml:"item"`
}

type rssInItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	Content     string `xml:"encoded"`
}

// Atom input structures

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title   atomText   `xml:"title"`
	Links   []atomLink `xml:"link"`
	Updated string     `xml:"updated"`
	Summary string     `xml:"summary"`
	Content string     `xml:"content"`
}

type atomText struct {
	Value string `xml:",chardata"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// dateFormats lists formats to try when parsing feed dates.
var dateFormats = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	time.RFC3339Nano,
	"Mon, 02 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 2006 15:04:05 MST",
	"2006-01-02T15:04:05Z",
	"2006-01-02",
}

func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, f := range dateFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func hostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Host
}

// Parse parses raw feed bytes (RSS 2.0 or Atom) and returns up to limit Items.
// feedURL is used to derive the source hostname prefix for each entry title.
func Parse(data []byte, feedURL string, limit int) ([]Item, error) {
	type probe struct {
		XMLName xml.Name
	}
	var p probe
	if err := xml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("xml probe: %w", err)
	}

	host := hostname(feedURL)

	switch strings.ToLower(p.XMLName.Local) {
	case "feed":
		return parseAtom(data, host, limit)
	case "rss":
		return parseRSS(data, host, limit)
	default:
		return nil, fmt.Errorf("unknown feed root element: %s", p.XMLName.Local)
	}
}

func parseAtom(data []byte, host string, limit int) ([]Item, error) {
	var f atomFeed
	if err := xml.Unmarshal(data, &f); err != nil {
		return nil, err
	}

	var items []Item
	for i, e := range f.Entries {
		if limit > 0 && i >= limit {
			break
		}
		link := ""
		for _, l := range e.Links {
			if l.Rel == "alternate" || l.Rel == "" {
				link = l.Href
				break
			}
		}
		desc := e.Summary
		if desc == "" {
			desc = e.Content
		}
		items = append(items, Item{
			Title:       fmt.Sprintf("[%s] %s", host, e.Title.Value),
			Link:        link,
			PubDate:     parseDate(e.Updated),
			Description: desc,
		})
	}
	return items, nil
}

func parseRSS(data []byte, host string, limit int) ([]Item, error) {
	var f rssInput
	if err := xml.Unmarshal(data, &f); err != nil {
		return nil, err
	}

	var items []Item
	for i, it := range f.Channel.Items {
		if limit > 0 && i >= limit {
			break
		}
		desc := it.Description
		if desc == "" {
			desc = it.Content
		}
		items = append(items, Item{
			Title:       fmt.Sprintf("[%s] %s", host, it.Title),
			Link:        it.Link,
			PubDate:     parseDate(it.PubDate),
			Description: desc,
		})
	}
	return items, nil
}

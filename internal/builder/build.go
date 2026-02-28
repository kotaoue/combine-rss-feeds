package builder

import (
	"encoding/xml"
	"sort"
	"time"

	"github.com/kotaoue/combine-rss-feeds/internal/parser"
)

// RSS 2.0 output structures

type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

// RSS builds an RSSFeed from parsed items using the given title and description.
func RSS(title, description string, items []parser.Item) RSSFeed {
	rssItems := make([]RSSItem, 0, len(items))
	for _, it := range items {
		pubDate := ""
		if !it.PubDate.IsZero() {
			pubDate = it.PubDate.UTC().Format(time.RFC1123Z)
		}
		rssItems = append(rssItems, RSSItem{
			Title:       it.Title,
			Link:        it.Link,
			PubDate:     pubDate,
			Description: it.Description,
		})
	}
	return RSSFeed{
		Version: "2.0",
		Channel: RSSChannel{
			Title:       title,
			Link:        "",
			Description: description,
			Items:       rssItems,
		},
	}
}

// SortItems sorts items by PubDate descending (newest first).
func SortItems(items []parser.Item) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].PubDate.After(items[j].PubDate)
	})
}

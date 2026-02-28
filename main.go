package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
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

// RSS 2.0 input parse structures

type rssInput struct {
	XMLName xml.Name      `xml:"rss"`
	Channel rssInChannel  `xml:"channel"`
}

type rssInChannel struct {
	Title string       `xml:"title"`
	Items []rssInItem  `xml:"item"`
}

type rssInItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	Content     string `xml:"encoded"`
}

// Atom input parse structures

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

// parsedItem is an internal normalised representation

type parsedItem struct {
	Title       string
	Link        string
	PubDate     time.Time
	Description string
}

// date formats to try when parsing feed dates
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

func fetchFeed(feedURL string, limit int) ([]parsedItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request %s: %w", feedURL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", feedURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body %s: %w", feedURL, err)
	}

	items, err := parseFeed(body, feedURL, limit)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", feedURL, err)
	}
	return items, nil
}

func parseFeed(data []byte, feedURL string, limit int) ([]parsedItem, error) {
	// Detect feed type by looking for root element name
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

func parseAtom(data []byte, host string, limit int) ([]parsedItem, error) {
	var feed atomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	var items []parsedItem
	for i, e := range feed.Entries {
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
		title := fmt.Sprintf("[%s] %s", host, e.Title.Value)
		items = append(items, parsedItem{
			Title:       title,
			Link:        link,
			PubDate:     parseDate(e.Updated),
			Description: desc,
		})
	}
	return items, nil
}

func parseRSS(data []byte, host string, limit int) ([]parsedItem, error) {
	var feed rssInput
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, err
	}

	var items []parsedItem
	for i, it := range feed.Channel.Items {
		if limit > 0 && i >= limit {
			break
		}
		desc := it.Description
		if desc == "" {
			desc = it.Content
		}
		title := fmt.Sprintf("[%s] %s", host, it.Title)
		items = append(items, parsedItem{
			Title:       title,
			Link:        it.Link,
			PubDate:     parseDate(it.PubDate),
			Description: desc,
		})
	}
	return items, nil
}

func buildRSS(title, description string, items []parsedItem) RSSFeed {
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

func sortItems(items []parsedItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].PubDate.After(items[j].PubDate)
	})
}

func main() {
	feedsRaw := os.Getenv("INPUT_FEEDS")
	outputFile := os.Getenv("INPUT_OUTPUT_FILE")
	limitStr := os.Getenv("INPUT_LIMIT")
	feedTitle := os.Getenv("INPUT_FEED_TITLE")
	feedDesc := os.Getenv("INPUT_FEED_DESCRIPTION")

	if outputFile == "" {
		outputFile = "combined_feed.xml"
	}
	if feedTitle == "" {
		feedTitle = "Combined RSS Feed"
	}
	if feedDesc == "" {
		feedDesc = "Merged feed generated by combine-rss-feeds action"
	}

	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(strings.TrimSpace(limitStr)); err == nil && v > 0 {
			limit = v
		}
	}

	var feedURLs []string
	for _, line := range strings.Split(feedsRaw, "\n") {
		u := strings.TrimSpace(line)
		if u != "" {
			feedURLs = append(feedURLs, u)
		}
	}

	if len(feedURLs) == 0 {
		fmt.Fprintln(os.Stderr, "No feed URLs provided in INPUT_FEEDS")
		os.Exit(1)
	}

	var allItems []parsedItem
	for _, u := range feedURLs {
		items, err := fetchFeed(u, limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}
		allItems = append(allItems, items...)
	}

	// Sort by pubDate descending
	sortItems(allItems)

	feed := buildRSS(feedTitle, feedDesc, allItems)

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal XML: %v\n", err)
		os.Exit(1)
	}

	content := xml.Header + string(out) + "\n"

	// Ensure parent directories exist
	dir := filepath.Dir(outputFile)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", outputFile, err)
		os.Exit(1)
	}

	fmt.Printf("Written %d items to %s\n", len(allItems), outputFile)
}

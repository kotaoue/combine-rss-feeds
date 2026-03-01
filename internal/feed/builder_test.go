package feed

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestRSS(t *testing.T) {
	items := []Item{
		{Title: "T1", Link: "https://a.com/1", PubDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Description: "D1"},
		{Title: "T2", Link: "https://a.com/2", PubDate: time.Time{}, Description: "D2"},
	}
	feed := RSS("My Feed", "My Desc", items)
	if feed.Version != "2.0" {
		t.Errorf("expected version 2.0, got %q", feed.Version)
	}
	if feed.Channel.Title != "My Feed" {
		t.Errorf("unexpected channel title: %q", feed.Channel.Title)
	}
	if len(feed.Channel.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(feed.Channel.Items))
	}
	if feed.Channel.Items[1].PubDate != "" {
		t.Errorf("zero time should produce empty pubDate, got %q", feed.Channel.Items[1].PubDate)
	}
}

func TestRSSXMLMarshal(t *testing.T) {
	items := []Item{
		{Title: "T1", Link: "https://a.com/1", PubDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Description: "D1"},
	}
	f := RSS("My Feed", "My Desc", items)
	out, err := xml.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatalf("xml.MarshalIndent: %v", err)
	}
	xmlStr := xml.Header + string(out)
	if !strings.Contains(xmlStr, "<rss") {
		t.Error("output missing <rss")
	}
	if !strings.Contains(xmlStr, "<item>") {
		t.Error("output missing <item>")
	}
}

func TestSortItemsAndLimit(t *testing.T) {
	// Simulate combining items from multiple feeds then applying a global limit.
	items := []Item{
		{Title: "A1", PubDate: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{Title: "A2", PubDate: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		{Title: "B1", PubDate: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)},
		{Title: "B2", PubDate: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Title: "C1", PubDate: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)},
	}

	SortItems(items)

	limit := 3
	if len(items) > limit {
		items = items[:limit]
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items after limit, got %d", len(items))
	}
	if items[0].Title != "B1" {
		t.Errorf("expected B1 first (newest), got %q", items[0].Title)
	}
	if items[1].Title != "A1" {
		t.Errorf("expected A1 second, got %q", items[1].Title)
	}
	if items[2].Title != "C1" {
		t.Errorf("expected C1 third, got %q", items[2].Title)
	}
}

func TestSortItems(t *testing.T) {
	items := []Item{
		{Title: "Old", PubDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Newest", PubDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Middle", PubDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	SortItems(items)

	if items[0].Title != "Newest" {
		t.Errorf("expected Newest first, got %q", items[0].Title)
	}
	if items[1].Title != "Middle" {
		t.Errorf("expected Middle second, got %q", items[1].Title)
	}
	if items[2].Title != "Old" {
		t.Errorf("expected Old last, got %q", items[2].Title)
	}
}

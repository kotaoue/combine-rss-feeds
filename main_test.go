package main

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/kotaoue/combine-rss-feeds/internal/parse"
)

func TestBuildRSS(t *testing.T) {
	items := []parse.Item{
		{Title: "T1", Link: "https://a.com/1", PubDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Description: "D1"},
		{Title: "T2", Link: "https://a.com/2", PubDate: time.Time{}, Description: "D2"},
	}
	feed := buildRSS("My Feed", "My Desc", items)
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

func TestBuildRSSXMLMarshal(t *testing.T) {
	items := []parse.Item{
		{Title: "T1", Link: "https://a.com/1", PubDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Description: "D1"},
	}
	feed := buildRSS("My Feed", "My Desc", items)
	out, err := xml.MarshalIndent(feed, "", "  ")
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

func TestSortOrder(t *testing.T) {
	items := []parse.Item{
		{Title: "Old", PubDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Newest", PubDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Middle", PubDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	sortItems(items)

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


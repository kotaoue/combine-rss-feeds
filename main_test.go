package main

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	cases := []struct {
		input    string
		wantZero bool
	}{
		{"Mon, 02 Jan 2006 15:04:05 +0000", false},
		{"2006-01-02T15:04:05Z", false},
		{"2006-01-02T15:04:05+09:00", false},
		{"", true},
		{"not-a-date", true},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got := parseDate(c.input)
			if c.wantZero && !got.IsZero() {
				t.Errorf("parseDate(%q) = %v, want zero time", c.input, got)
			}
			if !c.wantZero && got.IsZero() {
				t.Errorf("parseDate(%q) = zero, want non-zero", c.input)
			}
		})
	}
}

func TestHostname(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"https://zenn.dev/kotaoue/feed", "zenn.dev"},
		{"https://qiita.com/kotaoue/feed", "qiita.com"},
		{"not-a-url", "not-a-url"},
	}
	for _, c := range cases {
		got := hostname(c.input)
		if got != c.want {
			t.Errorf("hostname(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

var sampleRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>Article 1</title>
      <link>https://example.com/1</link>
      <pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
      <description>Desc 1</description>
    </item>
    <item>
      <title>Article 2</title>
      <link>https://example.com/2</link>
      <pubDate>Tue, 02 Jan 2024 00:00:00 +0000</pubDate>
      <description>Desc 2</description>
    </item>
    <item>
      <title>Article 3</title>
      <link>https://example.com/3</link>
      <pubDate>Wed, 03 Jan 2024 00:00:00 +0000</pubDate>
      <description>Desc 3</description>
    </item>
  </channel>
</rss>`

var sampleAtom = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Atom Feed</title>
  <entry>
    <title>Atom Entry 1</title>
    <link href="https://atom.example.com/1" rel="alternate"/>
    <updated>2024-03-01T10:00:00Z</updated>
    <summary>Atom Desc 1</summary>
  </entry>
  <entry>
    <title>Atom Entry 2</title>
    <link href="https://atom.example.com/2"/>
    <updated>2024-03-02T10:00:00Z</updated>
    <summary>Atom Desc 2</summary>
  </entry>
</feed>`

func TestParseRSS(t *testing.T) {
	items, err := parseFeed([]byte(sampleRSS), "https://example.com/feed", 10)
	if err != nil {
		t.Fatalf("parseFeed RSS: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].Title != "[example.com] Article 1" {
		t.Errorf("unexpected title: %q", items[0].Title)
	}
	if items[0].Link != "https://example.com/1" {
		t.Errorf("unexpected link: %q", items[0].Link)
	}
}

func TestParseRSSLimit(t *testing.T) {
	items, err := parseFeed([]byte(sampleRSS), "https://example.com/feed", 2)
	if err != nil {
		t.Fatalf("parseFeed RSS limit: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (limit), got %d", len(items))
	}
}

func TestParseAtom(t *testing.T) {
	items, err := parseFeed([]byte(sampleAtom), "https://atom.example.com/feed", 10)
	if err != nil {
		t.Fatalf("parseFeed Atom: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "[atom.example.com] Atom Entry 1" {
		t.Errorf("unexpected title: %q", items[0].Title)
	}
	if items[0].Link != "https://atom.example.com/1" {
		t.Errorf("unexpected link: %q", items[0].Link)
	}
}

func TestBuildRSS(t *testing.T) {
	items := []parsedItem{
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
	items := []parsedItem{
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
	items := []parsedItem{
		{Title: "Old", PubDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Newest", PubDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Middle", PubDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	// replicate the sort done in main()
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

func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

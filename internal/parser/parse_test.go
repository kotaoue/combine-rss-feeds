package parser

import (
	"testing"
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

func TestFeedRSS(t *testing.T) {
	items, err := Feed([]byte(sampleRSS), "https://example.com/feed", 10)
	if err != nil {
		t.Fatalf("Feed RSS: %v", err)
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

func TestFeedRSSLimit(t *testing.T) {
	items, err := Feed([]byte(sampleRSS), "https://example.com/feed", 2)
	if err != nil {
		t.Fatalf("Feed RSS limit: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items (limit), got %d", len(items))
	}
}

func TestFeedAtom(t *testing.T) {
	items, err := Feed([]byte(sampleAtom), "https://atom.example.com/feed", 10)
	if err != nil {
		t.Fatalf("Feed Atom: %v", err)
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

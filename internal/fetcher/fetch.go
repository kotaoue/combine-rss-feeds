package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kotaoue/combine-rss-feeds/internal/parser"
)

// Feed fetches feedURL and returns up to limit parsed items.
func Feed(feedURL string, limit int) ([]parser.Item, error) {
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

	items, err := parser.Feed(body, feedURL, limit)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", feedURL, err)
	}
	return items, nil
}

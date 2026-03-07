package websearch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WebSearchResult represents a single search result.
type WebSearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// DuckDuckGoSearch queries the DuckDuckGo Instant Answer API.
// No API key required. Returns a formatted summary of results.
func DuckDuckGoSearch(query string) string {
	encoded := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1", encoded)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err)
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Sprintf("Search failed: invalid response")
	}

	// Try abstract first (Wikipedia-style direct answer)
	if abstract, ok := data["AbstractText"].(string); ok && abstract != "" {
		source, _ := data["AbstractSource"].(string)
		srcURL, _ := data["AbstractURL"].(string)
		result := fmt.Sprintf("From %s: %s", source, abstract)
		if srcURL != "" {
			result += fmt.Sprintf("\nSource: %s", srcURL)
		}
		return result
	}

	// Try related topics
	if topics, ok := data["RelatedTopics"].([]any); ok && len(topics) > 0 {
		var lines []string
		count := 0
		for _, topic := range topics {
			if count >= 5 {
				break
			}
			if t, ok := topic.(map[string]any); ok {
				if text, ok := t["Text"].(string); ok && text != "" {
					lines = append(lines, "- "+text)
					count++
				}
			}
		}
		if len(lines) > 0 {
			return strings.Join(lines, "\n")
		}
	}

	return fmt.Sprintf("No results found for: %s", query)
}

// WebFetch fetches a URL and returns the response body, truncated to maxLen.
// If maxLen is 0, defaults to 50000.
func WebFetch(rawURL string, maxLen int) (string, error) {
	if maxLen <= 0 {
		maxLen = 50000
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", rawURL, err)
	}

	content := string(body)
	if len(content) > maxLen {
		content = content[:maxLen] + fmt.Sprintf("\n\n... (truncated, %d total chars)", len(body))
	}
	return content, nil
}

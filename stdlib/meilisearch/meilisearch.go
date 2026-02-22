package meilisearch

import (
	"fmt"
	"sort"
	"strings"

	haira "haira-go-runtime/haira"
)

// MeilisearchClient is an authenticated Meilisearch API client.
type MeilisearchClient struct {
	client *haira.HTTPClient
}

// MeilisearchNewClient creates a Meilisearch client with a base URL and API key.
func MeilisearchNewClient(url, apiKey string) *MeilisearchClient {
	headers := map[string]any{
		"Content-Type": "application/json",
	}
	if apiKey != "" {
		headers["Authorization"] = "Bearer " + apiKey
	}
	return &MeilisearchClient{
		client: haira.HttpClient(url, map[string]any{
			"headers": headers,
			"retry":   2,
			"timeout": 10000,
		}),
	}
}

// Search performs a search query on a Meilisearch index.
// opts: limit (int), offset (int), filter (string or []string),
//
//	facets ([]string), sort ([]string), attributesToRetrieve ([]string).
func (c *MeilisearchClient) Search(index, query string, opts map[string]any) (map[string]any, error) {
	body := map[string]any{"q": query}
	if opts != nil {
		for k, v := range opts {
			body[k] = v
		}
	}

	resp, err := c.client.Post(fmt.Sprintf("/indexes/%s/search", index), body)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("meilisearch search: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return haira.JSONUnmarshal(resp.Body)
}

// MultiSearch performs a federated search across multiple indices.
func (c *MeilisearchClient) MultiSearch(queries []any) ([]any, error) {
	body := map[string]any{"queries": queries}
	resp, err := c.client.Post("/multi-search", body)
	if err != nil {
		return nil, fmt.Errorf("meilisearch multi_search: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("meilisearch multi_search: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	data := haira.ParseJSON(resp.Body)
	if results, ok := data["results"].([]any); ok {
		return results, nil
	}
	return nil, fmt.Errorf("meilisearch multi_search: unexpected response format")
}

// Get retrieves a single document by its ID from an index.
func (c *MeilisearchClient) Get(index, documentID string) (map[string]any, error) {
	resp, err := c.client.Get(fmt.Sprintf("/indexes/%s/documents/%s", index, documentID))
	if err != nil {
		return nil, fmt.Errorf("meilisearch get: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("meilisearch get: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return haira.JSONUnmarshal(resp.Body)
}

// Stats returns global stats for the Meilisearch instance.
func (c *MeilisearchClient) Stats() (map[string]any, error) {
	resp, err := c.client.Get("/stats")
	if err != nil {
		return nil, fmt.Errorf("meilisearch stats: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("meilisearch stats: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return haira.JSONUnmarshal(resp.Body)
}

// ListIndexes returns the list of all indexes.
func (c *MeilisearchClient) ListIndexes() ([]any, error) {
	resp, err := c.client.Get("/indexes")
	if err != nil {
		return nil, fmt.Errorf("meilisearch list_indexes: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("meilisearch list_indexes: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	data := haira.ParseJSON(resp.Body)
	if results, ok := data["results"].([]any); ok {
		return results, nil
	}
	return nil, nil
}

// Health checks the Meilisearch instance health.
func (c *MeilisearchClient) Health() (map[string]any, error) {
	resp, err := c.client.Get("/health")
	if err != nil {
		return nil, fmt.Errorf("meilisearch health: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("meilisearch health: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return haira.JSONUnmarshal(resp.Body)
}

// Hits extracts the hits array from a search result map.
func MeilisearchHits(result map[string]any) []any {
	if hits, ok := result["hits"].([]any); ok {
		return hits
	}
	return nil
}

// HitsToTable converts Meilisearch hits into a headers + rows pair for table display.
// It selects only human-readable scalar fields, prioritises common fields,
// and caps columns at 8 for readability.
func MeilisearchHitsToTable(hits []any) ([]any, []any) {
	if len(hits) == 0 {
		return nil, nil
	}

	headers := pickDisplayHeaders(hits[0], 8)
	if len(headers) == 0 {
		return nil, nil
	}

	rows := make([]any, 0, len(hits))
	for _, hit := range hits {
		if m, ok := hit.(map[string]any); ok {
			row := make([]any, 0, len(headers))
			for _, h := range headers {
				row = append(row, formatCellValue(m[haira.Str(h)]))
			}
			rows = append(rows, row)
		}
	}
	return headers, rows
}

// priorityFields are field names commonly found in search records, ordered
// by display priority (most useful first).
var priorityFields = []string{
	"id", "objectID", "name", "title", "label",
	"brand", "category", "type", "status",
	"price", "amount", "quantity", "rating", "score",
	"description", "summary", "url", "slug",
	"email", "username", "city", "country",
	"created_at", "createdAt", "updated_at", "updatedAt", "date",
}

// pickDisplayHeaders selects up to maxCols display-worthy fields from a hit.
func pickDisplayHeaders(hit any, maxCols int) []any {
	m, ok := hit.(map[string]any)
	if !ok {
		return nil
	}

	scalarKeys := make(map[string]bool)
	for k, v := range m {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if isScalar(v) {
			scalarKeys[k] = true
		}
	}

	var headers []any
	seen := make(map[string]bool)
	for _, pf := range priorityFields {
		if scalarKeys[pf] && !seen[pf] {
			headers = append(headers, pf)
			seen[pf] = true
			if len(headers) >= maxCols {
				return headers
			}
		}
	}

	remaining := make([]string, 0)
	for k := range scalarKeys {
		if !seen[k] {
			remaining = append(remaining, k)
		}
	}
	sort.Strings(remaining)
	for _, k := range remaining {
		headers = append(headers, k)
		if len(headers) >= maxCols {
			break
		}
	}

	return headers
}

func isScalar(v any) bool {
	if v == nil {
		return true
	}
	switch v.(type) {
	case map[string]any, []any:
		return false
	default:
		return true
	}
}

// MeilisearchHitsToProductCards converts Meilisearch hits into product card items.
// Looks for common e-commerce fields: name/title, price, image/thumbnailUrl, brand, description, url/slug.
func MeilisearchHitsToProductCards(hits []any) []any {
	cards := make([]any, 0, len(hits))
	for _, hit := range hits {
		m, ok := hit.(map[string]any)
		if !ok {
			continue
		}
		card := haira.UiProductCardItem{
			Name:        firstString(m, "name", "title", "label", "productName"),
			Price:       firstString(m, "price", "formattedPrice", "salePrice", "amount"),
			Image:       firstString(m, "image", "thumbnailUrl", "thumbnail", "imageUrl", "image_url", "picture", "photo"),
			Brand:       firstString(m, "brand", "manufacturer", "vendor"),
			Description: truncate(firstString(m, "description", "shortDescription", "short_description", "summary"), 120),
			Badge:       firstString(m, "badge", "tag", "label"),
			URL:         firstString(m, "url", "slug", "link", "permalink"),
		}
		if card.Name == "" {
			card.Name = firstString(m, "id", "objectID")
		}
		cards = append(cards, card)
	}
	return cards
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			s := haira.Str(v)
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func formatCellValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		if len(val) > 120 {
			return val[:117] + "..."
		}
		return val
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		summary := strings.Join(keys, ", ")
		if len(summary) > 80 {
			summary = summary[:77] + "..."
		}
		return "{" + summary + "}"
	case []any:
		if len(val) == 0 {
			return ""
		}
		parts := make([]string, 0, 3)
		for i, item := range val {
			if i >= 3 {
				parts = append(parts, fmt.Sprintf("...+%d more", len(val)-3))
				break
			}
			parts = append(parts, haira.Str(item))
		}
		return strings.Join(parts, ", ")
	default:
		return haira.Str(v)
	}
}

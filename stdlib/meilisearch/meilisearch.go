package meilisearch

import (
	"fmt"

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

	headers := haira.PickDisplayHeaders(hits[0], 8)
	if len(headers) == 0 {
		return nil, nil
	}

	rows := make([]any, 0, len(hits))
	for _, hit := range hits {
		if m, ok := hit.(map[string]any); ok {
			row := make([]any, 0, len(headers))
			for _, h := range headers {
				row = append(row, haira.FormatCellValue(m[haira.Str(h)]))
			}
			rows = append(rows, row)
		}
	}
	return headers, rows
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
			Name:        haira.FirstString(m, "name", "title", "label", "productName"),
			Price:       haira.FirstString(m, "price", "formattedPrice", "salePrice", "amount"),
			Image:       haira.FirstString(m, "image", "thumbnailUrl", "thumbnail", "imageUrl", "image_url", "picture", "photo"),
			Brand:       haira.FirstString(m, "brand", "manufacturer", "vendor"),
			Description: haira.Truncate(haira.FirstString(m, "description", "shortDescription", "short_description", "summary"), 120),
			Badge:       haira.FirstString(m, "badge", "tag", "label"),
			URL:         haira.FirstString(m, "url", "slug", "link", "permalink"),
		}
		if card.Name == "" {
			card.Name = haira.FirstString(m, "id", "objectID")
		}
		cards = append(cards, card)
	}
	return cards
}

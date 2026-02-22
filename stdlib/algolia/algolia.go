package algolia

import (
	"fmt"
	"sort"
	"strings"

	haira "haira-go-runtime/haira"
)

// AlgoliaClient is an authenticated Algolia API client.
type AlgoliaClient struct {
	client *haira.HTTPClient
	appID  string
}

// AlgoliaNewClient creates an Algolia client with an application ID and API key.
func AlgoliaNewClient(appID, apiKey string) *AlgoliaClient {
	baseURL := fmt.Sprintf("https://%s-dsn.algolia.net", appID)
	return &AlgoliaClient{
		appID: appID,
		client: haira.HttpClient(baseURL, map[string]any{
			"headers": map[string]any{
				"X-Algolia-API-Key":        apiKey,
				"X-Algolia-Application-ID": appID,
				"Content-Type":             "application/json",
			},
			"retry":   2,
			"timeout": 10000,
		}),
	}
}

// Search performs a search query on an Algolia index.
// opts: hitsPerPage (int), page (int), filters (string), facets ([]string),
//
//	attributesToRetrieve ([]string), attributesToHighlight ([]string).
func (c *AlgoliaClient) Search(index, query string, opts map[string]any) (map[string]any, error) {
	body := map[string]any{"query": query}
	if opts != nil {
		for k, v := range opts {
			body[camelCase(k)] = v
		}
	}

	resp, err := c.client.Post(fmt.Sprintf("/1/indexes/%s/query", index), body)
	if err != nil {
		return nil, fmt.Errorf("algolia search: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("algolia search: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return haira.JSONUnmarshal(resp.Body)
}

// MultiSearch performs a search across multiple indices.
func (c *AlgoliaClient) MultiSearch(requests []any) ([]any, error) {
	body := map[string]any{"requests": requests}
	resp, err := c.client.Post("/1/indexes/*/queries", body)
	if err != nil {
		return nil, fmt.Errorf("algolia multi_search: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("algolia multi_search: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	data := haira.ParseJSON(resp.Body)
	if results, ok := data["results"].([]any); ok {
		return results, nil
	}
	return nil, fmt.Errorf("algolia multi_search: unexpected response format")
}

// Get retrieves a single object by its objectID from an index.
func (c *AlgoliaClient) Get(index, objectID string) (map[string]any, error) {
	resp, err := c.client.Get(fmt.Sprintf("/1/indexes/%s/%s", index, objectID))
	if err != nil {
		return nil, fmt.Errorf("algolia get: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("algolia get: HTTP %d: %s", resp.StatusCode, resp.Body)
	}
	return haira.JSONUnmarshal(resp.Body)
}

// Browse retrieves all objects from an index using the browse endpoint.
// opts: hitsPerPage (int), filters (string), cursor (string).
func (c *AlgoliaClient) Browse(index string, opts map[string]any) ([]any, error) {
	body := map[string]any{}
	if opts != nil {
		for k, v := range opts {
			body[camelCase(k)] = v
		}
	}

	resp, err := c.client.Post(fmt.Sprintf("/1/indexes/%s/browse", index), body)
	if err != nil {
		return nil, fmt.Errorf("algolia browse: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("algolia browse: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	data := haira.ParseJSON(resp.Body)
	if hits, ok := data["hits"].([]any); ok {
		return hits, nil
	}
	return nil, nil
}

// ListIndices returns the list of all indices for the application.
func (c *AlgoliaClient) ListIndices() ([]any, error) {
	resp, err := c.client.Get("/1/indexes")
	if err != nil {
		return nil, fmt.Errorf("algolia list_indices: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("algolia list_indices: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	data := haira.ParseJSON(resp.Body)
	if items, ok := data["items"].([]any); ok {
		return items, nil
	}
	return nil, nil
}

// Facets retrieves facet values for a given facet name on an index.
func (c *AlgoliaClient) Facets(index, facet string, opts map[string]any) ([]any, error) {
	body := map[string]any{}
	if opts != nil {
		for k, v := range opts {
			body[camelCase(k)] = v
		}
	}

	resp, err := c.client.Post(fmt.Sprintf("/1/indexes/%s/facets/%s/query", index, facet), body)
	if err != nil {
		return nil, fmt.Errorf("algolia facets: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("algolia facets: HTTP %d: %s", resp.StatusCode, resp.Body)
	}

	data := haira.ParseJSON(resp.Body)
	if values, ok := data["facetHits"].([]any); ok {
		return values, nil
	}
	return nil, nil
}

// Hits extracts the hits array from a search result map.
func AlgoliaHits(result map[string]any) []any {
	if hits, ok := result["hits"].([]any); ok {
		return hits
	}
	return nil
}

// HitsToTable converts Algolia hits into a headers + rows pair for table display.
// It selects only human-readable scalar fields, prioritises common fields
// (objectID, name, title, brand, price, etc.), and caps columns at maxCols.
func AlgoliaHitsToTable(hits []any) ([]any, []any) {
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
	"objectID", "id", "name", "title", "label",
	"brand", "category", "type", "status",
	"price", "amount", "quantity", "rating", "score",
	"description", "summary", "url", "slug",
	"email", "username", "city", "country",
	"created_at", "createdAt", "updated_at", "updatedAt", "date",
}

// pickDisplayHeaders selects up to maxCols display-worthy fields from a hit.
// It skips internal fields (prefixed with _), nested objects, and arrays.
// Priority fields come first, then remaining scalar fields alphabetically.
func pickDisplayHeaders(hit any, maxCols int) []any {
	m, ok := hit.(map[string]any)
	if !ok {
		return nil
	}

	// Partition keys into scalar vs complex
	scalarKeys := make(map[string]bool)
	for k, v := range m {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if isScalar(v) {
			scalarKeys[k] = true
		}
	}

	// Pick priority fields first (in order)
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

	// Fill remaining slots with other scalar fields (sorted for stability)
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

// isScalar returns true for values that display well in a table cell
// (strings, numbers, booleans) and false for maps and arrays.
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
		// Summarise nested objects as "key1, key2, ..." instead of raw JSON
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
		// Show first few items as comma-separated
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

// AlgoliaHitsToProductCards converts Algolia hits into product card items.
// Looks for common e-commerce fields: name/title, price, image/thumbnailUrl, brand, description, url/slug.
func AlgoliaHitsToProductCards(hits []any) []any {
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
			URL:         firstString(m, "url", "slug", "canonicalSlug", "link", "permalink"),
		}
		// Fallback: use objectID as name if no name field found
		if card.Name == "" {
			card.Name = firstString(m, "objectID", "id")
		}
		cards = append(cards, card)
	}
	return cards
}

// firstString returns the first non-empty string value from the map for the given keys.
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

// camelCase converts snake_case keys to camelCase for Algolia API compatibility.
func camelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

package algolia

import (
	"fmt"
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
			Name:        haira.FirstString(m, "name", "title", "label", "productName"),
			Price:       haira.FirstString(m, "price", "formattedPrice", "salePrice", "amount"),
			Image:       haira.FirstString(m, "image", "thumbnailUrl", "thumbnail", "imageUrl", "image_url", "picture", "photo"),
			Brand:       haira.FirstString(m, "brand", "manufacturer", "vendor"),
			Description: haira.Truncate(haira.FirstString(m, "description", "shortDescription", "short_description", "summary"), 120),
			Badge:       haira.FirstString(m, "badge", "tag", "label"),
			URL:         haira.FirstString(m, "url", "slug", "canonicalSlug", "link", "permalink"),
		}
		// Fallback: use objectID as name if no name field found
		if card.Name == "" {
			card.Name = haira.FirstString(m, "objectID", "id")
		}
		cards = append(cards, card)
	}
	return cards
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

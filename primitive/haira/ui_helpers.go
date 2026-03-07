package haira

import (
	"fmt"
	"sort"
	"strings"
)

// SearchPriorityFields are field names commonly found in search records, ordered
// by display priority (most useful first). Used by PickDisplayHeaders.
var SearchPriorityFields = []string{
	"id", "objectID", "name", "title", "label",
	"brand", "category", "type", "status",
	"price", "amount", "quantity", "rating", "score",
	"description", "summary", "url", "slug",
	"email", "username", "city", "country",
	"created_at", "createdAt", "updated_at", "updatedAt", "date",
}

// PickDisplayHeaders selects up to maxCols display-worthy fields from a hit.
// It skips internal fields (prefixed with _), nested objects, and arrays.
// Priority fields come first, then remaining scalar fields alphabetically.
func PickDisplayHeaders(hit any, maxCols int) []any {
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
		if IsScalar(v) {
			scalarKeys[k] = true
		}
	}

	// Pick priority fields first (in order)
	var headers []any
	seen := make(map[string]bool)
	for _, pf := range SearchPriorityFields {
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

// IsScalar returns true for values that display well in a table cell
// (strings, numbers, booleans) and false for maps and arrays.
func IsScalar(v any) bool {
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

// FormatCellValue formats a value for display in a table cell.
// Strings are truncated at 120 chars, maps show sorted keys, arrays show first 3 items.
func FormatCellValue(v any) string {
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
			parts = append(parts, Str(item))
		}
		return strings.Join(parts, ", ")
	default:
		return Str(v)
	}
}

// FirstString returns the first non-empty string value from the map for the given keys.
func FirstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			s := Str(v)
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

// Truncate truncates a string to maxLen characters, appending "..." if truncated.
func Truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

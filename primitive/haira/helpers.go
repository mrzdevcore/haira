package haira

import "encoding/json"

// parseJSON unmarshals a JSON string into a map. Returns empty map on error.
func parseJSON(body string) map[string]any {
	var result map[string]any
	json.Unmarshal([]byte(body), &result)
	if result == nil {
		return map[string]any{}
	}
	return result
}

// strVal safely extracts a string value from a map by key.
func strVal(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return Str(v)
	}
	return ""
}

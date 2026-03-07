package haira

import (
	"encoding/json"
	"fmt"
)

// ParseJSON unmarshals a JSON string into a map.
// Panics on parse error (caught by Haira's try/catch via recover).
func ParseJSON(body string) map[string]any {
	var result map[string]any
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		panic(fmt.Sprintf("json parse error: %v", err))
	}
	if result == nil {
		return map[string]any{}
	}
	return result
}

// ErrorMessage safely extracts the error message from a value.
// If the value implements the error interface, returns .Error().
// Otherwise, returns the fmt.Sprint representation.
func ErrorMessage(v any) string {
	if v == nil {
		return ""
	}
	if err, ok := v.(error); ok {
		return err.Error()
	}
	return fmt.Sprint(v)
}

// StrVal safely extracts a string value from a map by key.
func StrVal(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return Str(v)
	}
	return ""
}

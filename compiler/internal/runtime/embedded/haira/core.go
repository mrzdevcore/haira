// Package haira provides the minimal Haira standard library runtime.
// This is the embedded version with zero external dependencies.
package haira

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Response wraps an HTTP response for Haira programs.
type Response struct {
	StatusCode int
	Body       string
}

// Json parses the response body as JSON and returns the result.
func (r *Response) Json() map[string]any {
	var result map[string]any
	json.Unmarshal([]byte(r.Body), &result)
	if result == nil {
		result = map[string]any{}
	}
	return result
}

// JsonArray parses the response body as a JSON array and returns the result.
func (r *Response) JsonArray() []any {
	var result []any
	json.Unmarshal([]byte(r.Body), &result)
	return result
}

// Get performs dynamic indexing on maps and slices.
func Get(obj any, key any) any {
	switch o := obj.(type) {
	case map[string]any:
		if k, ok := key.(string); ok {
			return o[k]
		}
	case []any:
		idx := toInt(key)
		if idx >= 0 && idx < len(o) {
			return o[idx]
		}
	case []map[string]any:
		idx := toInt(key)
		if idx >= 0 && idx < len(o) {
			return o[idx]
		}
	case []string:
		idx := toInt(key)
		if idx >= 0 && idx < len(o) {
			return o[idx]
		}
	}
	return nil
}

func toInt(v any) int {
	switch k := v.(type) {
	case int:
		return k
	case float64:
		return int(k)
	default:
		return -1
	}
}

// toFloat64 converts any numeric value to float64.
func toFloat64(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case float64:
		return n
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	default:
		return 0
	}
}

// Str converts any value to a string.
func Str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// Concat concatenates two values as strings.
func Concat(a, b any) string {
	return Str(a) + Str(b)
}

// Set assigns a value to a map key dynamically.
func Set(obj any, key any, value any) {
	if m, ok := obj.(map[string]any); ok {
		m[Str(key)] = value
	}
}

// Len returns the length of a slice, map, or string.
func Len(v any) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case []any:
		return len(val)
	case []map[string]any:
		return len(val)
	case map[string]any:
		return len(val)
	case string:
		return len(val)
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Map {
			return rv.Len()
		}
		return 0
	}
}

// ToSlice converts an any value to []any for safe iteration.
func ToSlice(v any) []any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []any:
		return val
	case []map[string]any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = item
		}
		return result
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			result := make([]any, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				result[i] = rv.Index(i).Interface()
			}
			return result
		}
		return nil
	}
}

// ToMap converts an any value to map[string]any for safe map iteration.
func ToMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		result := make(map[string]any, rv.Len())
		for _, key := range rv.MapKeys() {
			result[key.String()] = rv.MapIndex(key).Interface()
		}
		return result
	}
	return nil
}

// Keys returns the keys of a map as a slice.
func Keys(obj any) []any {
	if m, ok := obj.(map[string]any); ok {
		keys := make([]any, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}
	return nil
}

// Join concatenates a slice of strings with a separator.
func Join(items any, sep any) string {
	sepStr := Str(sep)
	switch v := items.(type) {
	case []any:
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = Str(item)
		}
		return strings.Join(parts, sepStr)
	case []string:
		return strings.Join(v, sepStr)
	}
	return ""
}

// isTruthy converts a value to a boolean for filter/find/every/some.
func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return true
	}
}

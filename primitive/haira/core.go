package haira

import (
	"fmt"
	"reflect"
	"strings"
)

// Str converts any value to a string.
func Str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// Concat concatenates two values as strings.
// Handles any-typed values by converting to string first.
func Concat(a, b any) string {
	return Str(a) + Str(b)
}

// Get performs dynamic indexing on maps and slices.
// Supports string keys for maps and int/float64 indices for slices.
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

// Set assigns a value to a map key dynamically.
// Handles any-typed keys by converting to string.
func Set(obj any, key any, value any) {
	if m, ok := obj.(map[string]any); ok {
		m[Str(key)] = value
	}
}

// Len returns the length of a slice, map, or string. Returns 0 for nil/unknown types.
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
		// Use reflection for typed slices ([]int, []string, []float64, etc.)
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			return rv.Len()
		}
		if rv.Kind() == reflect.Map {
			return rv.Len()
		}
		return 0
	}
}

// ToSlice converts an any value to []any for safe iteration.
// Handles []any, []map[string]any, and typed slices ([]int, []string, etc.) via reflection.
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
		// Use reflection for typed slices ([]int, []string, []float64, etc.)
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
// Handles map[string]any and typed maps (map[string]string, map[string]bool, etc.) via reflection.
func ToMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	// Use reflection for typed maps (map[string]string, map[string]bool, etc.)
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

// Keys returns the keys of a map as a slice of strings.
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

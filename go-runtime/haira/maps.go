package haira

import (
	"fmt"
	"reflect"
)

// MapLen returns the number of entries in a map.
func MapLen(m any) int {
	return Len(m)
}

// MapIsEmpty returns true if the map is empty.
func MapIsEmpty(m any) bool {
	return Len(m) == 0
}

// MapGet returns the value for a key, or nil if not found.
func MapGet(m any, key string) any {
	if mp, ok := m.(map[string]any); ok {
		return mp[key]
	}
	return nil
}

// MapHas returns true if the map contains the key.
func MapHas(m any, key string) bool {
	if mp, ok := m.(map[string]any); ok {
		_, exists := mp[key]
		return exists
	}
	return false
}

// MapSet sets a key-value pair in the map and returns the map.
func MapSet(m any, key string, value any) map[string]any {
	if mp, ok := m.(map[string]any); ok {
		mp[key] = value
		return mp
	}
	return map[string]any{key: value}
}

// MapRemove removes a key from the map and returns the map.
func MapRemove(m any, key string) map[string]any {
	if mp, ok := m.(map[string]any); ok {
		delete(mp, key)
		return mp
	}
	return map[string]any{}
}

// MapKeys returns all keys as a list.
func MapKeys(m any) []any {
	return Keys(m)
}

// MapValues returns all values as a list.
func MapValues(m any) []any {
	if mp, ok := m.(map[string]any); ok {
		vals := make([]any, 0, len(mp))
		for _, v := range mp {
			vals = append(vals, v)
		}
		return vals
	}
	return nil
}

// MapEntries returns all entries as a list of [key, value] pairs.
func MapEntries(m any) []any {
	if mp, ok := m.(map[string]any); ok {
		entries := make([]any, 0, len(mp))
		for k, v := range mp {
			entries = append(entries, []any{k, v})
		}
		return entries
	}
	return nil
}

// MapMerge merges two maps. Values from b override values from a.
// Handles map[string]any and typed maps (map[string]string, etc.) via reflection.
func MapMerge(a, b any) map[string]any {
	result := map[string]any{}
	mergeInto(result, a)
	mergeInto(result, b)
	return result
}

func mergeInto(dst map[string]any, src any) {
	if src == nil {
		return
	}
	if m, ok := src.(map[string]any); ok {
		for k, v := range m {
			dst[k] = v
		}
		return
	}
	// Reflection fallback for typed maps like map[string]string
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		for _, key := range rv.MapKeys() {
			dst[key.String()] = rv.MapIndex(key).Interface()
		}
	}
}

// MapContainsValue returns true if the map contains the given value.
func MapContainsValue(m any, value any) bool {
	if mp, ok := m.(map[string]any); ok {
		target := fmt.Sprintf("%v", value)
		for _, v := range mp {
			if fmt.Sprintf("%v", v) == target {
				return true
			}
		}
	}
	return false
}

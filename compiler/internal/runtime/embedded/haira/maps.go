package haira

import (
	"fmt"
	"reflect"
)

func MapLen(m any) int      { return Len(m) }
func MapIsEmpty(m any) bool { return Len(m) == 0 }

func MapGet(m any, key string) any {
	if mp, ok := m.(map[string]any); ok {
		return mp[key]
	}
	return nil
}

func MapHas(m any, key string) bool {
	if mp, ok := m.(map[string]any); ok {
		_, exists := mp[key]
		return exists
	}
	return false
}

func MapSet(m any, key string, value any) map[string]any {
	if mp, ok := m.(map[string]any); ok {
		mp[key] = value
		return mp
	}
	return map[string]any{key: value}
}

func MapRemove(m any, key string) map[string]any {
	if mp, ok := m.(map[string]any); ok {
		delete(mp, key)
		return mp
	}
	return map[string]any{}
}

func MapKeys(m any) []any { return Keys(m) }

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
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		for _, key := range rv.MapKeys() {
			dst[key.String()] = rv.MapIndex(key).Interface()
		}
	}
}

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

package haira

import "encoding/json"

// JSONMarshal encodes a value as a JSON string.
func JSONMarshal(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSONMarshalPretty encodes a value as a pretty-printed JSON string.
func JSONMarshalPretty(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JSONUnmarshal decodes a JSON string into a map.
func JSONUnmarshal(data string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// JSONParse decodes a JSON string into an any value (handles arrays, primitives, objects).
func JSONParse(data string) (any, error) {
	var result any
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

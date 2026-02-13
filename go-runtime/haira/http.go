package haira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Response wraps an HTTP response for Haira programs.
type Response struct {
	StatusCode int
	Body       string
}

// Json parses the response body as JSON and returns the result.
func (r *Response) Json() any {
	var result any
	json.Unmarshal([]byte(r.Body), &result)
	return result
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
		switch k := key.(type) {
		case int:
			if k >= 0 && k < len(o) {
				return o[k]
			}
		case float64:
			idx := int(k)
			if idx >= 0 && idx < len(o) {
				return o[idx]
			}
		}
	}
	return nil
}

// Str converts any value to a string.
func Str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// HttpGet performs an HTTP GET request and returns a Response.
func HttpGet(url string) (*Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}, nil
}

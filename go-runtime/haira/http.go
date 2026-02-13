package haira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func toInt(v any) int {
	switch k := v.(type) {
	case int:
		return k
	case float64:
		return int(k)
	case string:
		return -1
	default:
		return -1
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
// Handles any-typed values by converting to string first.
func Concat(a, b any) string {
	return Str(a) + Str(b)
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
		return nil
	}
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

// HttpGet performs an HTTP GET request and returns a Response.
func HttpGet(url string) (*Response, error) {
	return httpRequest("GET", url, nil, nil)
}

// HttpGetWithHeaders performs an HTTP GET request with custom headers.
func HttpGetWithHeaders(url string, headers map[string]any) (*Response, error) {
	return httpRequest("GET", url, nil, headers)
}

// HttpPost performs an HTTP POST request with a JSON body.
func HttpPost(url string, body any) (*Response, error) {
	return httpRequest("POST", url, body, nil)
}

// HttpPostWithHeaders performs an HTTP POST request with a JSON body and custom headers.
func HttpPostWithHeaders(url string, body any, headers map[string]any) (*Response, error) {
	return httpRequest("POST", url, body, headers)
}

// HttpPut performs an HTTP PUT request with a JSON body.
func HttpPut(url string, body any) (*Response, error) {
	return httpRequest("PUT", url, body, nil)
}

// HttpPutWithHeaders performs an HTTP PUT request with a JSON body and custom headers.
func HttpPutWithHeaders(url string, body any, headers map[string]any) (*Response, error) {
	return httpRequest("PUT", url, body, headers)
}

// HttpDelete performs an HTTP DELETE request.
func HttpDelete(url string) (*Response, error) {
	return httpRequest("DELETE", url, nil, nil)
}

// HttpDeleteWithHeaders performs an HTTP DELETE request with custom headers.
func HttpDeleteWithHeaders(url string, headers map[string]any) (*Response, error) {
	return httpRequest("DELETE", url, nil, headers)
}

// httpRequest is the internal helper for all HTTP methods.
func httpRequest(method, url string, body any, headers map[string]any) (*Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating %s request: %w", method, err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, Str(v))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http %s failed: %w", method, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
	}, nil
}

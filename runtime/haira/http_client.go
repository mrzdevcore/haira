package haira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient is an authenticated HTTP client with base URL, default headers, retry, and timeout.
type HTTPClient struct {
	BaseURL    string
	Headers    map[string]string
	MaxRetries int
	Timeout    time.Duration
}

// HttpClient creates a new HTTPClient with the given base URL and options.
// Options map supports: headers (map), retry (int), timeout (int ms).
func HttpClient(baseURL string, opts map[string]any) *HTTPClient {
	c := &HTTPClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Headers:    make(map[string]string),
		MaxRetries: 0,
		Timeout:    30 * time.Second,
	}
	if opts == nil {
		return c
	}
	if h, ok := opts["headers"]; ok {
		if hm, ok := h.(map[string]any); ok {
			for k, v := range hm {
				c.Headers[k] = Str(v)
			}
		}
	}
	if r, ok := opts["retry"]; ok {
		switch v := r.(type) {
		case int:
			c.MaxRetries = v
		case float64:
			c.MaxRetries = int(v)
		}
	}
	if t, ok := opts["timeout"]; ok {
		switch v := t.(type) {
		case int:
			c.Timeout = time.Duration(v) * time.Millisecond
		case float64:
			c.Timeout = time.Duration(int(v)) * time.Millisecond
		}
	}
	return c
}

// Get performs a GET request to the given path.
func (c *HTTPClient) Get(path string) (*Response, error) {
	return c.do("GET", path, nil)
}

// Post performs a POST request to the given path with a JSON body.
func (c *HTTPClient) Post(path string, body any) (*Response, error) {
	return c.do("POST", path, body)
}

// Put performs a PUT request to the given path with a JSON body.
func (c *HTTPClient) Put(path string, body any) (*Response, error) {
	return c.do("PUT", path, body)
}

// Patch performs a PATCH request to the given path with a JSON body.
func (c *HTTPClient) Patch(path string, body any) (*Response, error) {
	return c.do("PATCH", path, body)
}

// Delete performs a DELETE request to the given path.
func (c *HTTPClient) Delete(path string) (*Response, error) {
	return c.do("DELETE", path, nil)
}

func (c *HTTPClient) do(method, path string, body any) (*Response, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	var lastErr error
	attempts := c.MaxRetries + 1
	for i := 0; i < attempts; i++ {
		// Reset body reader for retries
		if body != nil && i > 0 {
			jsonBytes, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(jsonBytes)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating %s request: %w", method, err)
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range c.Headers {
			req.Header.Set(k, v)
		}

		client := &http.Client{Timeout: c.Timeout}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http %s failed: %w", method, err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			continue
		}

		// Retry on 5xx errors
		if resp.StatusCode >= 500 && i < attempts-1 {
			lastErr = fmt.Errorf("http %s returned %d", method, resp.StatusCode)
			continue
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}, nil
	}

	return nil, lastErr
}

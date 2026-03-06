package haira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Response wraps an HTTP response for Haira programs.
type Response struct {
	StatusCode int
	Body       string
}

// Json parses the response body as JSON and returns the result.
// Panics on parse error (caught by Haira's try/catch via recover).
func (r *Response) Json() map[string]any {
	var result map[string]any
	if err := json.Unmarshal([]byte(r.Body), &result); err != nil {
		panic(fmt.Sprintf("json parse error: %v (body: %.200s)", err, r.Body))
	}
	if result == nil {
		result = map[string]any{}
	}
	return result
}

// JsonArray parses the response body as a JSON array and returns the result.
// Panics on parse error (caught by Haira's try/catch via recover).
func (r *Response) JsonArray() []any {
	var result []any
	if err := json.Unmarshal([]byte(r.Body), &result); err != nil {
		panic(fmt.Sprintf("json array parse error: %v (body: %.200s)", err, r.Body))
	}
	return result
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

// HttpEncodeURI percent-encodes a string for use in URL query parameters.
func HttpEncodeURI(s string) string {
	return url.QueryEscape(s)
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

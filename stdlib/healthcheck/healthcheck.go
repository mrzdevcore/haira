package healthcheck

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CheckResult holds the result of a health check.
type CheckResult struct {
	Name         string
	URL          string
	Status       string // "HEALTHY", "DEGRADED", "DOWN", "UNREACHABLE"
	StatusCode   int
	ResponseTime string
	Error        string
	CheckedAt    string
}

// Check performs an HTTP health check on a single service.
// Returns a map compatible with Haira's map[string]any.
func Check(name, url string) map[string]any {
	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	elapsed := time.Since(start)
	elapsedStr := fmt.Sprintf("%dms", elapsed.Milliseconds())
	ts := time.Now().Format(time.RFC3339)

	if err != nil {
		return map[string]any{
			"name":          name,
			"url":           url,
			"status":        "UNREACHABLE",
			"status_code":   0,
			"response_time": elapsedStr,
			"error":         fmt.Sprintf("%v", err),
			"checked_at":    ts,
		}
	}
	defer resp.Body.Close()

	status := "HEALTHY"
	if resp.StatusCode >= 500 {
		status = "DOWN"
	} else if resp.StatusCode >= 400 {
		status = "DEGRADED"
	}

	return map[string]any{
		"name":          name,
		"url":           url,
		"status":        status,
		"status_code":   resp.StatusCode,
		"response_time": elapsedStr,
		"error":         "",
		"checked_at":    ts,
	}
}

// CheckAll performs health checks on multiple services concurrently.
// services is a list of maps with "name" and "url" keys.
func CheckAll(services []any) []any {
	type result struct {
		index int
		data  map[string]any
	}

	var wg sync.WaitGroup
	results := make(chan result, len(services))

	for i, svc := range services {
		svcMap, ok := svc.(map[string]any)
		if !ok {
			continue
		}
		name, _ := svcMap["name"].(string)
		url, _ := svcMap["url"].(string)
		if name == "" || url == "" {
			continue
		}

		wg.Add(1)
		go func(idx int, n, u string) {
			defer wg.Done()
			results <- result{index: idx, data: Check(n, u)}
		}(i, name, url)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	ordered := make([]any, len(services))
	for r := range results {
		ordered[r.index] = r.data
	}

	// Filter out nil entries (from invalid service entries)
	var out []any
	for _, r := range ordered {
		if r != nil {
			out = append(out, r)
		}
	}
	return out
}

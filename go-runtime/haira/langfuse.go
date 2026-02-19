package haira

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// ── Configuration ──

type langfuseConfig struct {
	host      string
	publicKey string
	secretKey string
}

// ── Exporter ──

type langfuseExporter struct {
	mu     sync.Mutex
	config langfuseConfig
	buffer []langfuseEvent
	traces map[string]bool // track which traceIds have been sent
	ticker *time.Ticker
	stopCh chan struct{}
}

var (
	globalLangfuse *langfuseExporter
	langfuseOnce   sync.Once
)

// ── Event types (Langfuse ingestion API) ──

type langfuseEvent struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Timestamp string         `json:"timestamp"`
	Body      map[string]any `json:"body"`
}

type langfusePayload struct {
	Batch    []langfuseEvent   `json:"batch"`
	Metadata map[string]string `json:"metadata"`
}

// ── Public API ──

// ObserveLangfuse enables Langfuse export. Pass empty strings to auto-detect
// from LANGFUSE_HOST, LANGFUSE_PUBLIC_KEY, LANGFUSE_SECRET_KEY env vars.
func ObserveLangfuse(host, publicKey, secretKey string) {
	if host == "" {
		host = os.Getenv("LANGFUSE_HOST")
	}
	if publicKey == "" {
		publicKey = os.Getenv("LANGFUSE_PUBLIC_KEY")
	}
	if secretKey == "" {
		secretKey = os.Getenv("LANGFUSE_SECRET_KEY")
	}
	if host == "" || publicKey == "" || secretKey == "" {
		fmt.Println("[haira] Langfuse: missing config (set LANGFUSE_HOST, LANGFUSE_PUBLIC_KEY, LANGFUSE_SECRET_KEY) — export disabled")
		return
	}
	langfuseOnce.Do(func() {
		exp := &langfuseExporter{
			config: langfuseConfig{
				host:      host,
				publicKey: publicKey,
				secretKey: secretKey,
			},
			buffer: make([]langfuseEvent, 0, 64),
			traces: make(map[string]bool),
			ticker: time.NewTicker(5 * time.Second),
			stopCh: make(chan struct{}),
		}
		globalLangfuse = exp
		go exp.runLoop()
		fmt.Printf("[haira] Langfuse export enabled → %s\n", host)
	})
}

// ── Internal hook (called from RecordGeneration in observe.go) ──

func langfuseEnqueue(gen LLMGeneration) {
	if globalLangfuse == nil {
		return
	}

	now := gen.Timestamp.UTC().Format(time.RFC3339Nano)
	endTime := gen.Timestamp.Add(time.Duration(gen.LatencyMs) * time.Millisecond).UTC().Format(time.RFC3339Nano)

	// Use session ID as trace key; fall back to a per-generation trace.
	traceKey := gen.SessionID
	if traceKey == "" {
		traceKey = newUUID()
	}
	traceID := "trace-" + traceKey

	inputCost := float64(gen.InputTokens) * gen.InputTokenCost / 1_000_000
	outputCost := float64(gen.OutputTokens) * gen.OutputTokenCost / 1_000_000

	globalLangfuse.mu.Lock()

	// Create a trace-create event if we haven't seen this trace yet.
	if !globalLangfuse.traces[traceID] {
		globalLangfuse.traces[traceID] = true
		traceBody := map[string]any{
			"id":        traceID,
			"timestamp": now,
			"name":      gen.AgentName,
		}
		if gen.SessionID != "" {
			traceBody["sessionId"] = gen.SessionID
		}
		globalLangfuse.buffer = append(globalLangfuse.buffer, langfuseEvent{
			ID:        newUUID(),
			Type:      "trace-create",
			Timestamp: now,
			Body:      traceBody,
		})
	}

	// Create the generation event linked to the trace.
	genBody := map[string]any{
		"id":        newUUID(),
		"traceId":   traceID,
		"name":      gen.AgentName,
		"model":     gen.Model,
		"startTime": now,
		"endTime":   endTime,
		"usageDetails": map[string]any{
			"input":  gen.InputTokens,
			"output": gen.OutputTokens,
			"total":  gen.TotalTokens,
		},
		"costDetails": map[string]any{
			"input":  inputCost,
			"output": outputCost,
		},
		"metadata": map[string]any{
			"provider":      gen.Provider,
			"temperature":   gen.Temperature,
			"finish_reason": gen.FinishReason,
			"tool_calls":    gen.ToolCalls,
		},
	}
	if gen.SessionID != "" {
		genBody["sessionId"] = gen.SessionID
	}

	globalLangfuse.buffer = append(globalLangfuse.buffer, langfuseEvent{
		ID:        newUUID(),
		Type:      "generation-create",
		Timestamp: now,
		Body:      genBody,
	})

	buffered := len(globalLangfuse.buffer)
	shouldFlush := buffered >= 50
	globalLangfuse.mu.Unlock()

	fmt.Fprintf(os.Stderr, "[haira] Langfuse: enqueued generation for %s (%d input, %d output) — buffer size: %d\n",
		gen.AgentName, gen.InputTokens, gen.OutputTokens, buffered)

	if shouldFlush {
		go globalLangfuse.flush()
	}
}

// ── Background loop ──

func (e *langfuseExporter) runLoop() {
	for {
		select {
		case <-e.ticker.C:
			e.flush()
		case <-e.stopCh:
			e.ticker.Stop()
			e.flush()
			return
		}
	}
}

func (e *langfuseExporter) flush() {
	e.mu.Lock()
	if len(e.buffer) == 0 {
		e.mu.Unlock()
		return
	}
	batch := make([]langfuseEvent, len(e.buffer))
	copy(batch, e.buffer)
	e.buffer = e.buffer[:0]
	e.mu.Unlock()

	payload := langfusePayload{
		Batch: batch,
		Metadata: map[string]string{
			"sdk":         "haira",
			"sdk_version": "0.1.0",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[haira] Langfuse marshal error: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST",
		e.config.host+"/api/public/ingestion",
		bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[haira] Langfuse request error: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	creds := base64.StdEncoding.EncodeToString(
		[]byte(e.config.publicKey + ":" + e.config.secretKey))
	req.Header.Set("Authorization", "Basic "+creds)

	// Log event types being sent
	typeCounts := make(map[string]int)
	for _, evt := range batch {
		typeCounts[evt.Type]++
	}
	fmt.Fprintf(os.Stderr, "[haira] Langfuse: flushing %d events %v\n", len(batch), typeCounts)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[haira] Langfuse send error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	// Log all responses — Langfuse returns 207 for partial success with per-event errors
	if resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "[haira] Langfuse HTTP %d (%d events): %s\n",
			resp.StatusCode, len(batch), string(respBody))
	} else {
		// Check for per-event errors in 200/207 response body
		var result struct {
			Errors []any `json:"errors"`
		}
		if json.Unmarshal(respBody, &result) == nil && len(result.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "[haira] Langfuse: %d events sent, %d errors: %s\n",
				len(batch), len(result.Errors), string(respBody))
		}
	}
}

// ── UUID v4 helper (crypto/rand, no external deps) ──

func newUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

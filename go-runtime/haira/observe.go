package haira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// ── Event types (OTel-aligned naming) ──

// LLMGeneration records a single LLM API call.
type LLMGeneration struct {
	ID           string    `json:"id"`
	AgentName    string    `json:"agent_name"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	TotalTokens  int       `json:"total_tokens"`
	LatencyMs    int64     `json:"latency_ms"`
	Temperature  float64   `json:"temperature"`
	ToolCalls    int       `json:"tool_calls"`
	FinishReason string    `json:"finish_reason"`
	Timestamp    time.Time `json:"timestamp"`
	SessionID    string    `json:"session_id"`
}

// ToolExec records a single tool execution.
type ToolExec struct {
	ID        string    `json:"id"`
	AgentName string    `json:"agent_name"`
	ToolName  string    `json:"tool_name"`
	LatencyMs int64     `json:"latency_ms"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
}

// ── Global observer (goroutine-safe) ──

type observer struct {
	mu          sync.RWMutex
	generations []LLMGeneration
	toolExecs   []ToolExec
	idCounter   uint64
}

var globalObserver = &observer{}

func (o *observer) nextID(prefix string) string {
	n := atomic.AddUint64(&o.idCounter, 1)
	return fmt.Sprintf("%s_%d", prefix, n)
}

// ── Internal recording (called by agent.go) ──

// RecordGeneration records an LLM generation event.
func RecordGeneration(gen LLMGeneration) {
	gen.ID = globalObserver.nextID("gen")
	globalObserver.mu.Lock()
	globalObserver.generations = append(globalObserver.generations, gen)
	globalObserver.mu.Unlock()
}

// RecordToolExec records a tool execution event.
func RecordToolExec(exec ToolExec) {
	exec.ID = globalObserver.nextID("tool")
	globalObserver.mu.Lock()
	globalObserver.toolExecs = append(globalObserver.toolExecs, exec)
	globalObserver.mu.Unlock()
}

// ── Helper: build summary from filtered slices ──

func buildSummary(gens []LLMGeneration, tools []ToolExec) map[string]any {
	var inputTokens, outputTokens, totalTokens int
	var totalLatencyMs int64
	for _, g := range gens {
		inputTokens += g.InputTokens
		outputTokens += g.OutputTokens
		totalTokens += g.TotalTokens
		totalLatencyMs += g.LatencyMs
	}
	for _, t := range tools {
		totalLatencyMs += t.LatencyMs
	}
	return map[string]any{
		"input_tokens":     inputTokens,
		"output_tokens":    outputTokens,
		"total_tokens":     totalTokens,
		"llm_calls":        len(gens),
		"tool_calls":       len(tools),
		"total_latency_ms": totalLatencyMs,
	}
}

// ── Public query functions (called from Haira via stdlib) ──

// ObserveUsage returns total usage across all agents.
func ObserveUsage() map[string]any {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	return buildSummary(globalObserver.generations, globalObserver.toolExecs)
}

// ObserveAgentUsage returns usage for a specific agent.
func ObserveAgentUsage(name string) map[string]any {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	var gens []LLMGeneration
	var tools []ToolExec
	for _, g := range globalObserver.generations {
		if g.AgentName == name {
			gens = append(gens, g)
		}
	}
	for _, t := range globalObserver.toolExecs {
		if t.AgentName == name {
			tools = append(tools, t)
		}
	}
	return buildSummary(gens, tools)
}

// ObserveSessionUsage returns usage for a specific session.
func ObserveSessionUsage(id string) map[string]any {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	var gens []LLMGeneration
	var tools []ToolExec
	for _, g := range globalObserver.generations {
		if g.SessionID == id {
			gens = append(gens, g)
		}
	}
	for _, t := range globalObserver.toolExecs {
		if t.SessionID == id {
			tools = append(tools, t)
		}
	}
	return buildSummary(gens, tools)
}

// ObserveModelUsage returns usage for a specific model.
func ObserveModelUsage(model string) map[string]any {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	var gens []LLMGeneration
	for _, g := range globalObserver.generations {
		if g.Model == model {
			gens = append(gens, g)
		}
	}
	return buildSummary(gens, nil)
}

// ObserveEvents returns all recorded events as a list of maps.
func ObserveEvents() []map[string]any {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	events := make([]map[string]any, 0, len(globalObserver.generations)+len(globalObserver.toolExecs))
	for _, g := range globalObserver.generations {
		b, _ := json.Marshal(g)
		var m map[string]any
		json.Unmarshal(b, &m)
		m["type"] = "generation"
		events = append(events, m)
	}
	for _, t := range globalObserver.toolExecs {
		b, _ := json.Marshal(t)
		var m map[string]any
		json.Unmarshal(b, &m)
		m["type"] = "tool_exec"
		events = append(events, m)
	}
	return events
}

// ObserveAgentEvents returns events for a specific agent.
func ObserveAgentEvents(name string) []map[string]any {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	var events []map[string]any
	for _, g := range globalObserver.generations {
		if g.AgentName == name {
			b, _ := json.Marshal(g)
			var m map[string]any
			json.Unmarshal(b, &m)
			m["type"] = "generation"
			events = append(events, m)
		}
	}
	for _, t := range globalObserver.toolExecs {
		if t.AgentName == name {
			b, _ := json.Marshal(t)
			var m map[string]any
			json.Unmarshal(b, &m)
			m["type"] = "tool_exec"
			events = append(events, m)
		}
	}
	return events
}

// ObserveReset clears all recorded events.
func ObserveReset() {
	globalObserver.mu.Lock()
	globalObserver.generations = nil
	globalObserver.toolExecs = nil
	globalObserver.mu.Unlock()
}

// ── Dashboard activation ──

// ObserveStart attaches the observe dashboard to a Server or starts a standalone server on a port.
func ObserveStart(target any) {
	switch t := target.(type) {
	case *Server:
		ObserveStartServer(t)
	case int:
		ObserveStartPort(t)
	case float64:
		ObserveStartPort(int(t))
	}
}

// ObserveStartServer attaches /_observe API + dashboard routes to an existing Server.
func ObserveStartServer(s *Server) {
	s.mux.HandleFunc("/_observe/api/usage", handleObserveUsageAPI)
	s.mux.HandleFunc("/_observe/api/events", handleObserveEventsAPI)
	s.mux.HandleFunc("/_observe/api/agents", handleObserveAgentsAPI)
	s.mux.HandleFunc("/_observe", handleObserveDashboard)
	s.mux.HandleFunc("/_observe/", handleObserveDashboard)
	fmt.Println("[haira] Observe dashboard: /_observe")
}

// ObserveStartPort starts a standalone observe dashboard on the given port.
func ObserveStartPort(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/_observe/api/usage", handleObserveUsageAPI)
	mux.HandleFunc("/_observe/api/events", handleObserveEventsAPI)
	mux.HandleFunc("/_observe/api/agents", handleObserveAgentsAPI)
	mux.HandleFunc("/_observe", handleObserveDashboard)
	mux.HandleFunc("/_observe/", handleObserveDashboard)
	fmt.Printf("[haira] Observe dashboard: http://localhost:%d/_observe\n", port)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

// ── HTTP handlers ──

func handleObserveUsageAPI(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	// Support ?agent=X, ?session=X, ?model=X filters
	if agent := r.URL.Query().Get("agent"); agent != "" {
		json.NewEncoder(rw).Encode(ObserveAgentUsage(agent))
		return
	}
	if session := r.URL.Query().Get("session"); session != "" {
		json.NewEncoder(rw).Encode(ObserveSessionUsage(session))
		return
	}
	if model := r.URL.Query().Get("model"); model != "" {
		json.NewEncoder(rw).Encode(ObserveModelUsage(model))
		return
	}
	json.NewEncoder(rw).Encode(ObserveUsage())
}

func handleObserveEventsAPI(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	agent := r.URL.Query().Get("agent")
	session := r.URL.Query().Get("session")
	var events []map[string]any
	if agent != "" {
		events = ObserveAgentEvents(agent)
	} else {
		events = ObserveEvents()
	}
	if session != "" {
		var filtered []map[string]any
		for _, e := range events {
			if e["session_id"] == session {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}
	if events == nil {
		events = []map[string]any{}
	}
	json.NewEncoder(rw).Encode(events)
}

func handleObserveAgentsAPI(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	agentSet := make(map[string]bool)
	for _, g := range globalObserver.generations {
		agentSet[g.AgentName] = true
	}
	for _, t := range globalObserver.toolExecs {
		agentSet[t.AgentName] = true
	}
	agents := make([]string, 0, len(agentSet))
	for name := range agentSet {
		agents = append(agents, name)
	}
	json.NewEncoder(rw).Encode(agents)
}

func handleObserveDashboard(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(observeHTML))
}

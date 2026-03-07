package haira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ── Cost computation (prices come from provider declaration) ──

func computeCost(inputTokens, outputTokens int, inputCostPerMillion, outputCostPerMillion float64) float64 {
	if inputCostPerMillion == 0 && outputCostPerMillion == 0 {
		return 0
	}
	return (float64(inputTokens)*inputCostPerMillion + float64(outputTokens)*outputCostPerMillion) / 1_000_000
}

// ── Event types (OTel-aligned naming) ──

// LLMGeneration records a single LLM API call.
type LLMGeneration struct {
	ID              string    `json:"id"`
	AgentName       string    `json:"agent_name"`
	Model           string    `json:"model"`
	Provider        string    `json:"provider"`
	InputTokens     int       `json:"input_tokens"`
	OutputTokens    int       `json:"output_tokens"`
	TotalTokens     int       `json:"total_tokens"`
	CostUSD         float64   `json:"cost_usd"`
	LatencyMs       int64     `json:"latency_ms"`
	Temperature     float64   `json:"temperature"`
	ToolCalls       int       `json:"tool_calls"`
	FinishReason    string    `json:"finish_reason"`
	Timestamp       time.Time `json:"timestamp"`
	SessionID       string    `json:"session_id"`
	RunID           string    `json:"run_id,omitempty"`
	InputTokenCost  float64   `json:"-"` // USD per 1M input tokens (from provider, not serialized)
	OutputTokenCost float64   `json:"-"` // USD per 1M output tokens (from provider, not serialized)
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
	RunID     string    `json:"run_id,omitempty"`
}

// ── Exporter interface (for external services like Langfuse, Datadog, etc.) ──

// Exporter receives LLM generation events for external observability platforms.
type Exporter interface {
	OnGeneration(LLMGeneration)
}

var exporters []Exporter

// ObserveExport registers an exporter to receive all LLM generation events.
func ObserveExport(exp Exporter) {
	if exp != nil {
		exporters = append(exporters, exp)
	}
}

// ── Goroutine-local run ID tracking ──
// Allows workflow runs to tag all LLM/tool events with the current run ID.

var activeRunIDs sync.Map // goroutine ID → run ID string

// SetActiveRunID associates a run ID with the current goroutine.
func SetActiveRunID(runID string) {
	activeRunIDs.Store(goid(), runID)
}

// ClearActiveRunID removes the run ID association for the current goroutine.
func ClearActiveRunID() {
	activeRunIDs.Delete(goid())
}

// currentRunID returns the run ID for the current goroutine, if any.
func currentRunID() string {
	if v, ok := activeRunIDs.Load(goid()); ok {
		return v.(string)
	}
	return ""
}

// ── Global observer (goroutine-safe) ──

const maxObserveEntries = 10000

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
	gen.CostUSD = computeCost(gen.InputTokens, gen.OutputTokens, gen.InputTokenCost, gen.OutputTokenCost)
	if gen.RunID == "" {
		gen.RunID = currentRunID()
	}
	globalObserver.mu.Lock()
	globalObserver.generations = append(globalObserver.generations, gen)
	if len(globalObserver.generations) > maxObserveEntries {
		globalObserver.generations = globalObserver.generations[len(globalObserver.generations)-maxObserveEntries:]
	}
	globalObserver.mu.Unlock()
	if globalStore != nil {
		globalStore.SaveGeneration(gen)
	}
	for _, exp := range exporters {
		exp.OnGeneration(gen)
	}
}

// RecordToolExec records a tool execution event.
func RecordToolExec(exec ToolExec) {
	exec.ID = globalObserver.nextID("tool")
	if exec.RunID == "" {
		exec.RunID = currentRunID()
	}
	globalObserver.mu.Lock()
	globalObserver.toolExecs = append(globalObserver.toolExecs, exec)
	if len(globalObserver.toolExecs) > maxObserveEntries {
		globalObserver.toolExecs = globalObserver.toolExecs[len(globalObserver.toolExecs)-maxObserveEntries:]
	}
	globalObserver.mu.Unlock()
	if globalStore != nil {
		globalStore.SaveToolExec(exec)
	}
}

// ObserveLoadFromStore loads persisted observability data from the store on startup.
func ObserveLoadFromStore() {
	if globalStore == nil {
		return
	}
	gens, err := globalStore.LoadGenerations()
	if err != nil {
		return
	}
	tools, errT := globalStore.LoadToolExecs()
	if errT != nil {
		return
	}
	globalObserver.mu.Lock()
	globalObserver.generations = gens
	globalObserver.toolExecs = tools
	// Advance ID counter past loaded IDs to avoid collisions
	var maxID uint64
	for _, g := range gens {
		var n uint64
		fmt.Sscanf(g.ID, "gen_%d", &n)
		if n > maxID {
			maxID = n
		}
	}
	for _, t := range tools {
		var n uint64
		fmt.Sscanf(t.ID, "tool_%d", &n)
		if n > maxID {
			maxID = n
		}
	}
	atomic.StoreUint64(&globalObserver.idCounter, maxID)
	globalObserver.mu.Unlock()
}

// ── Helper: build summary from filtered slices ──

func buildSummary(gens []LLMGeneration, tools []ToolExec) map[string]any {
	var inputTokens, outputTokens, totalTokens int
	var totalLatencyMs int64
	var estimatedCostUSD float64
	for _, g := range gens {
		inputTokens += g.InputTokens
		outputTokens += g.OutputTokens
		totalTokens += g.TotalTokens
		totalLatencyMs += g.LatencyMs
		estimatedCostUSD += g.CostUSD
	}
	for _, t := range tools {
		totalLatencyMs += t.LatencyMs
	}
	return map[string]any{
		"input_tokens":       inputTokens,
		"output_tokens":      outputTokens,
		"total_tokens":       totalTokens,
		"llm_calls":          len(gens),
		"tool_calls":         len(tools),
		"total_latency_ms":   totalLatencyMs,
		"estimated_cost_usd": estimatedCostUSD,
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

// ObserveCost returns total estimated cost in USD across all agents.
func ObserveCost() float64 {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	var total float64
	for _, g := range globalObserver.generations {
		total += g.CostUSD
	}
	return total
}

// ObserveAgentCost returns estimated cost in USD for a specific agent.
func ObserveAgentCost(name string) float64 {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	var total float64
	for _, g := range globalObserver.generations {
		if g.AgentName == name {
			total += g.CostUSD
		}
	}
	return total
}

// sortEventsByTimestamp sorts events chronologically (most recent first).
func sortEventsByTimestamp(events []map[string]any) {
	sort.Slice(events, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339Nano, fmt.Sprint(events[i]["timestamp"]))
		tj, _ := time.Parse(time.RFC3339Nano, fmt.Sprint(events[j]["timestamp"]))
		return ti.After(tj)
	})
}

// ObserveEvents returns all recorded events sorted chronologically (most recent first).
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
	sortEventsByTimestamp(events)
	return events
}

// ObserveAgentEvents returns events for a specific agent, sorted chronologically.
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
	sortEventsByTimestamp(events)
	return events
}

// ObserveRunIDs returns all unique run IDs from recorded events, most recent first.
func ObserveRunIDs() []string {
	globalObserver.mu.RLock()
	defer globalObserver.mu.RUnlock()
	seen := make(map[string]time.Time)
	for _, g := range globalObserver.generations {
		if g.RunID != "" {
			if t, ok := seen[g.RunID]; !ok || g.Timestamp.After(t) {
				seen[g.RunID] = g.Timestamp
			}
		}
	}
	for _, t := range globalObserver.toolExecs {
		if t.RunID != "" {
			if ts, ok := seen[t.RunID]; !ok || t.Timestamp.After(ts) {
				seen[t.RunID] = t.Timestamp
			}
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return seen[ids[i]].After(seen[ids[j]])
	})
	return ids
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
	s.mux.HandleFunc("/_observe/api/run_ids", handleObserveRunIDsAPI)
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
	mux.HandleFunc("/_observe/api/run_ids", handleObserveRunIDsAPI)
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
	runID := r.URL.Query().Get("run")
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
	if runID != "" {
		var filtered []map[string]any
		for _, e := range events {
			if e["run_id"] == runID {
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

func handleObserveRunIDsAPI(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	ids := ObserveRunIDs()
	if ids == nil {
		ids = []string{}
	}
	json.NewEncoder(rw).Encode(ids)
}

func handleObserveDashboard(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(observeHTML))
}

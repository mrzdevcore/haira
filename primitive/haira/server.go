package haira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Server is an HTTP server that routes requests to workflows.
type Server struct {
	mux       *http.ServeMux
	workflows []*WorkflowDef
}

// NewServer creates a new server with the given workflow definitions.
func NewServer(workflows []*WorkflowDef) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		workflows: workflows,
	}

	// Initialize unified storage (SQLite or Postgres based on HAIRA_DATABASE_URL)
	if err := InitStore(); err != nil {
		fmt.Fprintf(os.Stderr, "haira: failed to init store: %v\n", err)
	}

	// Register ARP protocol + API routes
	s.registerProtocolRoutes()

	// Register embedded web UI (disable with HAIRA_NO_UI=1)
	s.registerUIRoutes()

	// Cleanup stale session upload directories in the background
	StartSessionCleanup(1*time.Hour, 24*time.Hour)

	for _, w := range workflows {
		wf := w // capture for closure
		s.mux.HandleFunc(wf.Path, func(rw http.ResponseWriter, r *http.Request) {
			if r.Method != wf.Method {
				http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			params := make(map[string]any)

			if r.Method == "GET" || r.Method == "DELETE" {
				// GET/DELETE: params from query string
				for k, v := range r.URL.Query() {
					params[k] = v[0]
				}
			} else {
				// POST/PUT: check for file params → multipart, else JSON body
				hasFileParams := false
				for _, p := range wf.Params {
					if p.Type == "file" {
						hasFileParams = true
						break
					}
				}

				if hasFileParams && strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
					r.ParseMultipartForm(32 << 20) // 32MB max

					// Stream/chat workflows use session-scoped storage (files persist across requests)
					isSessionScoped := wf.IsStream || wf.StreamHandler != nil
					sessionID := ""
					if isSessionScoped {
						sessionID = r.FormValue("session_id")
					}

					var tempFiles []string
					for _, p := range wf.Params {
						if p.Type == "file" {
							file, header, err := r.FormFile(p.Name)
							if err == nil {
								if isSessionScoped && sessionID != "" {
									path, saveErr := saveSessionFile(file, header.Filename, sessionID)
									file.Close()
									if saveErr == nil {
										params[p.Name] = path
									}
								} else {
									path, saveErr := saveTempFile(file, header.Filename)
									file.Close()
									if saveErr == nil {
										params[p.Name] = path
										tempFiles = append(tempFiles, path)
									}
								}
							}
						} else {
							if v := r.FormValue(p.Name); v != "" {
								params[p.Name] = v
							}
						}
					}
					// Only cleanup non-session temp files
					if len(tempFiles) > 0 {
						defer func() {
							for _, f := range tempFiles {
								os.Remove(f)
							}
						}()
					}
				} else {
					if r.Body != nil {
						defer r.Body.Close()
						json.NewDecoder(r.Body).Decode(&params)
					}
					if params == nil {
						params = make(map[string]any)
					}
				}
				// Include query parameters as fallback
				for k, v := range r.URL.Query() {
					if _, exists := params[k]; !exists {
						params[k] = v[0]
					}
				}
			}

			// Check if client wants streaming (Accept: text/event-stream)
			if r.Header.Get("Accept") == "text/event-stream" {
				if wf.StreamHandler != nil {
					s.handleSSE(rw, r, wf, params)
					return
				}
				if len(wf.Steps) > 0 {
					s.handleStepSSE(rw, r, wf, params)
					return
				}
			}

			result, err := wf.Handler(params)
			if err != nil {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(map[string]any{"error": err.Error()})
				return
			}

			rw.Header().Set("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(result)
		})
	}

	return s
}

func (s *Server) handleSSE(rw http.ResponseWriter, r *http.Request, wf *WorkflowDef, params map[string]any) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// --- Chat session tracking ---
	sessionID, _ := params["session_id"].(string)
	if sessionID != "" {
		globalStore.EnsureSession(sessionID, wf.Name, wf.Path, "")
		// Find the chat message param and record the user message
		chatParam := findChatParam(wf.Params)
		if userMsg, ok := params[chatParam].(string); ok && userMsg != "" {
			globalStore.AddMessage(sessionID, "user", userMsg, nil)
		}
	}

	ch, err := wf.StreamHandler(params)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]any{"error": err.Error()})
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	// Bridge StreamChunks to ARP messages, then write as SSE
	arpMessages := ArpBridge(sessionID, ch)
	result := WriteArpSSE(rw, flusher, arpMessages)

	// Persist assistant reply with UI events
	if sessionID != "" && (result.FullReply != "" || len(result.UIEvents) > 0) {
		globalStore.AddMessage(sessionID, "assistant", result.FullReply, result.UIEvents)
	}
}

// handleStepSSE runs a sync workflow handler in a goroutine, streaming step events
// via SSE, then sends the final result.
func (s *Server) handleStepSSE(rw http.ResponseWriter, r *http.Request, wf *WorkflowDef, params map[string]any) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	// Create a run in the store and send the run ID to the client
	runID := nextRunID()
	now := time.Now()
	run := &Run{
		ID:           runID,
		WorkflowName: wf.Name,
		WorkflowPath: wf.Path,
		Status:       RunStatusRunning,
		Params:       sanitizeParams(params),
		Steps:        []StepEvent{},
		StartedAt:    now,
	}
	globalStore.CreateRun(run)
	globalRunSubs.TrackRun(run)

	idData, _ := json.Marshal(map[string]string{"run_id": runID})
	fmt.Fprintf(rw, "event: run_id\ndata: %s\n\n", idData)
	flusher.Flush()

	stepCh := make(chan StepEvent, 16)
	type handlerResult struct {
		data any
		err  error
	}
	resultCh := make(chan handlerResult, 1)

	go func() {
		SetStepNotifier(stepCh)
		SetActiveRunID(runID)
		defer ClearStepNotifier()
		defer ClearActiveRunID()
		defer close(stepCh)

		result, err := wf.Handler(params)
		resultCh <- handlerResult{data: result, err: err}
	}()

	// Stream step events until the handler completes
	for event := range stepCh {
		globalRunSubs.RecordStepEvent(runID, event)
		globalStore.UpdateRun(run) // run.Steps updated by RecordStepEvent (same pointer)
		data, _ := json.Marshal(event)
		fmt.Fprintf(rw, "event: step\ndata: %s\n\n", data)
		flusher.Flush()
	}

	// Send final result
	res := <-resultCh
	finishedAt := time.Now()
	if res.err != nil {
		run.Status = RunStatusFailed
		run.Error = res.err.Error()
		run.FinishedAt = &finishedAt
		globalStore.UpdateRun(run)
		globalRunSubs.FinishRun(runID)
		errData, _ := json.Marshal(map[string]string{"error": res.err.Error()})
		fmt.Fprintf(rw, "event: error\ndata: %s\n\n", errData)
	} else {
		run.Status = RunStatusCompleted
		run.Result = res.data
		run.FinishedAt = &finishedAt
		globalStore.UpdateRun(run)
		globalRunSubs.FinishRun(runID)
		resultData, _ := json.Marshal(res.data)
		fmt.Fprintf(rw, "event: result\ndata: %s\n\n", resultData)
	}
	fmt.Fprintf(rw, "data: [DONE]\n\n")
	flusher.Flush()
}

// sanitizeParams creates a copy of params with file paths replaced by filenames only.
func sanitizeParams(params map[string]any) map[string]any {
	clean := make(map[string]any, len(params))
	for k, v := range params {
		if s, ok := v.(string); ok && strings.HasPrefix(s, "/tmp/") {
			// Replace temp file paths with just the filename
			parts := strings.Split(s, "/")
			clean[k] = parts[len(parts)-1]
		} else {
			clean[k] = v
		}
	}
	return clean
}

// registerProtocolRoutes sets up ARP protocol and data API routes.
func (s *Server) registerProtocolRoutes() {
	// ARP WebSocket endpoint
	s.mux.HandleFunc("/_arp/v1", s.arpWebSocketHandler())

	// ARP capability discovery (JSON, no WebSocket required)
	s.mux.HandleFunc("/_api/arp", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(ArpServerCapabilities())
	})

	// Workflow discovery API (used by haira console and custom clients)
	s.mux.HandleFunc("/_api/workflows", s.handleListWorkflows)

	// Run history API
	s.mux.HandleFunc("/_api/runs", s.handleListRuns)
	s.mux.HandleFunc("/_api/runs/", s.handleRunRoute)

	// Chat session API
	s.mux.HandleFunc("/_api/chats", s.handleListChats)
	s.mux.HandleFunc("/_api/chats/", s.handleChatRoute)
}

// --- Run History API ---

func (s *Server) handleListRuns(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	wfPath := r.URL.Query().Get("workflow")
	runs, err := globalStore.ListRuns(wfPath)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(runs)
}

func (s *Server) handleRunRoute(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/_api/runs/")

	// /_api/runs/stream/{id} — live SSE reconnection
	if strings.HasPrefix(path, "stream/") {
		runID := strings.TrimPrefix(path, "stream/")
		s.handleRunStream(rw, r, runID)
		return
	}

	// /_api/runs/{id} — get full run
	run, err := globalStore.GetRun(path)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if run == nil {
		http.NotFound(rw, r)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(run)
}

// handleRunStream replays buffered events and streams live events for in-progress runs.
func (s *Server) handleRunStream(rw http.ResponseWriter, r *http.Request, runID string) {
	run, err := globalStore.GetRun(runID)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if run == nil {
		http.NotFound(rw, r)
		return
	}

	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	// Replay all buffered events from the in-memory tracker (or from DB)
	buffered, status := globalRunSubs.GetSteps(runID)
	if buffered == nil {
		// Not tracked in-memory (already finished), use DB data
		buffered = run.Steps
		status = run.Status
	}

	for _, event := range buffered {
		data, _ := json.Marshal(event)
		fmt.Fprintf(rw, "event: step\ndata: %s\n\n", data)
	}
	flusher.Flush()

	// If already finished, send result and done
	if status != RunStatusRunning {
		if run.Error != "" {
			errData, _ := json.Marshal(map[string]string{"error": run.Error})
			fmt.Fprintf(rw, "event: error\ndata: %s\n\n", errData)
		} else if run.Result != nil {
			resultData, _ := json.Marshal(run.Result)
			fmt.Fprintf(rw, "event: result\ndata: %s\n\n", resultData)
		}
		fmt.Fprintf(rw, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}

	// Subscribe to live updates for in-progress run
	ch := globalRunSubs.Subscribe(runID)
	defer globalRunSubs.Unsubscribe(runID, ch)

	// Detect client disconnect
	ctx := r.Context()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				// Channel closed — run completed. Fetch final state.
				finalRun, _ := globalStore.GetRun(runID)
				if finalRun != nil {
					if finalRun.Error != "" {
						errData, _ := json.Marshal(map[string]string{"error": finalRun.Error})
						fmt.Fprintf(rw, "event: error\ndata: %s\n\n", errData)
					} else if finalRun.Result != nil {
						resultData, _ := json.Marshal(finalRun.Result)
						fmt.Fprintf(rw, "event: result\ndata: %s\n\n", resultData)
					}
				}
				fmt.Fprintf(rw, "data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(rw, "event: step\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

// --- Chat Session API ---

func (s *Server) handleListChats(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	wfPath := r.URL.Query().Get("workflow")
	owner := r.URL.Query().Get("owner")
	chats, err := globalStore.ListSessions(wfPath, owner)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(chats)
}

func (s *Server) handleChatRoute(rw http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/_api/chats/")
	if id == "" {
		http.NotFound(rw, r)
		return
	}

	switch r.Method {
	case "GET":
		sess, err := globalStore.GetSession(id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if sess == nil {
			http.NotFound(rw, r)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(sess)
	case "DELETE":
		globalStore.DeleteSession(id)
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]string{"status": "deleted"})
	default:
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Workflow Discovery API ---

func (s *Server) handleListWorkflows(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type wfInfo struct {
		Name        string          `json:"name"`
		Path        string          `json:"path"`
		Method      string          `json:"method"`
		IsStream    bool            `json:"is_stream"`
		Steps       []string        `json:"steps,omitempty"`
		ChatParam   string          `json:"chat_param,omitempty"`
		Title       string          `json:"title,omitempty"`
		Description string          `json:"description,omitempty"`
		Params      []WorkflowParam `json:"params,omitempty"`
		Suggestions []string        `json:"suggestions,omitempty"`
	}
	var list []wfInfo
	for _, wf := range s.workflows {
		info := wfInfo{
			Name:        wf.Name,
			Path:        wf.Path,
			Method:      wf.Method,
			IsStream:    wf.IsStream || wf.StreamHandler != nil,
			Steps:       wf.Steps,
			Title:       wf.UITitle,
			Description: wf.UIDescription,
			Params:      wf.Params,
			Suggestions: wf.Suggestions,
		}
		if info.IsStream {
			info.ChatParam = findChatParam(wf.Params)
		}
		list = append(list, info)
	}
	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(list)
}

// Handler returns the CORS-wrapped HTTP handler for this server.
// Used by the Workers target: workers.Serve(server.Handler())
func (s *Server) Handler() http.Handler {
	return corsMiddleware(s.mux)
}

// Listen starts the HTTP server on the given port.
// If HAIRA_PORT is set, it overrides the port argument.
func (s *Server) Listen(port int) error {
	if envPort := os.Getenv("HAIRA_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	addr := fmt.Sprintf(":%d", port)

	return http.ListenAndServe(addr, s.Handler())
}

// corsMiddleware adds CORS headers so external UI renderers (CDN, dev server)
// can access the ARP and API endpoints.
func corsMiddleware(next http.Handler) http.Handler {
	origin := os.Getenv("HAIRA_CORS_ORIGIN")
	if origin == "" {
		origin = "*"
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		if r.Method == "OPTIONS" {
			rw.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

package haira

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

// MCPServer exposes Haira workflows as MCP tools over stdio or SSE.
type MCPServer struct {
	workflows []*WorkflowDef
	// stdio transport
	reader *bufio.Reader
	writer io.Writer
	mu     sync.Mutex
	// SSE transport
	clients   map[string]*sseClient
	clientsMu sync.RWMutex
	sessionID atomic.Int64
}

// sseClient represents a connected SSE client session.
type sseClient struct {
	msgCh chan []byte
	done  chan struct{}
}

// NewMCPServer creates an MCP server from a list of workflow definitions.
func NewMCPServer(workflows []*WorkflowDef) *MCPServer {
	return &MCPServer{
		workflows: workflows,
		reader:    bufio.NewReader(os.Stdin),
		writer:    os.Stdout,
		clients:   make(map[string]*sseClient),
	}
}

// ---------------------------------------------------------------------------
// Shared dispatch
// ---------------------------------------------------------------------------

type jsonrpcErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// dispatch processes a JSON-RPC request and returns (result, error).
// Exactly one of result or errObj will be non-nil.
func (s *MCPServer) dispatch(req jsonrpcRequest) (any, *jsonrpcErrorObj) {
	switch req.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "haira",
				"version": "1.0.0",
			},
		}, nil

	case "tools/list":
		var tools []map[string]any
		for _, wf := range s.workflows {
			tools = append(tools, map[string]any{
				"name":        wf.Name,
				"description": wf.Description,
				"inputSchema": s.buildInputSchema(wf),
			})
		}
		return map[string]any{"tools": tools}, nil

	case "tools/call":
		return s.dispatchToolsCall(req.Params)

	case "ping":
		return map[string]any{}, nil

	default:
		return nil, &jsonrpcErrorObj{Code: -32601, Message: fmt.Sprintf("method not found: %s", req.Method)}
	}
}

func (s *MCPServer) dispatchToolsCall(params json.RawMessage) (any, *jsonrpcErrorObj) {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &call); err != nil {
		return nil, &jsonrpcErrorObj{Code: -32602, Message: fmt.Sprintf("invalid params: %v", err)}
	}

	var wf *WorkflowDef
	for _, w := range s.workflows {
		if w.Name == call.Name {
			wf = w
			break
		}
	}
	if wf == nil {
		return nil, &jsonrpcErrorObj{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", call.Name)}
	}

	var args map[string]any
	if call.Arguments != nil {
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			return nil, &jsonrpcErrorObj{Code: -32602, Message: fmt.Sprintf("invalid arguments: %v", err)}
		}
	}
	if args == nil {
		args = map[string]any{}
	}

	result, err := wf.Handler(args)
	if err != nil {
		return map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("error: %v", err)},
			},
			"isError": true,
		}, nil
	}

	resultJSON, _ := json.Marshal(result)
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(resultJSON)},
		},
	}, nil
}

func (s *MCPServer) buildInputSchema(wf *WorkflowDef) map[string]any {
	properties := map[string]any{}
	var required []string

	for _, p := range wf.Params {
		jsonType := "string"
		switch p.Type {
		case "int", "float":
			jsonType = "number"
		case "bool":
			jsonType = "boolean"
		}
		properties[p.Name] = map[string]any{"type": jsonType}
		required = append(required, p.Name)
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// ---------------------------------------------------------------------------
// Stdio transport
// ---------------------------------------------------------------------------

// Serve blocks on stdin and processes MCP JSON-RPC messages (stdio transport).
func (s *MCPServer) Serve() error {
	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("MCP server read: %w", err)
		}

		trimmed := strings.TrimSpace(string(line))
		if trimmed == "" {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal([]byte(trimmed), &req); err != nil {
			continue
		}

		// Notifications have no ID — don't respond
		if req.ID == nil {
			continue
		}

		id := *req.ID
		result, errObj := s.dispatch(req)
		if errObj != nil {
			s.sendStdioError(id, errObj.Code, errObj.Message)
		} else {
			s.sendStdioResponse(id, result)
		}
	}
}

func (s *MCPServer) sendStdioResponse(id int64, result any) {
	s.writeStdioJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	})
}

func (s *MCPServer) sendStdioError(id int64, code int, message string) {
	s.writeStdioJSON(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func (s *MCPServer) writeStdioJSON(v any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, _ := json.Marshal(v)
	s.writer.Write(append(data, '\n'))
}

// ---------------------------------------------------------------------------
// SSE transport
// ---------------------------------------------------------------------------

// Listen starts an HTTP server with SSE transport for MCP on the given port.
// Endpoints:
//   - GET /sse          — SSE stream (sends endpoint event, then message events)
//   - POST /messages    — receives JSON-RPC requests, responds via SSE stream
func (s *MCPServer) Listen(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/sse", s.handleSSEConnect)
	mux.HandleFunc("/messages", s.handleSSEMessage)

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}

func (s *MCPServer) handleSSEConnect(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID := fmt.Sprintf("%d", s.sessionID.Add(1))
	client := &sseClient{
		msgCh: make(chan []byte, 64),
		done:  make(chan struct{}),
	}

	s.clientsMu.Lock()
	s.clients[sessionID] = client
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, sessionID)
		s.clientsMu.Unlock()
		close(client.done)
	}()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send endpoint event — tells the client where to POST requests
	fmt.Fprintf(w, "event: endpoint\ndata: /messages?session=%s\n\n", sessionID)
	flusher.Flush()

	// Stream messages until client disconnects
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-client.msgCh:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (s *MCPServer) handleSSEMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "missing session parameter", http.StatusBadRequest)
		return
	}

	s.clientsMu.RLock()
	client, ok := s.clients[sessionID]
	s.clientsMu.RUnlock()
	if !ok {
		http.Error(w, "unknown session", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var req jsonrpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON-RPC", http.StatusBadRequest)
		return
	}

	// Notifications have no ID — accept but don't respond
	if req.ID == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	id := *req.ID
	result, errObj := s.dispatch(req)

	var resp map[string]any
	if errObj != nil {
		resp = map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"error": map[string]any{
				"code":    errObj.Code,
				"message": errObj.Message,
			},
		}
	} else {
		resp = map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  result,
		}
	}

	data, _ := json.Marshal(resp)

	// Send response via SSE stream
	select {
	case client.msgCh <- data:
	default:
		// Channel full — client is slow
	}

	w.WriteHeader(http.StatusAccepted)
}

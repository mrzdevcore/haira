package haira

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// MCPServer exposes Haira workflows as MCP tools over stdio.
type MCPServer struct {
	workflows []*WorkflowDef
	reader    *bufio.Reader
	writer    io.Writer
	mu        sync.Mutex
}

// NewMCPServer creates an MCP server from a list of workflow definitions.
func NewMCPServer(workflows []*WorkflowDef) *MCPServer {
	return &MCPServer{
		workflows: workflows,
		reader:    bufio.NewReader(os.Stdin),
		writer:    os.Stdout,
	}
}

// Serve blocks on stdin and processes MCP JSON-RPC messages.
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

		switch req.Method {
		case "initialize":
			s.handleInitialize(id)
		case "tools/list":
			s.handleToolsList(id)
		case "tools/call":
			s.handleToolsCall(id, req.Params)
		case "ping":
			s.sendResponse(id, map[string]any{})
		default:
			s.sendError(id, -32601, fmt.Sprintf("method not found: %s", req.Method))
		}
	}
}

func (s *MCPServer) handleInitialize(id int64) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "haira",
			"version": "1.0.0",
		},
	}
	s.sendResponse(id, result)
}

func (s *MCPServer) handleToolsList(id int64) {
	var tools []map[string]any
	for _, wf := range s.workflows {
		tools = append(tools, map[string]any{
			"name":        wf.Name,
			"description": wf.Description,
			"inputSchema": s.buildInputSchema(wf),
		})
	}
	s.sendResponse(id, map[string]any{"tools": tools})
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

func (s *MCPServer) handleToolsCall(id int64, params json.RawMessage) {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &call); err != nil {
		s.sendError(id, -32602, fmt.Sprintf("invalid params: %v", err))
		return
	}

	// Find matching workflow
	var wf *WorkflowDef
	for _, w := range s.workflows {
		if w.Name == call.Name {
			wf = w
			break
		}
	}
	if wf == nil {
		s.sendError(id, -32602, fmt.Sprintf("unknown tool: %s", call.Name))
		return
	}

	// Parse arguments into param map
	var args map[string]any
	if call.Arguments != nil {
		if err := json.Unmarshal(call.Arguments, &args); err != nil {
			s.sendError(id, -32602, fmt.Sprintf("invalid arguments: %v", err))
			return
		}
	}
	if args == nil {
		args = map[string]any{}
	}

	// Call workflow handler
	result, err := wf.Handler(args)
	if err != nil {
		s.sendResponse(id, map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("error: %v", err)},
			},
			"isError": true,
		})
		return
	}

	// Serialize result as JSON text
	resultJSON, _ := json.Marshal(result)
	s.sendResponse(id, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(resultJSON)},
		},
	})
}

func (s *MCPServer) sendResponse(id int64, result any) {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	s.writeJSON(resp)
}

func (s *MCPServer) sendError(id int64, code int, message string) {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
	s.writeJSON(resp)
}

func (s *MCPServer) writeJSON(v any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, _ := json.Marshal(v)
	s.writer.Write(append(data, '\n'))
}

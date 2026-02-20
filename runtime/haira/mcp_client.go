package haira

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

// MCPConfig holds the configuration for connecting to an MCP server.
type MCPConfig struct {
	Name      string            // provider name from Haira source
	Transport string            // "stdio" or "sse"
	Command   string            // for stdio: command to run (e.g. "npx")
	Args      []string          // for stdio: command arguments
	Endpoint  string            // for sse: HTTP endpoint URL
	Env       map[string]string // optional environment variables for subprocess
	Headers   map[string]string // optional HTTP headers for SSE
}

// MCPClient manages a connection to a single MCP server.
type MCPClient struct {
	config    MCPConfig
	transport mcpTransport
	tools     []*ToolDef
	mu        sync.Mutex
	nextID    atomic.Int64
}

// mcpTransport is the interface for MCP communication.
type mcpTransport interface {
	Send(msg json.RawMessage) error
	Receive() (json.RawMessage, error)
	Close() error
}

// JSON-RPC types

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP-specific response types

type mcpInitializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type mcpToolsListResult struct {
	Tools []mcpToolDef `json:"tools"`
}

type mcpToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type mcpToolCallResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// --- Stdio Transport ---

type mcpStdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
}

func newStdioTransport(command string, args []string, env map[string]string) (*mcpStdioTransport, error) {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr

	if len(env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server %q: %w", command, err)
	}

	return &mcpStdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}, nil
}

func (t *mcpStdioTransport) Send(msg json.RawMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := t.stdin.Write(append(msg, '\n'))
	return err
}

func (t *mcpStdioTransport) Receive() (json.RawMessage, error) {
	line, err := t.stdout.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return json.RawMessage(strings.TrimSpace(string(line))), nil
}

func (t *mcpStdioTransport) Close() error {
	t.stdin.Close()
	if t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}

// --- SSE Transport ---

type mcpSSETransport struct {
	endpoint string
	headers  map[string]string
	client   *http.Client
	msgCh    chan json.RawMessage
	postURL  string
	cancel   context.CancelFunc
	mu       sync.Mutex
}

func newSSETransport(endpoint string, headers map[string]string) (*mcpSSETransport, error) {
	ctx, cancel := context.WithCancel(context.Background())

	t := &mcpSSETransport{
		endpoint: endpoint,
		headers:  headers,
		client:   &http.Client{},
		msgCh:    make(chan json.RawMessage, 64),
		cancel:   cancel,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("SSE request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("SSE connect to %s: %w", endpoint, err)
	}

	// Read SSE stream in background
	go t.readSSE(resp.Body)

	return t, nil
}

func (t *mcpSSETransport) readSSE(body io.ReadCloser) {
	defer body.Close()
	scanner := bufio.NewScanner(body)
	var eventType string
	var data strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event
			if data.Len() > 0 {
				content := data.String()
				if eventType == "endpoint" {
					// Parse the POST URL from the endpoint event
					t.mu.Lock()
					// Handle relative URLs
					if strings.HasPrefix(content, "/") {
						// Extract base URL from endpoint
						base := t.endpoint
						if idx := strings.Index(base, "//"); idx >= 0 {
							if slashIdx := strings.Index(base[idx+2:], "/"); slashIdx >= 0 {
								base = base[:idx+2+slashIdx]
							}
						}
						t.postURL = base + content
					} else {
						t.postURL = content
					}
					t.mu.Unlock()
				} else if eventType == "message" || eventType == "" {
					t.msgCh <- json.RawMessage(content)
				}
				data.Reset()
				eventType = ""
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
}

func (t *mcpSSETransport) Send(msg json.RawMessage) error {
	t.mu.Lock()
	postURL := t.postURL
	t.mu.Unlock()

	if postURL == "" {
		return fmt.Errorf("SSE transport: POST URL not yet received from server")
	}

	req, err := http.NewRequest("POST", postURL, strings.NewReader(string(msg)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("SSE POST: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("SSE POST returned %d", resp.StatusCode)
	}
	return nil
}

func (t *mcpSSETransport) Receive() (json.RawMessage, error) {
	msg, ok := <-t.msgCh
	if !ok {
		return nil, fmt.Errorf("SSE connection closed")
	}
	return msg, nil
}

func (t *mcpSSETransport) Close() error {
	t.cancel()
	return nil
}

// --- MCPClient ---

// NewMCPClient creates a new MCP client from configuration.
func NewMCPClient(config MCPConfig) *MCPClient {
	return &MCPClient{config: config}
}

// Connect establishes the transport and performs the MCP handshake.
func (c *MCPClient) Connect() error {
	var transport mcpTransport
	var err error

	switch c.config.Transport {
	case "mcp", "stdio":
		transport, err = newStdioTransport(c.config.Command, c.config.Args, c.config.Env)
	case "sse":
		transport, err = newSSETransport(c.config.Endpoint, c.config.Headers)
	default:
		return fmt.Errorf("unsupported MCP transport: %q", c.config.Transport)
	}
	if err != nil {
		return err
	}
	c.transport = transport

	if err := c.initialize(); err != nil {
		transport.Close()
		return fmt.Errorf("MCP initialization failed for %q: %w", c.config.Name, err)
	}

	fmt.Fprintf(os.Stderr, "[haira] MCP connected: %s (%s)\n", c.config.Name, c.config.Transport)
	return nil
}

func (c *MCPClient) initialize() error {
	params, _ := json.Marshal(map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]string{
			"name":    "haira",
			"version": "1.0.0",
		},
	})

	result, err := c.call("initialize", params)
	if err != nil {
		return err
	}

	var initResult mcpInitializeResult
	json.Unmarshal(result, &initResult)
	fmt.Fprintf(os.Stderr, "[haira] MCP server: %s v%s (protocol %s)\n",
		initResult.ServerInfo.Name, initResult.ServerInfo.Version, initResult.ProtocolVersion)

	// Send initialized notification
	c.notify("notifications/initialized", nil)
	return nil
}

// call sends a JSON-RPC request and waits for the response.
func (c *MCPClient) call(method string, params json.RawMessage) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(req)

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.transport.Send(data); err != nil {
		return nil, fmt.Errorf("MCP send: %w", err)
	}

	// Read responses, skipping notifications until we get our response
	for {
		respData, err := c.transport.Receive()
		if err != nil {
			return nil, fmt.Errorf("MCP receive: %w", err)
		}

		var resp jsonrpcResponse
		if err := json.Unmarshal(respData, &resp); err != nil {
			// Could be a notification, skip it
			continue
		}

		// Skip notifications (no ID match)
		if resp.ID != id {
			continue
		}

		if resp.Error != nil {
			return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// notify sends a JSON-RPC notification (no ID, no response expected).
func (c *MCPClient) notify(method string, params json.RawMessage) {
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, _ := json.Marshal(req)
	c.transport.Send(data)
}

// ListTools discovers tools from the MCP server and returns them as ToolDefs.
func (c *MCPClient) ListTools() ([]*ToolDef, error) {
	result, err := c.call("tools/list", nil)
	if err != nil {
		return nil, err
	}

	var toolsResult mcpToolsListResult
	if err := json.Unmarshal(result, &toolsResult); err != nil {
		return nil, fmt.Errorf("MCP tools/list parse: %w", err)
	}

	var defs []*ToolDef
	for _, mt := range toolsResult.Tools {
		toolName := mt.Name
		client := c

		defs = append(defs, &ToolDef{
			Name:        mt.Name,
			Description: mt.Description,
			Parameters:  mt.InputSchema,
			Handler: func(args json.RawMessage) (any, error) {
				return client.CallTool(toolName, args)
			},
		})
	}

	c.tools = defs
	fmt.Fprintf(os.Stderr, "[haira] MCP %s: discovered %d tools\n", c.config.Name, len(defs))
	return defs, nil
}

// CallTool invokes a tool on the MCP server.
func (c *MCPClient) CallTool(name string, args json.RawMessage) (any, error) {
	params, _ := json.Marshal(map[string]any{
		"name":      name,
		"arguments": json.RawMessage(args),
	})

	result, err := c.call("tools/call", params)
	if err != nil {
		return nil, err
	}

	var callResult mcpToolCallResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		return nil, fmt.Errorf("MCP tools/call parse: %w", err)
	}

	if callResult.IsError {
		var errText string
		for _, c := range callResult.Content {
			if c.Type == "text" {
				errText += c.Text
			}
		}
		return nil, fmt.Errorf("%s", errText)
	}

	// Combine text content
	var texts []string
	for _, c := range callResult.Content {
		if c.Type == "text" {
			texts = append(texts, c.Text)
		}
	}
	if len(texts) == 1 {
		return texts[0], nil
	}
	if len(texts) > 1 {
		return strings.Join(texts, "\n"), nil
	}
	return nil, nil
}

// Close shuts down the MCP connection and subprocess.
func (c *MCPClient) Close() error {
	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}

// --- Global MCP Shutdown Registry ---

var (
	mcpClients   []*MCPClient
	mcpClientsMu sync.Mutex
)

// RegisterMCPClient adds a client to the global shutdown list.
func RegisterMCPClient(c *MCPClient) {
	mcpClientsMu.Lock()
	defer mcpClientsMu.Unlock()
	mcpClients = append(mcpClients, c)
}

// ShutdownMCP closes all registered MCP clients.
func ShutdownMCP() {
	mcpClientsMu.Lock()
	defer mcpClientsMu.Unlock()
	for _, c := range mcpClients {
		c.Close()
	}
	mcpClients = nil
}

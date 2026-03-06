package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Server is the Haira LSP server.
type Server struct {
	reader  *bufio.Reader
	writer  io.Writer
	mu      sync.Mutex
	handler *Handler
}

// NewServer creates a new LSP server reading from r and writing to w.
func NewServer(r io.Reader, w io.Writer) *Server {
	s := &Server{
		reader: bufio.NewReader(r),
		writer: w,
	}
	s.handler = NewHandler(s)
	return s
}

// Run starts the LSP server main loop. It blocks until stdin is closed.
func (s *Server) Run() error {
	for {
		msg, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		s.handleMessage(msg)
	}
}

// RunStdio starts the server on stdin/stdout.
func RunStdio() error {
	s := NewServer(os.Stdin, os.Stdout)
	return s.Run()
}

// readMessage reads one JSON-RPC message from the input stream.
// LSP uses "Content-Length: N\r\n\r\n" headers followed by N bytes of JSON.
func (s *Server) readMessage() (json.RawMessage, error) {
	contentLength := -1

	// Read headers
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			// Empty line = end of headers
			break
		}

		if strings.HasPrefix(line, "Content-Length: ") {
			val := strings.TrimPrefix(line, "Content-Length: ")
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("bad Content-Length: %q", val)
			}
			contentLength = n
		}
		// Ignore other headers (Content-Type, etc.)
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	// Read body
	body := make([]byte, contentLength)
	_, err := io.ReadFull(s.reader, body)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(body), nil
}

// handleMessage dispatches a JSON-RPC message to the handler.
func (s *Server) handleMessage(msg json.RawMessage) {
	// Parse the method and id
	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id,omitempty"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}
	if err := json.Unmarshal(msg, &req); err != nil {
		s.logf("failed to parse message: %s", err)
		return
	}

	isNotification := req.ID == nil || string(req.ID) == "null"

	switch req.Method {
	case "initialize":
		result := s.handler.Initialize(req.Params)
		if !isNotification {
			s.sendResult(req.ID, result)
		}

	case "initialized":
		// Client acknowledges — no response needed

	case "shutdown":
		if !isNotification {
			s.sendResult(req.ID, nil)
		}

	case "exit":
		os.Exit(0)

	case "textDocument/didOpen":
		s.handler.DidOpen(req.Params)

	case "textDocument/didChange":
		s.handler.DidChange(req.Params)

	case "textDocument/didClose":
		s.handler.DidClose(req.Params)

	case "textDocument/didSave":
		s.handler.DidSave(req.Params)

	case "textDocument/hover":
		result := s.handler.Hover(req.Params)
		if !isNotification {
			s.sendResult(req.ID, result)
		}

	case "textDocument/completion":
		result := s.handler.Completion(req.Params)
		if !isNotification {
			s.sendResult(req.ID, result)
		}

	case "textDocument/definition":
		result := s.handler.Definition(req.Params)
		if !isNotification {
			s.sendResult(req.ID, result)
		}

	case "textDocument/documentSymbol":
		result := s.handler.DocumentSymbol(req.Params)
		if !isNotification {
			s.sendResult(req.ID, result)
		}

	case "textDocument/signatureHelp":
		result := s.handler.SignatureHelp(req.Params)
		if !isNotification {
			s.sendResult(req.ID, result)
		}

	case "textDocument/formatting":
		result := s.handler.Formatting(req.Params)
		if !isNotification {
			s.sendResult(req.ID, result)
		}

	default:
		if !isNotification {
			// Return null for unimplemented methods instead of an error,
			// so clients don't log deserialization failures.
			s.sendResult(req.ID, nil)
		}
	}
}

// sendResult sends a successful JSON-RPC response.
func (s *Server) sendResult(id json.RawMessage, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeMessage(resp)
}

// sendError sends a JSON-RPC error response.
func (s *Server) sendError(id json.RawMessage, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &ResponseError{Code: code, Message: message},
	}
	s.writeMessage(resp)
}

// SendNotification sends a JSON-RPC notification (no id).
func (s *Server) SendNotification(method string, params interface{}) {
	notif := Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	s.writeMessage(notif)
}

// writeMessage writes a JSON-RPC message with LSP framing.
func (s *Server) writeMessage(msg interface{}) {
	body, err := json.Marshal(msg)
	if err != nil {
		s.logf("marshal error: %s", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := s.writer.Write([]byte(header)); err != nil {
		s.logf("write header error: %s", err)
		return
	}
	if _, err := s.writer.Write(body); err != nil {
		s.logf("write body error: %s", err)
	}
}

func (s *Server) logf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[haira-lsp] "+format+"\n", args...)
}

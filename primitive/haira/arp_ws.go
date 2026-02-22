package haira

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// ARP WebSocket Transport Binding (Section 13.2 of the ARP Spec)
// ---------------------------------------------------------------------------

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS handled at server level
	},
}

// handleArpWebSocket upgrades an HTTP connection to WebSocket and runs the
// ARP Minimal Mode protocol for bidirectional agent-renderer communication.
func (s *Server) handleArpWebSocket(rw http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("arp: websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Serializes writes — gorilla/websocket requires single writer at a time
	var writeMu sync.Mutex
	writeJSON := func(v any) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteJSON(v)
	}

	// 1. Send ArpHello with capabilities
	hello := ArpServerCapabilities()
	if err := writeJSON(hello); err != nil {
		return
	}

	// 2. Read loop — process client messages
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break // Client disconnected
		}

		var msg ArpInputMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			writeJSON(ArpMessage{V: 1, Type: "error", Payload: mustMarshal(map[string]string{
				"error": "invalid message format",
			})})
			continue
		}

		switch msg.Type {
		case "input":
			s.handleArpInput(msg, writeJSON)
		default:
			writeJSON(ArpMessage{V: 1, Type: "error", Payload: mustMarshal(map[string]string{
				"error": "unknown message type: " + msg.Type,
			})})
		}
	}
}

// handleArpInput processes an input message from the ARP client.
// It finds the appropriate streaming workflow, executes it, and streams
// ARP messages back through the write function.
func (s *Server) handleArpInput(msg ArpInputMessage, writeJSON func(any) error) {
	// Extract text from input data
	var text string
	switch msg.InputType {
	case "text":
		var d struct {
			Text string `json:"text"`
		}
		json.Unmarshal(msg.Data, &d)
		text = d.Text
	case "form_submit":
		// Forward form data as JSON text to the agent
		text = string(msg.Data)
	case "action":
		var d struct {
			Action  string `json:"action"`
			Payload any    `json:"payload,omitempty"`
		}
		json.Unmarshal(msg.Data, &d)
		actionJSON, _ := json.Marshal(d)
		text = "[Action: " + d.Action + "] " + string(actionJSON)
	default:
		text = string(msg.Data)
	}

	if text == "" {
		writeJSON(ArpMessage{V: 1, Type: "error", SessionID: msg.SessionID,
			Payload: mustMarshal(map[string]string{"error": "empty input"})})
		return
	}

	// Find the streaming workflow for this session
	wf := s.findStreamWorkflow(msg.SessionID)
	if wf == nil {
		writeJSON(ArpMessage{V: 1, Type: "error", SessionID: msg.SessionID,
			Payload: mustMarshal(map[string]string{"error": "no streaming workflow found"})})
		return
	}

	// Ensure session exists and record user message
	sessionID := msg.SessionID
	if sessionID != "" && globalStore != nil {
		globalStore.EnsureSession(sessionID, wf.Name, wf.Path, "")
		globalStore.AddMessage(sessionID, "user", text, nil)
	}

	// Build params for the workflow
	params := map[string]any{
		"session_id": sessionID,
	}
	// Find the chat param name and set the message
	chatParam := findChatParam(wf.Params)
	params[chatParam] = text

	// Execute the workflow
	ch, err := wf.StreamHandler(params)
	if err != nil {
		writeJSON(ArpMessage{V: 1, Type: "error", SessionID: sessionID,
			Payload: mustMarshal(map[string]string{"error": err.Error()})})
		return
	}

	// Bridge StreamChunks to ARP and write each message
	arpMessages := ArpBridge(sessionID, ch)

	var fullReply string
	var uiEvents []json.RawMessage

	for arpMsg := range arpMessages {
		// Collect for persistence
		switch arpMsg.Type {
		case "delta":
			var p struct {
				Delta string `json:"delta"`
			}
			json.Unmarshal(arpMsg.Payload, &p)
			fullReply += p.Delta
		case "render":
			// Persist in ToolRenderEvent format (tool, component, props)
			// so the client can restore sessions correctly.
			if len(arpMsg.Components) > 0 {
				comp := arpMsg.Components[0]
				data, _ := json.Marshal(map[string]any{
					"tool":      arpMsg.ToolName,
					"component": comp.Type,
					"props":     comp.Props,
				})
				uiEvents = append(uiEvents, data)
			}
		}

		// Send to client
		if err := writeJSON(arpMsg); err != nil {
			break // Client disconnected mid-stream
		}
	}

	// Persist assistant reply
	if sessionID != "" && globalStore != nil && (fullReply != "" || len(uiEvents) > 0) {
		globalStore.AddMessage(sessionID, "assistant", fullReply, uiEvents)
	}
}

// findStreamWorkflow finds the first streaming workflow, or the one matching
// the session's stored workflow path.
func (s *Server) findStreamWorkflow(sessionID string) *WorkflowDef {
	// Try to find by session's stored workflow path
	if sessionID != "" && globalStore != nil {
		if session, err := globalStore.GetSession(sessionID); err == nil && session != nil {
			for _, wf := range s.workflows {
				if wf.Path == session.WorkflowPath && wf.StreamHandler != nil {
					return wf
				}
			}
		}
	}

	// Fallback: use the first streaming workflow
	for _, wf := range s.workflows {
		if wf.StreamHandler != nil {
			return wf
		}
	}
	return nil
}

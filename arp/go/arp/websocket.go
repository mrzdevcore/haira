package arp

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS handled at server level
	},
}

// InputHandler is called when the ARP WebSocket receives an input message.
// The implementation should execute the agent/workflow and return a StreamChunk channel.
//
// Parameters:
//   - sessionID: the session identifier from the client
//   - inputType: "text", "action", or "form_submit"
//   - text: extracted text content from the input
//   - rawData: the raw JSON data field from the input message
type InputHandler func(sessionID, inputType, text string, rawData json.RawMessage) (<-chan StreamChunk, error)

// PersistenceCallbacks provides optional hooks for session persistence.
// If nil, no persistence is performed.
type PersistenceCallbacks struct {
	// OnUserMessage is called when a user sends a message.
	OnUserMessage func(sessionID, text string)
	// OnAssistantReply is called when the assistant finishes responding.
	OnAssistantReply func(sessionID, reply string, uiEvents []json.RawMessage)
}

// WebSocketHandler returns an http.HandlerFunc that upgrades connections to
// the ARP WebSocket protocol and dispatches input to the given handler.
func WebSocketHandler(caps ArpHello, handler InputHandler, persist *PersistenceCallbacks) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
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
		if err := writeJSON(caps); err != nil {
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
				writeJSON(ArpMessage{V: 1, Type: TypeError, Payload: mustMarshal(map[string]string{
					"error": "invalid message format",
				})})
				continue
			}

			switch msg.Type {
			case TypeInput:
				handleArpInput(msg, writeJSON, handler, persist)
			default:
				writeJSON(ArpMessage{V: 1, Type: TypeError, Payload: mustMarshal(map[string]string{
					"error": "unknown message type: " + msg.Type,
				})})
			}
		}
	}
}

// handleArpInput processes an input message from the ARP client.
func handleArpInput(
	msg ArpInputMessage,
	writeJSON func(any) error,
	handler InputHandler,
	persist *PersistenceCallbacks,
) {
	// Extract text from input data
	text := ExtractInputText(msg)

	if text == "" {
		writeJSON(ArpMessage{V: 1, Type: TypeError, SessionID: msg.SessionID,
			Payload: mustMarshal(map[string]string{"error": "empty input"})})
		return
	}

	// Persist user message
	if persist != nil && persist.OnUserMessage != nil {
		persist.OnUserMessage(msg.SessionID, text)
	}

	// Execute the handler
	ch, err := handler(msg.SessionID, msg.InputType, text, msg.Data)
	if err != nil {
		writeJSON(ArpMessage{V: 1, Type: TypeError, SessionID: msg.SessionID,
			Payload: mustMarshal(map[string]string{"error": err.Error()})})
		return
	}

	// Bridge StreamChunks to ARP and write each message
	arpMessages := ArpBridge(msg.SessionID, ch)

	var fullReply string
	var uiEvents []json.RawMessage

	for arpMsg := range arpMessages {
		// Collect for persistence
		switch arpMsg.Type {
		case TypeDelta:
			var p struct {
				Delta string `json:"delta"`
			}
			json.Unmarshal(arpMsg.Payload, &p)
			fullReply += p.Delta
		case TypeRender:
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
	if persist != nil && persist.OnAssistantReply != nil && (fullReply != "" || len(uiEvents) > 0) {
		persist.OnAssistantReply(msg.SessionID, fullReply, uiEvents)
	}
}

// ExtractInputText extracts the text content from an ARP input message.
func ExtractInputText(msg ArpInputMessage) string {
	switch msg.InputType {
	case "text":
		var d struct {
			Text string `json:"text"`
		}
		json.Unmarshal(msg.Data, &d)
		return d.Text
	case "form_submit":
		return string(msg.Data)
	case "action":
		var d struct {
			Action  string `json:"action"`
			Payload any    `json:"payload,omitempty"`
		}
		json.Unmarshal(msg.Data, &d)
		actionJSON, _ := json.Marshal(d)
		return "[Action: " + d.Action + "] " + string(actionJSON)
	default:
		return string(msg.Data)
	}
}

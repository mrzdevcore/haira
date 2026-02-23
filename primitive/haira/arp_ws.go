package haira

import (
	"encoding/json"
	"fmt"
	"net/http"

	"haira-go-runtime/arp"
)

// ---------------------------------------------------------------------------
// ARP WebSocket Transport Binding
//
// Delegates the WebSocket protocol (upgrade, hello, read loop, bridging,
// persistence) to arp.WebSocketHandler. This file provides the Haira-specific
// InputHandler closure that does workflow lookup and session management.
// ---------------------------------------------------------------------------

// arpWebSocketHandler builds an http.HandlerFunc that implements the ARP
// WebSocket protocol using the go/arp package.
func (s *Server) arpWebSocketHandler() http.HandlerFunc {
	caps := ArpServerCapabilities()

	// InputHandler: called by arp.WebSocketHandler when a client sends input.
	handler := arp.InputHandler(func(sessionID, inputType, text string, rawData json.RawMessage) (<-chan arp.StreamChunk, error) {
		// Find the streaming workflow for this session
		wf := s.findStreamWorkflow(sessionID)
		if wf == nil {
			return nil, fmt.Errorf("no streaming workflow found")
		}

		// Ensure session exists in the store
		if sessionID != "" && globalStore != nil {
			globalStore.EnsureSession(sessionID, wf.Name, wf.Path, "")
		}

		// Build params for the workflow
		params := map[string]any{
			"session_id": sessionID,
		}
		chatParam := findChatParam(wf.Params)
		params[chatParam] = text

		// Execute the workflow's stream handler
		ch, err := wf.StreamHandler(params)
		if err != nil {
			return nil, err
		}

		// Convert haira.StreamChunk -> arp.StreamChunk
		return toArpChunks(ch), nil
	})

	// PersistenceCallbacks: called by arp.WebSocketHandler to persist
	// user messages and assistant replies to the globalStore.
	persist := &arp.PersistenceCallbacks{
		OnUserMessage: func(sessionID, text string) {
			if sessionID != "" && globalStore != nil {
				globalStore.AddMessage(sessionID, "user", text, nil)
			}
		},
		OnAssistantReply: func(sessionID, reply string, uiEvents []json.RawMessage) {
			if sessionID != "" && globalStore != nil && (reply != "" || len(uiEvents) > 0) {
				globalStore.AddMessage(sessionID, "assistant", reply, uiEvents)
			}
		},
	}

	return arp.WebSocketHandler(caps, handler, persist)
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

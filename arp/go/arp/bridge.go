package arp

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// Component ID generation
// ---------------------------------------------------------------------------

var componentCounter uint64

func nextComponentID() string {
	n := atomic.AddUint64(&componentCounter, 1)
	return fmt.Sprintf("c_%d", n)
}

// ---------------------------------------------------------------------------
// Bridge: StreamChunk → ArpMessage
// ---------------------------------------------------------------------------

// ArpBridge consumes a StreamChunk channel and produces an ArpMessage channel
// in ARP Minimal Mode format.
//
// This is the single translation point between agent streaming output and the
// transport-agnostic ARP protocol. Transport bindings (SSE, WebSocket, stdio)
// read from the ArpMessage channel.
func ArpBridge(sessionID string, chunks <-chan StreamChunk) <-chan ArpMessage {
	out := make(chan ArpMessage, 64)
	go func() {
		defer close(out)

		for chunk := range chunks {
			if chunk.Done {
				// Error with delta — send as error before final commit
				if chunk.Delta != "" && strings.HasPrefix(chunk.Delta, "error:") {
					errMsg := strings.TrimPrefix(chunk.Delta, "error: ")
					out <- ArpMessage{
						V:         1,
						Type:      TypeError,
						SessionID: sessionID,
						Payload:   mustMarshal(map[string]string{"error": errMsg}),
					}
				}
				out <- ArpMessage{
					V:         1,
					Type:      TypeCommit,
					SessionID: sessionID,
					Final:     true,
				}
				break
			}

			switch chunk.Type {
			case "tool_start":
				out <- ArpMessage{
					V:         1,
					Type:      TypeToolStart,
					SessionID: sessionID,
					Payload:   mustMarshal(map[string]any{"tool": chunk.ToolName, "args": chunk.ToolArgs}),
				}

			case "tool_render":
				comp := ArpComponent{
					ID:    nextComponentID(),
					Type:  chunk.RenderComponent,
					Props: json.RawMessage(chunk.RenderProps),
				}
				out <- ArpMessage{
					V:          1,
					Type:       TypeRender,
					SessionID:  sessionID,
					Components: []ArpComponent{comp},
					ToolName:   chunk.ToolName,
				}

			case "tool_end":
				out <- ArpMessage{
					V:         1,
					Type:      TypeToolEnd,
					SessionID: sessionID,
					Payload:   mustMarshal(map[string]any{"tool": chunk.ToolName, "ok": chunk.ToolOK}),
				}

			default:
				// Text delta
				if chunk.Delta != "" {
					out <- ArpMessage{
						V:         1,
						Type:      TypeDelta,
						SessionID: sessionID,
						Payload:   mustMarshal(map[string]string{"delta": chunk.Delta}),
					}
				}
			}
		}
	}()
	return out
}

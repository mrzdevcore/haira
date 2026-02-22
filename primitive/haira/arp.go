package haira

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// ARP Protocol Types (Minimal Mode — Section 16.3 of the ARP Spec)
// ---------------------------------------------------------------------------

// ArpMessage is the logical message envelope for ARP Minimal Mode.
// In Minimal Mode, messages use flat session-based addressing instead of
// the full object model (ArpDisplay/ArpRegistry/bind).
type ArpMessage struct {
	V         uint             `json:"v"`
	Type      string           `json:"type"`
	SessionID string           `json:"session_id,omitempty"`
	Payload   json.RawMessage  `json:"payload,omitempty"`

	// Render fields (type="render")
	Components []ArpComponent  `json:"components,omitempty"`

	// Patch fields (type="patch")
	Ops []ArpPatchOp           `json:"ops,omitempty"`

	// Commit fields (type="commit")
	Final bool                 `json:"final,omitempty"`

	// ToolName is set on render/tool_start/tool_end messages to identify
	// which tool produced this event. Used by SSE transport for backward compat.
	ToolName string            `json:"tool_name,omitempty"`
}

// ArpComponent is a typed UI component with props and optional fallback.
type ArpComponent struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Version  uint            `json:"version,omitempty"`
	Props    json.RawMessage `json:"props"`
	Fallback *ArpComponent   `json:"fallback,omitempty"`
}

// ArpPatchOp is an incremental update operation on the component tree.
type ArpPatchOp struct {
	Op        string        `json:"op"`                  // update, insert, remove, replace, reorder
	Target    string        `json:"target,omitempty"`
	Path      string        `json:"path,omitempty"`
	Value     any           `json:"value,omitempty"`
	After     string        `json:"after,omitempty"`
	Component *ArpComponent `json:"component,omitempty"`
}

// ArpHello is the capabilities message sent to clients on connect.
type ArpHello struct {
	V            uint            `json:"v"`
	Type         string          `json:"type"` // always "hello"
	Capabilities ArpCapabilities `json:"capabilities"`
}

// ArpCapabilities describes what the server supports.
type ArpCapabilities struct {
	Components []string `json:"components"`
	Features   []string `json:"features"`
}

// ArpInputMessage is an input message from the client (renderer → agent).
type ArpInputMessage struct {
	V               uint            `json:"v"`
	Type            string          `json:"type"` // "input"
	SessionID       string          `json:"session_id"`
	SourceComponent string          `json:"source_component,omitempty"`
	InputType       string          `json:"input_type"`
	Data            json.RawMessage `json:"data"`
}

// ---------------------------------------------------------------------------
// Component ID generation
// ---------------------------------------------------------------------------

var arpComponentCounter uint64

func arpNextComponentID() string {
	n := atomic.AddUint64(&arpComponentCounter, 1)
	return fmt.Sprintf("c_%d", n)
}

// ---------------------------------------------------------------------------
// Bridge: StreamChunk → ArpMessage
// ---------------------------------------------------------------------------

// ArpBridge consumes a StreamChunk channel (from agent.Stream()) and produces
// an ArpMessage channel in ARP Minimal Mode format.
//
// This is the single translation point between Haira's internal agent output
// and the transport-agnostic ARP protocol. Transport bindings (SSE, WebSocket,
// stdio) read from the ArpMessage channel.
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
						Type:      "error",
						SessionID: sessionID,
						Payload:   mustMarshal(map[string]string{"error": errMsg}),
					}
				}
				out <- ArpMessage{
					V:         1,
					Type:      "commit",
					SessionID: sessionID,
					Final:     true,
				}
				break
			}

			switch chunk.Type {
			case "tool_start":
				out <- ArpMessage{
					V:         1,
					Type:      "tool_start",
					SessionID: sessionID,
					Payload:   mustMarshal(map[string]any{"tool": chunk.ToolName, "args": chunk.ToolArgs}),
				}

			case "tool_render":
				comp := ArpComponent{
					ID:    arpNextComponentID(),
					Type:  chunk.RenderComponent,
					Props: json.RawMessage(chunk.RenderProps),
				}
				out <- ArpMessage{
					V:          1,
					Type:       "render",
					SessionID:  sessionID,
					Components: []ArpComponent{comp},
					ToolName:   chunk.ToolName,
				}

			case "tool_end":
				out <- ArpMessage{
					V:         1,
					Type:      "tool_end",
					SessionID: sessionID,
					Payload:   mustMarshal(map[string]any{"tool": chunk.ToolName, "ok": chunk.ToolOK}),
				}

			default:
				// Text delta
				if chunk.Delta != "" {
					out <- ArpMessage{
						V:         1,
						Type:      "delta",
						SessionID: sessionID,
						Payload:   mustMarshal(map[string]string{"delta": chunk.Delta}),
					}
				}
			}
		}
	}()
	return out
}

// ---------------------------------------------------------------------------
// SSE Transport Binding — writes ArpMessages as SSE text
// ---------------------------------------------------------------------------

// ArpSSEResult holds the collected output from an SSE stream for persistence.
type ArpSSEResult struct {
	FullReply string
	UIEvents  []json.RawMessage
}

// WriteArpSSE reads ArpMessages and writes them in the existing Haira SSE format.
// It returns the accumulated text reply and UI events for session persistence.
//
// This produces byte-identical output to the previous handleSSE() implementation,
// ensuring full backward compatibility with existing browser clients.
func WriteArpSSE(rw httpResponseWriter, flusher httpFlusher, messages <-chan ArpMessage) ArpSSEResult {
	var result ArpSSEResult

	for msg := range messages {
		switch msg.Type {
		case "tool_start":
			// Reformat to match existing SSE: {"tool": "...", "args": "..."}
			var p struct {
				Tool string `json:"tool"`
				Args string `json:"args"`
			}
			json.Unmarshal(msg.Payload, &p)
			data, _ := json.Marshal(map[string]any{"tool": p.Tool, "args": p.Args})
			fmt.Fprintf(rw, "event: tool_start\ndata: %s\n\n", data)

		case "render":
			if len(msg.Components) > 0 {
				comp := msg.Components[0]
				data, _ := json.Marshal(map[string]any{
					"tool":      msg.ToolName,
					"component": comp.Type,
					"props":     comp.Props,
				})
				fmt.Fprintf(rw, "event: tool_render\ndata: %s\n\n", data)
				result.UIEvents = append(result.UIEvents, data)
			}

		case "tool_end":
			var p struct {
				Tool string `json:"tool"`
				OK   bool   `json:"ok"`
			}
			json.Unmarshal(msg.Payload, &p)
			data, _ := json.Marshal(map[string]any{"tool": p.Tool, "ok": p.OK})
			fmt.Fprintf(rw, "event: tool_end\ndata: %s\n\n", data)

		case "delta":
			var p struct {
				Delta string `json:"delta"`
			}
			json.Unmarshal(msg.Payload, &p)
			if p.Delta != "" {
				result.FullReply += p.Delta
				data, _ := json.Marshal(map[string]string{"delta": p.Delta})
				fmt.Fprintf(rw, "data: %s\n\n", data)
			}

		case "error":
			var p struct {
				Error string `json:"error"`
			}
			json.Unmarshal(msg.Payload, &p)
			errData, _ := json.Marshal(map[string]string{"error": p.Error})
			fmt.Fprintf(rw, "event: error\ndata: %s\n\n", errData)

		case "commit":
			if msg.Final {
				fmt.Fprintf(rw, "data: [DONE]\n\n")
			}
		}
		flusher.Flush()
	}

	return result
}

// Interfaces to avoid importing net/http in this file.
// server.go passes http.ResponseWriter and http.Flusher.
type httpResponseWriter interface {
	Write([]byte) (int, error)
}

type httpFlusher interface {
	Flush()
}

// ---------------------------------------------------------------------------
// Capabilities
// ---------------------------------------------------------------------------

// knownUIComponents is the list of built-in ARP-compatible component names.
var knownUIComponents = []string{
	"text",
	"status-card",
	"table",
	"code-block",
	"diff",
	"key-value",
	"progress",
	"chart",
	"form",
	"confirm",
	"choices",
	"product-cards",
	"markdown",
	"image",
}

// ArpServerCapabilities builds the ArpHello for this server's capabilities.
func ArpServerCapabilities() ArpHello {
	return ArpHello{
		V:    1,
		Type: "hello",
		Capabilities: ArpCapabilities{
			Components: knownUIComponents,
			Features:   []string{"streaming", "input"},
		},
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

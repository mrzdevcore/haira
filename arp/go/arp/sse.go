package arp

import (
	"encoding/json"
	"fmt"
	"io"
)

// ArpSSEResult holds the collected output from an SSE stream for persistence.
type ArpSSEResult struct {
	FullReply string
	UIEvents  []json.RawMessage
}

// WriteArpSSE reads ArpMessages and writes them as Server-Sent Events.
// It returns the accumulated text reply and UI events for session persistence.
//
// The writer should be an http.ResponseWriter and flush should call
// http.Flusher.Flush() to push data to the client.
func WriteArpSSE(w io.Writer, flush func(), messages <-chan ArpMessage) ArpSSEResult {
	var result ArpSSEResult

	for msg := range messages {
		switch msg.Type {
		case TypeToolStart:
			var p struct {
				Tool string `json:"tool"`
				Args string `json:"args"`
			}
			json.Unmarshal(msg.Payload, &p)
			data, _ := json.Marshal(map[string]any{"tool": p.Tool, "args": p.Args})
			fmt.Fprintf(w, "event: tool_start\ndata: %s\n\n", data)

		case TypeRender:
			if len(msg.Components) > 0 {
				comp := msg.Components[0]
				data, _ := json.Marshal(map[string]any{
					"tool":      msg.ToolName,
					"component": comp.Type,
					"props":     comp.Props,
				})
				fmt.Fprintf(w, "event: tool_render\ndata: %s\n\n", data)
				result.UIEvents = append(result.UIEvents, data)
			}

		case TypeToolEnd:
			var p struct {
				Tool string `json:"tool"`
				OK   bool   `json:"ok"`
			}
			json.Unmarshal(msg.Payload, &p)
			data, _ := json.Marshal(map[string]any{"tool": p.Tool, "ok": p.OK})
			fmt.Fprintf(w, "event: tool_end\ndata: %s\n\n", data)

		case TypeDelta:
			var p struct {
				Delta string `json:"delta"`
			}
			json.Unmarshal(msg.Payload, &p)
			if p.Delta != "" {
				result.FullReply += p.Delta
				data, _ := json.Marshal(map[string]string{"delta": p.Delta})
				fmt.Fprintf(w, "event: delta\ndata: %s\n\n", data)
			}

		case TypeError:
			var p struct {
				Error string `json:"error"`
			}
			json.Unmarshal(msg.Payload, &p)
			errData, _ := json.Marshal(map[string]string{"error": p.Error})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", errData)

		case TypeCommit:
			if msg.Final {
				fmt.Fprintf(w, "event: done\ndata: [DONE]\n\n")
			}
		}
		flush()
	}

	return result
}

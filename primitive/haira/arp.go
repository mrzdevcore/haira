package haira

import (
	"haira-go-runtime/arp"
)

// ---------------------------------------------------------------------------
// ARP Protocol Type Aliases
//
// These are Go type aliases (using =), meaning ArpMessage and arp.ArpMessage
// are the exact same type. No conversion is needed when passing values between
// the haira and arp packages.
// ---------------------------------------------------------------------------

type ArpMessage = arp.ArpMessage
type ArpComponent = arp.ArpComponent
type ArpPatchOp = arp.ArpPatchOp
type ArpHello = arp.ArpHello
type ArpCapabilities = arp.ArpCapabilities
type ArpInputMessage = arp.ArpInputMessage
type ArpSSEResult = arp.ArpSSEResult

// ---------------------------------------------------------------------------
// StreamChunk Adapter
// ---------------------------------------------------------------------------

// toArpChunks converts a haira.StreamChunk channel to an arp.StreamChunk
// channel. The two types have identical fields but live in different packages.
func toArpChunks(in <-chan StreamChunk) <-chan arp.StreamChunk {
	out := make(chan arp.StreamChunk, 64)
	go func() {
		defer close(out)
		for chunk := range in {
			out <- arp.StreamChunk{
				Delta:           chunk.Delta,
				Done:            chunk.Done,
				Type:            chunk.Type,
				ToolName:        chunk.ToolName,
				ToolArgs:        chunk.ToolArgs,
				ToolOK:          chunk.ToolOK,
				RenderComponent: chunk.RenderComponent,
				RenderProps:     chunk.RenderProps,
			}
		}
	}()
	return out
}

// ---------------------------------------------------------------------------
// Bridge: StreamChunk -> ArpMessage (delegates to arp.ArpBridge)
// ---------------------------------------------------------------------------

// ArpBridge consumes a haira.StreamChunk channel and produces an ArpMessage
// channel in ARP Minimal Mode format.
func ArpBridge(sessionID string, chunks <-chan StreamChunk) <-chan ArpMessage {
	return arp.ArpBridge(sessionID, toArpChunks(chunks))
}

// ---------------------------------------------------------------------------
// SSE Transport (delegates to arp.WriteArpSSE)
// ---------------------------------------------------------------------------

// Interfaces to avoid importing net/http in this file.
// server.go passes http.ResponseWriter and http.Flusher.
type httpResponseWriter interface {
	Write([]byte) (int, error)
}

type httpFlusher interface {
	Flush()
}

// WriteArpSSE reads ArpMessages and writes them as SSE events.
func WriteArpSSE(rw httpResponseWriter, flusher httpFlusher, messages <-chan ArpMessage) ArpSSEResult {
	return arp.WriteArpSSE(rw, flusher.Flush, messages)
}

// ---------------------------------------------------------------------------
// Capabilities (delegates to arp.DefaultCapabilities)
// ---------------------------------------------------------------------------

// ArpServerCapabilities builds the ArpHello for this server.
func ArpServerCapabilities() ArpHello {
	return arp.DefaultCapabilities()
}

// Package arp implements the Agent Rendering Protocol (ARP) — a transport-agnostic
// protocol for bidirectional communication between agent backends and frontend renderers.
package arp

import "encoding/json"

// ---------------------------------------------------------------------------
// Protocol Messages (server → client)
// ---------------------------------------------------------------------------

// ArpMessage is the logical message envelope for ARP Minimal Mode.
// In Minimal Mode, messages use flat session-based addressing instead of
// the full object model (ArpDisplay/ArpRegistry/bind).
type ArpMessage struct {
	V         uint            `json:"v"`
	Type      string          `json:"type"`
	SessionID string          `json:"session_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`

	// Render fields (type="render")
	Components []ArpComponent `json:"components,omitempty"`

	// Patch fields (type="patch")
	Ops []ArpPatchOp `json:"ops,omitempty"`

	// Commit fields (type="commit")
	Final bool `json:"final,omitempty"`

	// ToolName is set on render/tool_start/tool_end messages to identify
	// which tool produced this event.
	ToolName string `json:"tool_name,omitempty"`
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

// ---------------------------------------------------------------------------
// Protocol Messages (client → server)
// ---------------------------------------------------------------------------

// ArpInputMessage is an input message from the client (renderer → agent).
type ArpInputMessage struct {
	V               uint            `json:"v"`
	Type            string          `json:"type"` // "input"
	SessionID       string          `json:"session_id"`
	SourceComponent string          `json:"source_component,omitempty"`
	InputType       string          `json:"input_type"` // "text", "action", "form_submit"
	Data            json.RawMessage `json:"data"`
}

// ---------------------------------------------------------------------------
// Message type constants
// ---------------------------------------------------------------------------

const (
	TypeHello     = "hello"
	TypeDelta     = "delta"
	TypeToolStart = "tool_start"
	TypeToolEnd   = "tool_end"
	TypeRender    = "render"
	TypePatch     = "patch"
	TypeError     = "error"
	TypeCommit    = "commit"
	TypeInput     = "input"
)

package arp

// StreamChunk represents a piece of streaming agent output.
// This is the universal input type for the ARP bridge — any agent framework
// can produce these to communicate with ARP frontends.
type StreamChunk struct {
	Delta string // Incremental text (for Type="" or "delta")
	Done  bool   // True when stream is complete

	// Event type: "" or "delta" for text, "tool_start", "tool_end", "tool_render"
	Type string

	// Tool-related fields (only set when Type is "tool_start", "tool_end", or "tool_render")
	ToolName string
	ToolArgs string // JSON args (tool_start only)
	ToolOK   bool   // Success status (tool_end only)

	// Render fields (tool_render only)
	RenderComponent string // Component name ("status-card", "table", etc.)
	RenderProps     string // JSON-serialized component props
}

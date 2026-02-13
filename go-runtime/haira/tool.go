package haira

import "encoding/json"

// ToolHandler is a function that handles a tool call.
// It receives JSON-encoded arguments and returns a result or error.
type ToolHandler func(args json.RawMessage) (any, error)

// ToolDef defines a tool that an agent can use.
type ToolDef struct {
	Name        string
	Description string
	Parameters  json.RawMessage // JSON Schema for parameters
	Handler     ToolHandler
}

// ToolRegistry holds registered tools for an agent.
type ToolRegistry struct {
	tools map[string]*ToolDef
}

// NewToolRegistry creates a new empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]*ToolDef)}
}

// Register adds a tool to the registry.
func (r *ToolRegistry) Register(tool *ToolDef) {
	r.tools[tool.Name] = tool
}

// Get returns a tool by name.
func (r *ToolRegistry) Get(name string) (*ToolDef, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns all registered tools.
func (r *ToolRegistry) All() []*ToolDef {
	result := make([]*ToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

package haira

// WorkflowDef defines a workflow that can be exposed as an HTTP endpoint.
type WorkflowDef struct {
	Name          string
	Method        string // HTTP method (GET, POST, etc.)
	Path          string // URL path
	Handler       func(params map[string]any) (any, error)
	StreamHandler func(params map[string]any) (<-chan StreamChunk, error)
}

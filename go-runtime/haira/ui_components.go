package haira

import "encoding/json"

// UiNode is the interface for all renderable UI components.
type UiNode interface {
	UiComponentName() string
}

// --- StatusCard ---

type UiStatusCard struct {
	Status   string `json:"status"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Sections []any  `json:"sections,omitempty"`
}

func (UiStatusCard) UiComponentName() string { return "status-card" }

type UiSection struct {
	Label   string `json:"label"`
	Content string `json:"content"`
	Style   string `json:"style"`
}

// --- Table ---

type UiTable struct {
	Title     string `json:"title"`
	Headers   []any  `json:"headers"`
	Rows      []any  `json:"rows"`
	Highlight []any  `json:"highlight,omitempty"`
}

func (UiTable) UiComponentName() string { return "table" }

// --- CodeBlock ---

type UiCodeBlock struct {
	Title    string `json:"title"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

func (UiCodeBlock) UiComponentName() string { return "code-block" }

// --- Diff ---

type UiDiff struct {
	Title       string `json:"title"`
	BeforeLabel string `json:"before_label"`
	AfterLabel  string `json:"after_label"`
	Before      string `json:"before"`
	After       string `json:"after"`
	Language    string `json:"language"`
}

func (UiDiff) UiComponentName() string { return "diff" }

// --- KeyValue ---

type UiKeyValue struct {
	Title string `json:"title"`
	Items []any  `json:"items"`
}

func (UiKeyValue) UiComponentName() string { return "key-value" }

type UiKVItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Style string `json:"style"`
}

// --- Progress ---

type UiProgress struct {
	Title string `json:"title"`
	Steps []any  `json:"steps"`
}

func (UiProgress) UiComponentName() string { return "progress" }

type UiStep struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// --- Form ---

type UiForm struct {
	Title        string `json:"title"`
	Fields       []any  `json:"fields"`
	SubmitLabel  string `json:"submit_label"`
	SubmitAction string `json:"submit_action"`
}

func (UiForm) UiComponentName() string { return "form" }

type UiField struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	FieldType string `json:"field_type"`
	Value     string `json:"value"`
	Options   []any  `json:"options,omitempty"`
	Required  bool   `json:"required"`
}

// --- Confirm (yes/no approval) ---

type UiConfirm struct {
	Title        string `json:"title"`
	Message      string `json:"message,omitempty"`
	ConfirmLabel string `json:"confirm_label"`
	DenyLabel    string `json:"deny_label"`
}

func (UiConfirm) UiComponentName() string { return "confirm" }

// --- Choices (option picker) ---

type UiChoices struct {
	Title   string `json:"title"`
	Options []any  `json:"options"`
	Style   string `json:"style,omitempty"`
}

func (UiChoices) UiComponentName() string { return "choices" }

// --- Group (composition) ---

type UiGroup struct {
	Children []any `json:"-"`
}

func (UiGroup) UiComponentName() string { return "group" }

// MarshalJSON wraps each UiNode child into {component, props} for the frontend.
func (g UiGroup) MarshalJSON() ([]byte, error) {
	type wrappedChild struct {
		Component string `json:"component"`
		Props     any    `json:"props"`
	}
	wrapped := make([]wrappedChild, 0, len(g.Children))
	for _, child := range g.Children {
		if node, ok := child.(UiNode); ok {
			wrapped = append(wrapped, wrappedChild{
				Component: node.UiComponentName(),
				Props:     child,
			})
		}
	}
	return json.Marshal(map[string]any{"children": wrapped})
}

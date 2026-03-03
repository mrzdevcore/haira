package haira

import (
	"encoding/json"
	"fmt"
)

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
	Tabs      []any  `json:"tabs,omitempty"`
}

func (UiTable) UiComponentName() string { return "table" }

type UiTab struct {
	Name      string `json:"name"`
	Headers   []any  `json:"headers"`
	Rows      []any  `json:"rows"`
	Highlight []any  `json:"highlight,omitempty"`
}

// --- CodeBlock ---

type UiCodeBlock struct {
	Title    string `json:"title"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Tabs     []any  `json:"tabs,omitempty"`
}

func (UiCodeBlock) UiComponentName() string { return "code-block" }

type UiCodeTab struct {
	Name     string `json:"name"`
	Language string `json:"language"`
	Code     string `json:"code"`
}

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

// --- Chart ---

type UiChartDataset struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
	Color string    `json:"color,omitempty"`
}

type UiChart struct {
	Type     string  `json:"type"`
	Title    string  `json:"title,omitempty"`
	Labels   []any   `json:"labels,omitempty"`
	Datasets []any   `json:"datasets"`
	Height   float64 `json:"height,omitempty"`
}

func (UiChart) UiComponentName() string { return "chart" }

// --- ProductCards (image card grid) ---

type UiProductCards struct {
	Title string `json:"title"`
	Cards []any  `json:"cards"`
}

func (UiProductCards) UiComponentName() string { return "product-cards" }

type UiProductCardItem struct {
	Name        string `json:"name"`
	Price       string `json:"price"`
	Image       string `json:"image,omitempty"`
	Brand       string `json:"brand,omitempty"`
	Description string `json:"description,omitempty"`
	Badge       string `json:"badge,omitempty"`
	URL         string `json:"url,omitempty"`
}

// --- Markdown ---

type UiMarkdown struct {
	Content string `json:"content"`
	Title   string `json:"title,omitempty"`
}

func (UiMarkdown) UiComponentName() string { return "markdown" }

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

// uiNodeSummary returns a compact text summary of a UiNode for the LLM.
// The full payload is already sent to the frontend via tool_render,
// so the LLM only needs enough context to decide the next step.
func uiNodeSummary(node UiNode, raw any) string {
	switch v := raw.(type) {
	case UiStatusCard:
		if v.Message != "" {
			return fmt.Sprintf("[%s] %s: %s", v.Status, v.Title, v.Message)
		}
		return fmt.Sprintf("[%s] %s", v.Status, v.Title)
	case UiTable:
		rows := len(v.Rows)
		tabs := len(v.Tabs)
		if tabs > 0 {
			return fmt.Sprintf("Rendered table %q with %d tabs", v.Title, tabs)
		}
		return fmt.Sprintf("Rendered table %q with %d rows", v.Title, rows)
	case UiCodeBlock:
		tabs := len(v.Tabs)
		if tabs > 0 {
			return fmt.Sprintf("Rendered code block %q with %d tabs", v.Title, tabs)
		}
		return fmt.Sprintf("Rendered code block %q (%d chars)", v.Title, len(v.Code))
	case UiConfirm:
		return fmt.Sprintf("Rendered confirmation: %s (options: %s / %s)", v.Title, v.ConfirmLabel, v.DenyLabel)
	case UiChoices:
		return fmt.Sprintf("Rendered choices: %s (%d options)", v.Title, len(v.Options))
	case UiGroup:
		parts := make([]string, 0, len(v.Children))
		for _, child := range v.Children {
			if cn, ok := child.(UiNode); ok {
				parts = append(parts, uiNodeSummary(cn, child))
			}
		}
		summary := "Rendered group:"
		for _, p := range parts {
			summary += "\n- " + p
		}
		return summary
	case UiKeyValue:
		return fmt.Sprintf("Rendered key-value card %q with %d items", v.Title, len(v.Items))
	case UiProgress:
		return fmt.Sprintf("Rendered progress %q with %d steps", v.Title, len(v.Steps))
	case UiDiff:
		return fmt.Sprintf("Rendered diff %q", v.Title)
	case UiForm:
		return fmt.Sprintf("Rendered form %q with %d fields", v.Title, len(v.Fields))
	case UiChart:
		return fmt.Sprintf("Rendered %s chart %q with %d datasets", v.Type, v.Title, len(v.Datasets))
	case UiProductCards:
		return fmt.Sprintf("Rendered %d product cards: %q", len(v.Cards), v.Title)
	case UiMarkdown:
		n := len(v.Content)
		if n > 80 {
			n = 80
		}
		if v.Title != "" {
			return fmt.Sprintf("Rendered markdown %q (%d chars)", v.Title, len(v.Content))
		}
		return fmt.Sprintf("Rendered markdown (%d chars)", len(v.Content))
	default:
		return fmt.Sprintf("Rendered %s component", node.UiComponentName())
	}
}

// --- Constructors for building UI nodes from Haira tool code ---

// UiNewKeyValue builds a UiKeyValue from a title and a map of key-value pairs.
// Usage in Haira: ui.key_value("Title", {"Key1": "val1", "Key2": "val2"})
// Optionally accepts a third argument: a map of key → style ("success", "error", "warning", "code").
func UiNewKeyValue(title string, data map[string]any, styles ...map[string]any) UiKeyValue {
	var items []any
	for k, v := range data {
		style := ""
		if len(styles) > 0 {
			if s, ok := styles[0][k]; ok {
				style = Str(s)
			}
		}
		items = append(items, UiKVItem{Key: k, Value: Str(v), Style: style})
	}
	return UiKeyValue{Title: title, Items: items}
}

// UiNewStatusCard builds a UiStatusCard.
// Usage in Haira: ui.status_card("success", "Title", "Message")
func UiNewStatusCard(status, title string, message ...string) UiStatusCard {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return UiStatusCard{Status: status, Title: title, Message: msg}
}

// UiNewGroup builds a UiGroup from multiple UiNode children.
// Usage in Haira: ui.group(node1, node2, ...)
func UiNewGroup(children ...any) UiGroup {
	return UiGroup{Children: children}
}

// UiNewConfirm builds a UiConfirm dialog.
// Usage in Haira: ui.confirm("Title", "Confirm", "Cancel")
// Optional 4th argument: message string.
func UiNewConfirm(title, confirmLabel, denyLabel string, message ...string) UiConfirm {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return UiConfirm{Title: title, ConfirmLabel: confirmLabel, DenyLabel: denyLabel, Message: msg}
}

// UiNewChart builds a UiChart from type, labels, and datasets.
// Usage in Haira: ui.chart("bar", "Revenue", ["Q1","Q2","Q3"], [dataset1, dataset2])
func UiNewChart(chartType string, title string, labels any, datasets any, height ...float64) UiChart {
	h := float64(0)
	if len(height) > 0 {
		h = height[0]
	}
	return UiChart{Type: chartType, Title: title, Labels: toAnySlice(labels), Datasets: toAnySlice(datasets), Height: h}
}

// UiNewTable builds a UiTable from a title, headers, and rows.
// Usage in Haira: ui.table("Title", ["col1","col2"], [["a","b"],["c","d"]])
func UiNewTable(title string, headers any, rows any) UiTable {
	return UiTable{Title: title, Headers: toAnySlice(headers), Rows: toAnySlice(rows)}
}

// UiNewProductCards builds a UiProductCards grid.
// Usage in Haira: ui.product_cards("Title", cards)
func UiNewProductCards(title string, cards any) UiProductCards {
	return UiProductCards{Title: title, Cards: toAnySlice(cards)}
}

// UiNewMarkdown builds a UiMarkdown component for rendering rich text.
// Usage in Haira: ui.markdown(content) or ui.markdown(content, "Title")
func UiNewMarkdown(content string, title ...string) UiMarkdown {
	t := ""
	if len(title) > 0 {
		t = title[0]
	}
	return UiMarkdown{Content: content, Title: t}
}

// toAnySlice coerces a value to []any. Accepts []any directly or any as a pass-through.
func toAnySlice(v any) []any {
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		return s
	}
	// Single value — wrap in slice
	return []any{v}
}

package haira

import "encoding/json"

// uiToolSpec defines a synthetic UI tool.
type uiToolSpec struct {
	component   string
	toolName    string
	description string
	schema      string
	handler     ToolHandler
}

func (s *uiToolSpec) toToolDef() *ToolDef {
	return &ToolDef{
		Name:        s.toolName,
		Description: s.description,
		Parameters:  json.RawMessage(s.schema),
		Handler:     s.handler,
	}
}

var uiToolSpecs = []uiToolSpec{
	{
		component:   "StatusCard",
		toolName:    "render_statuscard",
		description: "Display a status card with an icon. Use for showing success, error, warning, or info messages to the user.",
		schema:      `{"type":"object","properties":{"status":{"type":"string","enum":["success","error","warning","info"],"description":"Status type that determines the icon and color"},"title":{"type":"string","description":"Bold title text"},"message":{"type":"string","description":"Message body text"}},"required":["status","title"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Status  string `json:"status"`
				Title   string `json:"title"`
				Message string `json:"message"`
			}
			json.Unmarshal(args, &p)
			return UiStatusCard{Status: p.Status, Title: p.Title, Message: p.Message}, nil
		},
	},
	{
		component:   "Confirm",
		toolName:    "render_confirm",
		description: "Show a confirmation dialog with two buttons (e.g. approve/deny, yes/no). The user clicks a button and their choice is sent as their next chat message. Use for binary decisions that need explicit user approval.",
		schema:      `{"type":"object","properties":{"title":{"type":"string","description":"Question or action to confirm"},"message":{"type":"string","description":"Additional context or details"},"confirm_label":{"type":"string","description":"Label for the confirm button (e.g. Deploy, Yes, Approve)"},"deny_label":{"type":"string","description":"Label for the cancel button (e.g. Cancel, No, Deny)"}},"required":["title","confirm_label","deny_label"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Title        string `json:"title"`
				Message      string `json:"message"`
				ConfirmLabel string `json:"confirm_label"`
				DenyLabel    string `json:"deny_label"`
			}
			json.Unmarshal(args, &p)
			return UiConfirm{Title: p.Title, Message: p.Message, ConfirmLabel: p.ConfirmLabel, DenyLabel: p.DenyLabel}, nil
		},
	},
	{
		component:   "Choices",
		toolName:    "render_choices",
		description: "Show clickable options for the user to pick from. The user's selection is sent as their next chat message. Use when offering multiple choices or asking the user what they want to do.",
		schema:      `{"type":"object","properties":{"title":{"type":"string","description":"Prompt or question for the user"},"options":{"type":"array","items":{"type":"string"},"description":"List of options to choose from"},"style":{"type":"string","enum":["buttons","list"],"description":"Display style: buttons (horizontal pills) or list (vertical with radio indicators). Default: buttons"}},"required":["title","options"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Title   string   `json:"title"`
				Options []string `json:"options"`
				Style   string   `json:"style"`
			}
			json.Unmarshal(args, &p)
			opts := make([]any, len(p.Options))
			for i, o := range p.Options {
				opts[i] = o
			}
			return UiChoices{Title: p.Title, Options: opts, Style: p.Style}, nil
		},
	},
	{
		component:   "Table",
		toolName:    "render_table",
		description: "Display data in a searchable, scrollable table with sticky headers. Use for presenting structured data, lists of items, or query results. Supports optional tabs for multi-dataset display.",
		schema:      `{"type":"object","properties":{"title":{"type":"string","description":"Table title"},"headers":{"type":"array","items":{"type":"string"},"description":"Column header names"},"rows":{"type":"array","items":{"type":"array","items":{"type":"string"}},"description":"Table rows, each row is an array of cell values"},"tabs":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"headers":{"type":"array","items":{"type":"string"}},"rows":{"type":"array","items":{"type":"array","items":{"type":"string"}}}},"required":["name","headers","rows"]},"description":"Optional named tabs, each with its own headers and rows. When provided, renders as a tabbed table."}},"required":["title"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Title   string     `json:"title"`
				Headers []string   `json:"headers"`
				Rows    [][]string `json:"rows"`
				Tabs    []struct {
					Name    string     `json:"name"`
					Headers []string   `json:"headers"`
					Rows    [][]string `json:"rows"`
				} `json:"tabs"`
			}
			json.Unmarshal(args, &p)
			hdrs := make([]any, len(p.Headers))
			for i, h := range p.Headers {
				hdrs[i] = h
			}
			rows := make([]any, len(p.Rows))
			for i, r := range p.Rows {
				cells := make([]any, len(r))
				for j, c := range r {
					cells[j] = c
				}
				rows[i] = cells
			}
			var tabs []any
			for _, t := range p.Tabs {
				th := make([]any, len(t.Headers))
				for i, h := range t.Headers {
					th[i] = h
				}
				tr := make([]any, len(t.Rows))
				for i, r := range t.Rows {
					cells := make([]any, len(r))
					for j, c := range r {
						cells[j] = c
					}
					tr[i] = cells
				}
				tabs = append(tabs, UiTab{Name: t.Name, Headers: th, Rows: tr})
			}
			return UiTable{Title: p.Title, Headers: hdrs, Rows: rows, Tabs: tabs}, nil
		},
	},
	{
		component:   "CodeBlock",
		toolName:    "render_codeblock",
		description: "Display code with syntax highlighting. Use for showing code snippets, SQL queries, configuration files, or any formatted text. Supports optional tabs for multiple code sections.",
		schema:      `{"type":"object","properties":{"title":{"type":"string","description":"Title above the code block"},"language":{"type":"string","description":"Programming language for syntax highlighting (e.g. sql, go, python, json)"},"code":{"type":"string","description":"The code content to display"},"tabs":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"language":{"type":"string"},"code":{"type":"string"}},"required":["name","code"]},"description":"Optional named tabs, each with its own code content. When provided, renders as a tabbed code block."}},"required":["title"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Title    string `json:"title"`
				Language string `json:"language"`
				Code     string `json:"code"`
				Tabs     []struct {
					Name     string `json:"name"`
					Language string `json:"language"`
					Code     string `json:"code"`
				} `json:"tabs"`
			}
			json.Unmarshal(args, &p)
			var tabs []any
			for _, t := range p.Tabs {
				tabs = append(tabs, UiCodeTab{Name: t.Name, Language: t.Language, Code: t.Code})
			}
			return UiCodeBlock{Title: p.Title, Language: p.Language, Code: p.Code, Tabs: tabs}, nil
		},
	},
	{
		component:   "Diff",
		toolName:    "render_diff",
		description: "Show a before/after code comparison. Use for displaying changes, migrations, or configuration differences.",
		schema:      `{"type":"object","properties":{"title":{"type":"string","description":"Title for the diff"},"before_label":{"type":"string","description":"Label for the before panel"},"after_label":{"type":"string","description":"Label for the after panel"},"before":{"type":"string","description":"Original content"},"after":{"type":"string","description":"Modified content"},"language":{"type":"string","description":"Language for syntax highlighting"}},"required":["title","before","after"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Title       string `json:"title"`
				BeforeLabel string `json:"before_label"`
				AfterLabel  string `json:"after_label"`
				Before      string `json:"before"`
				After       string `json:"after"`
				Language    string `json:"language"`
			}
			json.Unmarshal(args, &p)
			return UiDiff{Title: p.Title, BeforeLabel: p.BeforeLabel, AfterLabel: p.AfterLabel, Before: p.Before, After: p.After, Language: p.Language}, nil
		},
	},
	{
		component:   "KeyValue",
		toolName:    "render_keyvalue",
		description: "Display key-value pairs in a structured card. Use for showing metadata, configuration details, or summary information.",
		schema:      `{"type":"object","properties":{"title":{"type":"string","description":"Card title"},"items":{"type":"array","items":{"type":"object","properties":{"key":{"type":"string"},"value":{"type":"string"},"style":{"type":"string","enum":["","success","error","warning","code"]}},"required":["key","value"]},"description":"List of key-value pairs"}},"required":["title","items"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Title string `json:"title"`
				Items []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
					Style string `json:"style"`
				} `json:"items"`
			}
			json.Unmarshal(args, &p)
			items := make([]any, len(p.Items))
			for i, item := range p.Items {
				items[i] = UiKVItem{Key: item.Key, Value: item.Value, Style: item.Style}
			}
			return UiKeyValue{Title: p.Title, Items: items}, nil
		},
	},
	{
		component:   "Progress",
		toolName:    "render_progress",
		description: "Show a multi-step progress indicator. Use for displaying pipeline status, task progress, or workflow steps.",
		schema:      `{"type":"object","properties":{"title":{"type":"string","description":"Progress title"},"steps":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"status":{"type":"string","enum":["pending","running","done","failed"]},"detail":{"type":"string"}},"required":["name","status"]},"description":"List of steps with their status"}},"required":["title","steps"]}`,
		handler: func(args json.RawMessage) (any, error) {
			var p struct {
				Title string `json:"title"`
				Steps []struct {
					Name   string `json:"name"`
					Status string `json:"status"`
					Detail string `json:"detail"`
				} `json:"steps"`
			}
			json.Unmarshal(args, &p)
			steps := make([]any, len(p.Steps))
			for i, s := range p.Steps {
				steps[i] = UiStep{Name: s.Name, Status: s.Status, Detail: s.Detail}
			}
			return UiProgress{Title: p.Title, Steps: steps}, nil
		},
	},
}

// RegisterUITools registers ALL built-in UI component tools into a registry.
func RegisterUITools(reg *ToolRegistry) {
	for i := range uiToolSpecs {
		reg.Register(uiToolSpecs[i].toToolDef())
	}
}

// RegisterUIToolsByName registers specific UI component tools by name.
func RegisterUIToolsByName(reg *ToolRegistry, names []string) {
	for _, name := range names {
		for i := range uiToolSpecs {
			if uiToolSpecs[i].component == name {
				reg.Register(uiToolSpecs[i].toToolDef())
				break
			}
		}
	}
}

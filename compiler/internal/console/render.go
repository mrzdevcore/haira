// Package console implements the haira console CLI client.
package console

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ---------------------------------------------------------------------------
// ANSI escape codes
// ---------------------------------------------------------------------------

var (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiWhite  = "\033[37m"
)

func init() {
	if !isTTY() || os.Getenv("NO_COLOR") != "" {
		ansiReset = ""
		ansiBold = ""
		ansiDim = ""
		ansiRed = ""
		ansiGreen = ""
		ansiYellow = ""
		ansiCyan = ""
		ansiWhite = ""
	}
}

// isTTY returns true if stdout is a terminal.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ---------------------------------------------------------------------------
// Renderer — dispatches SSE events to terminal output
// ---------------------------------------------------------------------------

// renderer writes ARP-compatible SSE events to the terminal.
type renderer struct {
	out         io.Writer    // stdout for content
	status      io.Writer    // stderr for status indicators
	term        *Terminal    // terminal for width, spinners
	thinking    *Spinner     // "Thinking" spinner
	toolSpinner *ToolSpinner // active tool spinner
	thinkingStopped bool     // tracks if thinking was already stopped
}

func newRenderer(term *Terminal) *renderer {
	return &renderer{
		out:    os.Stdout,
		status: os.Stderr,
		term:   term,
	}
}

// startThinking begins a "Thinking" spinner before the HTTP request.
func (r *renderer) startThinking() {
	r.thinkingStopped = false
	if r.term.IsRaw() {
		r.thinking = NewSpinner(r.term, "Thinking")
		r.thinking.Start()
	} else {
		fmt.Fprintf(r.status, "%s  Thinking...%s\n", ansiDim, ansiReset)
		r.thinkingStopped = true
	}
}

// stopThinking stops the "Thinking" spinner (idempotent).
func (r *renderer) stopThinking() {
	if r.thinkingStopped {
		return
	}
	r.thinkingStopped = true
	if r.thinking != nil {
		r.thinking.Stop()
		r.thinking = nil
	}
}

// renderEvent dispatches a parsed SSE event to the appropriate handler.
func (r *renderer) renderEvent(event, data string) {
	switch event {
	case "tool_start":
		r.renderToolStart(data)
	case "tool_end":
		r.renderToolEnd(data)
	case "tool_render":
		r.renderToolRender(data)
	case "error":
		r.renderError(data)
	case "step":
		r.renderStep(data)
	case "":
		// Default data event — either delta or [DONE]
		r.renderData(data)
	}
}

func (r *renderer) renderToolStart(data string) {
	r.stopThinking()
	var p struct {
		Tool string `json:"tool"`
		Args string `json:"args"`
	}
	json.Unmarshal([]byte(data), &p)

	if r.term.IsRaw() {
		r.toolSpinner = NewToolSpinner(r.term, p.Tool)
		r.toolSpinner.Start()
	} else {
		fmt.Fprintf(r.status, "%s  ◌ %s ...%s\n", ansiDim, p.Tool, ansiReset)
	}
}

func (r *renderer) renderToolEnd(data string) {
	var p struct {
		Tool string `json:"tool"`
		OK   bool   `json:"ok"`
	}
	json.Unmarshal([]byte(data), &p)

	if r.toolSpinner != nil {
		r.toolSpinner.Done(p.OK)
		r.toolSpinner = nil
	} else {
		if p.OK {
			fmt.Fprintf(r.status, "%s  ✓ %s%s\n", ansiGreen, p.Tool, ansiReset)
		} else {
			fmt.Fprintf(r.status, "%s  ✗ %s%s\n", ansiRed, p.Tool, ansiReset)
		}
	}
}

func (r *renderer) renderToolRender(data string) {
	r.stopThinking()
	var p struct {
		Tool      string          `json:"tool"`
		Component string          `json:"component"`
		Props     json.RawMessage `json:"props"`
	}
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return
	}

	r.renderComponent(p.Component, p.Props)
}

// renderComponent renders a single ARP component by type.
// Used both for top-level tool_render events and for nested children in groups.
func (r *renderer) renderComponent(component string, props json.RawMessage) {
	switch component {
	case "table":
		r.renderTable(props)
	case "status-card":
		r.renderStatusCard(props)
	case "code-block":
		r.renderCodeBlock(props)
	case "key-value":
		r.renderKeyValue(props)
	case "progress":
		r.renderProgress(props)
	case "choices":
		r.renderChoices(props)
	case "text", "markdown":
		r.renderText(props)
	case "group":
		r.renderGroup(props)
	case "product-cards":
		r.renderProductCards(props)
	case "image":
		r.renderImage(props)
	case "diff":
		r.renderDiff(props)
	default:
		// Fallback: try to extract and render children if present
		var generic struct {
			Children []struct {
				Component string          `json:"component"`
				Props     json.RawMessage `json:"props"`
			} `json:"children"`
		}
		if json.Unmarshal(props, &generic) == nil && len(generic.Children) > 0 {
			for _, child := range generic.Children {
				r.renderComponent(child.Component, child.Props)
			}
			return
		}
		r.renderFallback(component, props)
	}
}

func (r *renderer) renderData(data string) {
	if data == "[DONE]" {
		fmt.Fprintln(r.out)
		return
	}
	r.stopThinking()
	var p struct {
		Delta string `json:"delta"`
	}
	if err := json.Unmarshal([]byte(data), &p); err == nil && p.Delta != "" {
		fmt.Fprint(r.out, p.Delta)
	}
}

func (r *renderer) renderError(data string) {
	r.stopThinking()
	var p struct {
		Error string `json:"error"`
	}
	json.Unmarshal([]byte(data), &p)
	fmt.Fprintf(r.status, "%s%sError: %s%s\n", ansiRed, ansiBold, p.Error, ansiReset)
}

func (r *renderer) renderStep(data string) {
	r.stopThinking()
	var p struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}
	json.Unmarshal([]byte(data), &p)
	switch p.Status {
	case "start":
		fmt.Fprintf(r.status, "%s  ◌ %s ...%s\n", ansiDim, p.Name, ansiReset)
	case "end":
		fmt.Fprintf(r.status, "%s  ✓ %s%s\n", ansiGreen, p.Name, ansiReset)
	case "failed":
		msg := p.Name
		if p.Error != "" {
			msg += ": " + p.Error
		}
		fmt.Fprintf(r.status, "%s  ✗ %s%s\n", ansiRed, msg, ansiReset)
	}
}

// ---------------------------------------------------------------------------
// Component renderers
// ---------------------------------------------------------------------------

func (r *renderer) renderTable(props json.RawMessage) {
	var t struct {
		Headers []string        `json:"headers"`
		Rows    [][]interface{} `json:"rows"`
		Title   string          `json:"title"`
	}
	if err := json.Unmarshal(props, &t); err != nil {
		fmt.Fprintf(r.out, "%s\n", props)
		return
	}

	if t.Title != "" {
		fmt.Fprintf(r.out, "\n%s%s%s\n", ansiBold, t.Title, ansiReset)
	}

	if len(t.Headers) == 0 && len(t.Rows) == 0 {
		return
	}

	// Calculate column widths
	cols := len(t.Headers)
	for _, row := range t.Rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	widths := make([]int, cols)
	for i, h := range t.Headers {
		if len(h) > widths[i] {
			widths[i] = len(h)
		}
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			s := fmt.Sprintf("%v", cell)
			if len(s) > widths[i] {
				widths[i] = len(s)
			}
		}
	}

	// Width-aware: constrain columns to fit terminal
	termWidth := r.term.Width()
	// Overhead: 2 (indent) + 1 (│) per column start + 3 (space+│+space) per column
	overhead := 2 + 1 + cols*3
	available := termWidth - overhead
	if available < cols*3 {
		available = cols * 3 // minimum 3 chars per column
	}

	totalWidth := 0
	for _, w := range widths {
		totalWidth += w
	}
	if totalWidth > available {
		ratio := float64(available) / float64(totalWidth)
		for i := range widths {
			w := int(float64(widths[i]) * ratio)
			if w < 3 {
				w = 3
			}
			widths[i] = w
		}
	}

	// Print header
	if len(t.Headers) > 0 {
		r.printTableRow(t.Headers, widths, true)
		r.printTableSep(widths)
	}
	// Print rows
	for _, row := range t.Rows {
		strs := make([]string, len(row))
		for i, cell := range row {
			strs[i] = fmt.Sprintf("%v", cell)
		}
		r.printTableRow(strs, widths, false)
	}
	fmt.Fprintln(r.out)
}

func (r *renderer) printTableRow(cells []string, widths []int, bold bool) {
	fmt.Fprint(r.out, "  │")
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		// Truncate if too wide
		if len(cell) > w {
			if w > 3 {
				cell = cell[:w-3] + "..."
			} else {
				cell = cell[:w]
			}
		}
		padding := w - len(cell)
		if padding < 0 {
			padding = 0
		}
		if bold {
			fmt.Fprintf(r.out, " %s%s%s%s │", ansiBold, cell, ansiReset, strings.Repeat(" ", padding))
		} else {
			fmt.Fprintf(r.out, " %s%s │", cell, strings.Repeat(" ", padding))
		}
	}
	fmt.Fprintln(r.out)
}

func (r *renderer) printTableSep(widths []int) {
	fmt.Fprint(r.out, "  ├")
	for i, w := range widths {
		fmt.Fprint(r.out, strings.Repeat("─", w+2))
		if i < len(widths)-1 {
			fmt.Fprint(r.out, "┼")
		}
	}
	fmt.Fprintln(r.out, "┤")
}

func (r *renderer) renderStatusCard(props json.RawMessage) {
	var c struct {
		Title   string `json:"title"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	json.Unmarshal(props, &c)
	color := ansiCyan
	switch c.Status {
	case "error", "failed":
		color = ansiRed
	case "success", "ok":
		color = ansiGreen
	case "warning":
		color = ansiYellow
	}
	fmt.Fprintf(r.out, "\n  %s■ %s%s", color, c.Title, ansiReset)
	if c.Status != "" {
		fmt.Fprintf(r.out, " %s(%s)%s", ansiDim, c.Status, ansiReset)
	}
	fmt.Fprintln(r.out)
	if c.Message != "" {
		fmt.Fprintf(r.out, "  %s\n", c.Message)
	}
}

func (r *renderer) renderCodeBlock(props json.RawMessage) {
	var c struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}
	json.Unmarshal(props, &c)
	label := "code"
	if c.Language != "" {
		label = c.Language
	}
	fmt.Fprintf(r.out, "\n  %s─── %s ───%s\n", ansiDim, label, ansiReset)
	for _, line := range strings.Split(c.Code, "\n") {
		fmt.Fprintf(r.out, "  %s\n", line)
	}
	fmt.Fprintf(r.out, "  %s──────────%s\n", ansiDim, ansiReset)
}

func (r *renderer) renderKeyValue(props json.RawMessage) {
	var kv struct {
		Title string `json:"title"`
		Items []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"items"`
	}
	if err := json.Unmarshal(props, &kv); err != nil {
		// Try flat map fallback
		var m map[string]interface{}
		if err2 := json.Unmarshal(props, &m); err2 == nil {
			for k, v := range m {
				fmt.Fprintf(r.out, "    %s%s:%s %v\n", ansiBold, k, ansiReset, v)
			}
			return
		}
		return
	}
	if kv.Title != "" {
		fmt.Fprintf(r.out, "  %s%s%s\n", ansiBold, kv.Title, ansiReset)
	}
	for _, item := range kv.Items {
		fmt.Fprintf(r.out, "    %s%s:%s %s\n", ansiBold, item.Key, ansiReset, item.Value)
	}
}

func (r *renderer) renderProgress(props json.RawMessage) {
	var p struct {
		Value   float64 `json:"value"`
		Max     float64 `json:"max"`
		Label   string  `json:"label"`
		Percent float64 `json:"percent"`
	}
	json.Unmarshal(props, &p)

	pct := p.Percent
	if pct == 0 && p.Max > 0 {
		pct = (p.Value / p.Max) * 100
	}

	// Width-aware bar size
	barWidth := r.term.Width() - 20 // room for brackets, label, indent
	if barWidth < 10 {
		barWidth = 10
	}
	if barWidth > 50 {
		barWidth = 50
	}

	filled := int(pct / 100.0 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	empty := barWidth - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	label := p.Label
	if label == "" {
		label = fmt.Sprintf("%.0f%%", pct)
	}
	fmt.Fprintf(r.out, "  [%s] %s\n", bar, label)
}

func (r *renderer) renderChoices(props json.RawMessage) {
	var c struct {
		Title   string   `json:"title"`
		Options []string `json:"options"`
		Items   []struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"items"`
	}
	json.Unmarshal(props, &c)
	if c.Title != "" {
		fmt.Fprintf(r.out, "\n  %s%s%s\n", ansiBold, c.Title, ansiReset)
	}
	if len(c.Options) > 0 {
		for i, opt := range c.Options {
			fmt.Fprintf(r.out, "  %s%d.%s %s\n", ansiCyan, i+1, ansiReset, opt)
		}
	} else {
		for i, item := range c.Items {
			fmt.Fprintf(r.out, "  %s%d.%s %s\n", ansiCyan, i+1, ansiReset, item.Label)
		}
	}
}

func (r *renderer) renderText(props json.RawMessage) {
	var t struct {
		Text    string `json:"text"`
		Content string `json:"content"`
	}
	json.Unmarshal(props, &t)
	text := t.Text
	if text == "" {
		text = t.Content
	}
	if text != "" {
		fmt.Fprintf(r.out, "%s\n", text)
	}
}

func (r *renderer) renderGroup(props json.RawMessage) {
	var g struct {
		Title    string `json:"title"`
		Children []struct {
			Component string          `json:"component"`
			Props     json.RawMessage `json:"props"`
		} `json:"children"`
	}
	if err := json.Unmarshal(props, &g); err != nil {
		return
	}
	if g.Title != "" {
		fmt.Fprintf(r.out, "\n  %s%s%s\n", ansiBold, g.Title, ansiReset)
	}
	for _, child := range g.Children {
		r.renderComponent(child.Component, child.Props)
	}
}

func (r *renderer) renderImage(props json.RawMessage) {
	var img struct {
		URL string `json:"url"`
		Alt string `json:"alt"`
		Src string `json:"src"`
	}
	json.Unmarshal(props, &img)
	url := img.URL
	if url == "" {
		url = img.Src
	}
	alt := img.Alt
	if alt == "" {
		alt = "image"
	}
	fmt.Fprintf(r.out, "  %s[%s]%s %s\n", ansiDim, alt, ansiReset, url)
}

func (r *renderer) renderDiff(props json.RawMessage) {
	var d struct {
		Old      string `json:"old"`
		New      string `json:"new"`
		Language string `json:"language"`
		Diff     string `json:"diff"`
	}
	json.Unmarshal(props, &d)

	if d.Diff != "" {
		fmt.Fprintf(r.out, "\n  %s─── diff ───%s\n", ansiDim, ansiReset)
		for _, line := range strings.Split(d.Diff, "\n") {
			if strings.HasPrefix(line, "+") {
				fmt.Fprintf(r.out, "  %s%s%s\n", ansiGreen, line, ansiReset)
			} else if strings.HasPrefix(line, "-") {
				fmt.Fprintf(r.out, "  %s%s%s\n", ansiRed, line, ansiReset)
			} else {
				fmt.Fprintf(r.out, "  %s\n", line)
			}
		}
		fmt.Fprintf(r.out, "  %s──────────%s\n", ansiDim, ansiReset)
		return
	}

	if d.Old != "" || d.New != "" {
		label := "diff"
		if d.Language != "" {
			label = d.Language
		}
		fmt.Fprintf(r.out, "\n  %s─── %s ───%s\n", ansiDim, label, ansiReset)
		if d.Old != "" {
			for _, line := range strings.Split(d.Old, "\n") {
				fmt.Fprintf(r.out, "  %s- %s%s\n", ansiRed, line, ansiReset)
			}
		}
		if d.New != "" {
			for _, line := range strings.Split(d.New, "\n") {
				fmt.Fprintf(r.out, "  %s+ %s%s\n", ansiGreen, line, ansiReset)
			}
		}
		fmt.Fprintf(r.out, "  %s──────────%s\n", ansiDim, ansiReset)
	}
}

func (r *renderer) renderProductCards(props json.RawMessage) {
	var pc struct {
		Title string `json:"title"`
		Cards []struct {
			Name        string `json:"name"`
			Price       string `json:"price"`
			Brand       string `json:"brand,omitempty"`
			Badge       string `json:"badge,omitempty"`
			Description string `json:"description,omitempty"`
			URL         string `json:"url,omitempty"`
		} `json:"cards"`
	}
	if err := json.Unmarshal(props, &pc); err != nil {
		return
	}

	if pc.Title != "" {
		fmt.Fprintf(r.out, "\n  %s%s%s\n", ansiBold, pc.Title, ansiReset)
	}

	// Width-aware description truncation
	maxDesc := r.term.Width() - 10
	if maxDesc < 30 {
		maxDesc = 30
	}
	if maxDesc > 80 {
		maxDesc = 80
	}

	for i, card := range pc.Cards {
		fmt.Fprintf(r.out, "  %s%d.%s %s%s%s", ansiCyan, i+1, ansiReset, ansiBold, card.Name, ansiReset)
		if card.Badge != "" {
			fmt.Fprintf(r.out, " %s[%s]%s", ansiYellow, card.Badge, ansiReset)
		}
		fmt.Fprintln(r.out)

		details := ""
		if card.Brand != "" {
			details = card.Brand
		}
		if card.Price != "" {
			if details != "" {
				details += " — "
			}
			details += card.Price
		}
		if details != "" {
			fmt.Fprintf(r.out, "     %s\n", details)
		}

		if card.Description != "" {
			desc := card.Description
			if len(desc) > maxDesc {
				desc = desc[:maxDesc-3] + "..."
			}
			fmt.Fprintf(r.out, "     %s%s%s\n", ansiDim, desc, ansiReset)
		}

		if card.URL != "" {
			fmt.Fprintf(r.out, "     %s%s%s\n", ansiDim, card.URL, ansiReset)
		}
	}
	if len(pc.Cards) > 0 {
		fmt.Fprintln(r.out)
	}
}

func (r *renderer) renderFallback(component string, props json.RawMessage) {
	var fields map[string]interface{}
	if json.Unmarshal(props, &fields) != nil {
		return
	}

	for _, key := range []string{"text", "content", "message", "title", "label", "description"} {
		if val, ok := fields[key]; ok {
			if s, ok := val.(string); ok && s != "" {
				fmt.Fprintf(r.out, "  %s\n", s)
				return
			}
		}
	}

	fmt.Fprintf(r.out, "%s  [%s]%s\n", ansiDim, component, ansiReset)
}

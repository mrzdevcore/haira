package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// testServer creates a server with a buffer for output and returns
// the server and a function to read sent messages.
func testServer() (*Server, *bytes.Buffer) {
	var out bytes.Buffer
	r := strings.NewReader("") // empty input — we call handler methods directly
	s := NewServer(r, &out)
	return s, &out
}

// readMessages parses all LSP messages from the output buffer.
func readMessages(buf *bytes.Buffer) []json.RawMessage {
	var msgs []json.RawMessage
	data := buf.String()
	for {
		idx := strings.Index(data, "Content-Length: ")
		if idx < 0 {
			break
		}
		data = data[idx:]
		nlIdx := strings.Index(data, "\r\n")
		if nlIdx < 0 {
			break
		}
		lenStr := data[len("Content-Length: "):nlIdx]
		var length int
		fmt.Sscanf(lenStr, "%d", &length)

		// Find end of headers
		headerEnd := strings.Index(data, "\r\n\r\n")
		if headerEnd < 0 {
			break
		}
		bodyStart := headerEnd + 4
		if bodyStart+length > len(data) {
			break
		}
		body := data[bodyStart : bodyStart+length]
		msgs = append(msgs, json.RawMessage(body))
		data = data[bodyStart+length:]
	}
	return msgs
}

func TestInitialize(t *testing.T) {
	s, _ := testServer()
	result := s.handler.Initialize(nil)
	if result == nil {
		t.Fatal("expected InitializeResult")
	}
	if result.ServerInfo.Name != "haira-lsp" {
		t.Errorf("server name = %q, want %q", result.ServerInfo.Name, "haira-lsp")
	}
	if result.Capabilities.TextDocumentSync != 1 {
		t.Error("expected full text sync (1)")
	}
	if !result.Capabilities.HoverProvider {
		t.Error("expected hover provider")
	}
	if !result.Capabilities.DefinitionProvider {
		t.Error("expected definition provider")
	}
	if result.Capabilities.CompletionProvider == nil {
		t.Error("expected completion provider")
	}
}

func TestDidOpenPublishesDiagnostics(t *testing.T) {
	s, out := testServer()

	params, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///test.haira",
			LanguageID: "haira",
			Version:    1,
			Text:       `fn main() { io.println("hello") }`,
		},
	})

	s.handler.DidOpen(params)

	msgs := readMessages(out)
	if len(msgs) == 0 {
		t.Fatal("expected diagnostics notification")
	}

	// Parse the notification
	var notif struct {
		Method string                   `json:"method"`
		Params PublishDiagnosticsParams `json:"params"`
	}
	json.Unmarshal(msgs[0], &notif)

	if notif.Method != "textDocument/publishDiagnostics" {
		t.Errorf("method = %q, want textDocument/publishDiagnostics", notif.Method)
	}
	if notif.Params.URI != "file:///test.haira" {
		t.Errorf("uri = %q, want file:///test.haira", notif.Params.URI)
	}
	// Valid code should have no error diagnostics
	for _, d := range notif.Params.Diagnostics {
		if d.Severity == 1 {
			t.Errorf("unexpected error: %s", d.Message)
		}
	}
}

func TestDidOpenWithErrors(t *testing.T) {
	s, out := testServer()

	params, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        "file:///bad.haira",
			LanguageID: "haira",
			Version:    1,
			Text:       `fn { bad syntax`,
		},
	})

	s.handler.DidOpen(params)

	msgs := readMessages(out)
	if len(msgs) == 0 {
		t.Fatal("expected diagnostics notification")
	}

	var notif struct {
		Params PublishDiagnosticsParams `json:"params"`
	}
	json.Unmarshal(msgs[0], &notif)

	if len(notif.Params.Diagnostics) == 0 {
		t.Error("expected parse error diagnostics")
	}
	foundError := false
	for _, d := range notif.Params.Diagnostics {
		if d.Severity == 1 {
			foundError = true
		}
	}
	if !foundError {
		t.Error("expected at least one error-severity diagnostic")
	}
}

func TestDidChange(t *testing.T) {
	s, out := testServer()

	// Open
	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: `fn main() {}`,
		},
	})
	s.handler.DidOpen(openParams)
	out.Reset()

	// Change to broken code
	changeParams, _ := json.Marshal(DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			URI:     "file:///test.haira",
			Version: 2,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: `fn { broken`},
		},
	})
	s.handler.DidChange(changeParams)

	msgs := readMessages(out)
	if len(msgs) == 0 {
		t.Fatal("expected diagnostics after change")
	}

	var notif struct {
		Params PublishDiagnosticsParams `json:"params"`
	}
	json.Unmarshal(msgs[0], &notif)

	if len(notif.Params.Diagnostics) == 0 {
		t.Error("expected errors for broken code")
	}
}

func TestDidClose(t *testing.T) {
	s, out := testServer()

	// Open
	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: `fn main() {}`,
		},
	})
	s.handler.DidOpen(openParams)
	out.Reset()

	// Close
	closeParams, _ := json.Marshal(DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.haira"},
	})
	s.handler.DidClose(closeParams)

	msgs := readMessages(out)
	if len(msgs) == 0 {
		t.Fatal("expected empty diagnostics on close")
	}

	var notif struct {
		Params PublishDiagnosticsParams `json:"params"`
	}
	json.Unmarshal(msgs[0], &notif)

	if len(notif.Params.Diagnostics) != 0 {
		t.Error("expected diagnostics to be cleared on close")
	}
}

func TestHoverOnFunction(t *testing.T) {
	s, _ := testServer()

	src := `fn add(a: int, b: int) -> int {
	return a + b
}

fn main() {
	x = add(1, 2)
}`

	// Open the document
	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: src,
		},
	})
	s.handler.DidOpen(openParams)

	// Hover over "add" in the function definition (line 0, char 3)
	hoverParams, _ := json.Marshal(TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.haira"},
		Position:     Position{Line: 0, Character: 4},
	})
	result := s.handler.Hover(hoverParams)
	if result == nil {
		t.Fatal("expected hover result")
	}
	if !strings.Contains(result.Contents.Value, "add") {
		t.Errorf("hover should mention 'add', got: %s", result.Contents.Value)
	}
}

func TestHoverOnVariable(t *testing.T) {
	s, _ := testServer()

	src := `fn main() {
	x = 42
	y = x
}`

	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: src,
		},
	})
	s.handler.DidOpen(openParams)

	// Hover over "42" (line 1, char 5)
	hoverParams, _ := json.Marshal(TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.haira"},
		Position:     Position{Line: 1, Character: 5},
	})
	result := s.handler.Hover(hoverParams)
	if result == nil {
		t.Fatal("expected hover result for literal")
	}
	if !strings.Contains(result.Contents.Value, "int") {
		t.Errorf("hover should show 'int', got: %s", result.Contents.Value)
	}
}

func TestCompletion(t *testing.T) {
	s, _ := testServer()

	src := `fn greet(name: string) -> string {
	return "Hello, " + name
}

fn main() {
	gr
}`

	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: src,
		},
	})
	s.handler.DidOpen(openParams)

	// Complete at "gr" (line 5, char 3)
	compParams, _ := json.Marshal(CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.haira"},
		Position:     Position{Line: 5, Character: 3},
	})
	result := s.handler.Completion(compParams)
	if result == nil {
		t.Fatal("expected completion result")
	}

	found := false
	for _, item := range result.Items {
		if item.Label == "greet" {
			found = true
			break
		}
	}
	if !found {
		labels := make([]string, len(result.Items))
		for i, item := range result.Items {
			labels[i] = item.Label
		}
		t.Errorf("expected 'greet' in completions, got: %v", labels)
	}
}

func TestCompletionKeywords(t *testing.T) {
	s, _ := testServer()

	src := `fn main() {
	fo
}`

	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: src,
		},
	})
	s.handler.DidOpen(openParams)

	compParams, _ := json.Marshal(CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.haira"},
		Position:     Position{Line: 1, Character: 3},
	})
	result := s.handler.Completion(compParams)
	if result == nil {
		t.Fatal("expected completion result")
	}

	found := false
	for _, item := range result.Items {
		if item.Label == "for" && item.Kind == 14 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'for' keyword in completions")
	}
}

func TestDefinition(t *testing.T) {
	s, _ := testServer()

	src := `fn add(a: int, b: int) -> int {
	return a + b
}

fn main() {
	x = add(1, 2)
}`

	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: src,
		},
	})
	s.handler.DidOpen(openParams)

	// Go to definition of "add" on line 5 (the call site)
	defParams, _ := json.Marshal(TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.haira"},
		Position:     Position{Line: 5, Character: 5},
	})
	result := s.handler.Definition(defParams)
	if result == nil {
		t.Fatal("expected definition location")
	}
	if result.URI != "file:///test.haira" {
		t.Errorf("uri = %q, want file:///test.haira", result.URI)
	}
	// The definition should point to line 0 (where fn add is defined)
	if result.Range.Start.Line != 0 {
		t.Errorf("definition line = %d, want 0", result.Range.Start.Line)
	}
}

func TestDefinitionTypeDef(t *testing.T) {
	s, _ := testServer()

	src := `struct User {
	name: string
}

fn main() {
	u = User{name = "Alice"}
}`

	openParams, _ := json.Marshal(DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  "file:///test.haira",
			Text: src,
		},
	})
	s.handler.DidOpen(openParams)

	// Go to definition of "User" on line 5
	defParams, _ := json.Marshal(TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///test.haira"},
		Position:     Position{Line: 5, Character: 5},
	})
	result := s.handler.Definition(defParams)
	if result == nil {
		t.Fatal("expected definition location for User")
	}
	if result.Range.Start.Line != 0 {
		t.Errorf("definition line = %d, want 0", result.Range.Start.Line)
	}
}

// --- Helper tests ---

func TestOffsetToPosition(t *testing.T) {
	text := "abc\ndef\nghi"
	cases := []struct {
		offset int
		want   Position
	}{
		{0, Position{0, 0}},
		{2, Position{0, 2}},
		{4, Position{1, 0}},
		{7, Position{1, 3}},
		{8, Position{2, 0}},
	}
	for _, c := range cases {
		got := offsetToPosition(text, c.offset)
		if got != c.want {
			t.Errorf("offsetToPosition(%d) = %v, want %v", c.offset, got, c.want)
		}
	}
}

func TestPositionToOffset(t *testing.T) {
	text := "abc\ndef\nghi"
	cases := []struct {
		pos  Position
		want int
	}{
		{Position{0, 0}, 0},
		{Position{0, 2}, 2},
		{Position{1, 0}, 4},
		{Position{2, 0}, 8},
	}
	for _, c := range cases {
		got := positionToOffset(text, c.pos)
		if got != c.want {
			t.Errorf("positionToOffset(%v) = %d, want %d", c.pos, got, c.want)
		}
	}
}

func TestGetWordAt(t *testing.T) {
	text := "hello world_test 123"
	cases := []struct {
		offset int
		want   string
	}{
		{0, "hello"},
		{3, "hello"},
		{6, "world_test"},
		{5, "hello"}, // boundary — backward scan finds "hello"
	}
	for _, c := range cases {
		got := getWordAt(text, c.offset)
		if got != c.want {
			t.Errorf("getWordAt(%d) = %q, want %q", c.offset, got, c.want)
		}
	}
}

func TestURIToPath(t *testing.T) {
	got := URIToPath("file:///home/user/test.haira")
	if got != "/home/user/test.haira" {
		t.Errorf("URIToPath = %q, want /home/user/test.haira", got)
	}

	got = URIToPath("/plain/path")
	if got != "/plain/path" {
		t.Errorf("URIToPath plain = %q, want /plain/path", got)
	}
}

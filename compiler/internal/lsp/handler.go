package lsp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
	hairaerr "github.com/haira-lang/haira/internal/errors"
	"github.com/haira-lang/haira/internal/parser"
	"github.com/haira-lang/haira/internal/token"
)

// Handler processes LSP requests and notifications.
type Handler struct {
	server *Server
	mu     sync.RWMutex

	// Open documents: URI → content
	documents map[string]string

	// Cached parse results: URI → *ast.SourceFile
	asts map[string]*ast.SourceFile

	// Cached type info: URI → *checker.TypeInfo
	typeInfos map[string]*checker.TypeInfo
}

// NewHandler creates a new Handler.
func NewHandler(server *Server) *Handler {
	return &Handler{
		server:    server,
		documents: make(map[string]string),
		asts:      make(map[string]*ast.SourceFile),
		typeInfos: make(map[string]*checker.TypeInfo),
	}
}

// Initialize handles the initialize request.
func (h *Handler) Initialize(params json.RawMessage) *InitializeResult {
	return &InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full sync
			HoverProvider:    true,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", "\""},
			},
			DefinitionProvider: true,
		},
		ServerInfo: &ServerInfo{
			Name:    "haira-lsp",
			Version: "0.1.0",
		},
	}
}

// DidOpen handles textDocument/didOpen.
func (h *Handler) DidOpen(params json.RawMessage) {
	var p DidOpenTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return
	}

	h.mu.Lock()
	h.documents[p.TextDocument.URI] = p.TextDocument.Text
	h.mu.Unlock()

	h.analyzeAndPublish(p.TextDocument.URI, p.TextDocument.Text)
}

// DidChange handles textDocument/didChange.
func (h *Handler) DidChange(params json.RawMessage) {
	var p DidChangeTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return
	}

	if len(p.ContentChanges) == 0 {
		return
	}

	// Full sync — last change has the full text
	text := p.ContentChanges[len(p.ContentChanges)-1].Text

	h.mu.Lock()
	h.documents[p.TextDocument.URI] = text
	h.mu.Unlock()

	h.analyzeAndPublish(p.TextDocument.URI, text)
}

// DidClose handles textDocument/didClose.
func (h *Handler) DidClose(params json.RawMessage) {
	var p DidCloseTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return
	}

	h.mu.Lock()
	delete(h.documents, p.TextDocument.URI)
	delete(h.asts, p.TextDocument.URI)
	delete(h.typeInfos, p.TextDocument.URI)
	h.mu.Unlock()

	// Clear diagnostics
	h.server.SendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         p.TextDocument.URI,
		Diagnostics: []Diagnostic{},
	})
}

// DidSave handles textDocument/didSave.
func (h *Handler) DidSave(params json.RawMessage) {
	var p DidSaveTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return
	}

	h.mu.RLock()
	text, ok := h.documents[p.TextDocument.URI]
	h.mu.RUnlock()

	if ok {
		h.analyzeAndPublish(p.TextDocument.URI, text)
	}
}

// Hover handles textDocument/hover.
func (h *Handler) Hover(params json.RawMessage) *Hover {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil
	}

	h.mu.RLock()
	text, ok := h.documents[p.TextDocument.URI]
	file := h.asts[p.TextDocument.URI]
	typeInfo := h.typeInfos[p.TextDocument.URI]
	h.mu.RUnlock()

	if !ok || file == nil {
		return nil
	}

	offset := positionToOffset(text, p.Position)

	// Try to find a type for the expression or variable at this position
	if typeInfo != nil {
		// Check ExprTypes — find the smallest span containing the offset
		var bestType checker.Type
		bestSize := int(^uint(0) >> 1) // max int
		for span, ty := range typeInfo.ExprTypes {
			if offset >= span.Start && offset < span.End {
				size := span.End - span.Start
				if size < bestSize {
					bestSize = size
					bestType = ty
				}
			}
		}
		// Also check VarTypes
		for span, ty := range typeInfo.VarTypes {
			if offset >= span.Start && offset < span.End {
				size := span.End - span.Start
				if size < bestSize {
					bestSize = size
					bestType = ty
				}
			}
		}
		if bestType != nil {
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: fmt.Sprintf("```haira\n%s\n```", bestType.String()),
				},
			}
		}
	}

	// Fall back to item-level hover (function signatures, type defs, etc.)
	for _, item := range file.Items {
		if offset < item.Span.Start || offset >= item.Span.End {
			continue
		}
		if info := itemHoverInfo(item); info != "" {
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: info,
				},
			}
		}
	}

	return nil
}

// Completion handles textDocument/completion.
func (h *Handler) Completion(params json.RawMessage) *CompletionList {
	var p CompletionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return &CompletionList{Items: []CompletionItem{}}
	}

	h.mu.RLock()
	text, ok := h.documents[p.TextDocument.URI]
	file := h.asts[p.TextDocument.URI]
	h.mu.RUnlock()

	if !ok {
		return &CompletionList{Items: []CompletionItem{}}
	}

	offset := positionToOffset(text, p.Position)
	prefix := getWordPrefix(text, offset)

	items := []CompletionItem{}

	// Keywords
	for kw := range token.Keywords {
		if strings.HasPrefix(kw, prefix) {
			items = append(items, CompletionItem{
				Label: kw,
				Kind:  14, // Keyword
			})
		}
	}

	// Stdlib modules
	for _, mod := range []string{"io", "http", "json", "time", "env", "postgres", "slack", "excel"} {
		if strings.HasPrefix(mod, prefix) {
			items = append(items, CompletionItem{
				Label: mod,
				Kind:  9, // Module
			})
		}
	}

	// Stdlib functions
	stdlibFuncs := map[string]string{
		"io.println": "Print a value with newline",
		"io.print":   "Print a value",
		"http.get":   "HTTP GET request",
		"http.post":  "HTTP POST request",
		"len":        "Get length of a collection",
		"keys":       "Get keys of a map",
		"join":       "Join list elements with separator",
		"env":        "Read environment variable",
	}
	for name, detail := range stdlibFuncs {
		if strings.HasPrefix(name, prefix) {
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   3, // Function
				Detail: detail,
			})
		}
	}

	// Symbols from the current file
	if file != nil {
		for _, item := range file.Items {
			switch it := item.Node.(type) {
			case ast.FunctionDef:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   3, // Function
						Detail: funcSignature(it),
					})
				}
			case ast.TypeDef:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   7, // Class (struct)
						Detail: fmt.Sprintf("type %s (%d fields)", it.Name.Node, len(it.Fields)),
					})
				}
			case ast.EnumDef:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   7,
						Detail: fmt.Sprintf("enum %s (%d variants)", it.Name.Node, len(it.Variants)),
					})
				}
			case ast.ToolDecl:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   3,
						Detail: "tool " + it.Name.Node,
					})
				}
			case ast.AgentDecl:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   7,
						Detail: "agent " + it.Name.Node,
					})
				}
			case ast.WorkflowDecl:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   3,
						Detail: "workflow " + it.Name.Node,
					})
				}
			case ast.ProviderDecl:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   6,
						Detail: "provider " + it.Name.Node,
					})
				}
			case ast.TestDecl:
				if strings.HasPrefix(it.Name.Node, prefix) {
					items = append(items, CompletionItem{
						Label:  it.Name.Node,
						Kind:   12, // Value
						Detail: "test " + it.Name.Node,
					})
				}
			}
		}
	}

	return &CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

// Definition handles textDocument/definition.
func (h *Handler) Definition(params json.RawMessage) []Location {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return []Location{}
	}

	h.mu.RLock()
	text, ok := h.documents[p.TextDocument.URI]
	file := h.asts[p.TextDocument.URI]
	h.mu.RUnlock()

	if !ok || file == nil {
		return []Location{}
	}

	offset := positionToOffset(text, p.Position)
	name := getWordAt(text, offset)
	if name == "" {
		return []Location{}
	}

	// Search for definition in the file
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.FunctionDef:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		case ast.TypeDef:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		case ast.EnumDef:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		case ast.ToolDecl:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		case ast.AgentDecl:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		case ast.WorkflowDecl:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		case ast.MethodDef:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		case ast.TestDecl:
			if it.Name.Node == name {
				return []Location{*spanToLocation(p.TextDocument.URI, text, it.Name.Span)}
			}
		}
	}

	return []Location{}
}

// ---------------------------------------------------------------------------
// Analysis
// ---------------------------------------------------------------------------

// analyzeAndPublish parses and type-checks a document, then publishes diagnostics.
func (h *Handler) analyzeAndPublish(uri, text string) {
	lspDiags := []Diagnostic{}

	// Parse
	file, parseErrs := parser.Parse(text)
	for _, e := range parseErrs {
		lspDiags = append(lspDiags, diagFromHaira(hairaerr.Diagnostic{
			Level:   hairaerr.Error,
			Message: e.Message,
			Span:    e.Span,
		}, text))
	}

	h.mu.Lock()
	if file != nil {
		h.asts[uri] = file
	}
	h.mu.Unlock()

	// Type check (only if parse succeeded)
	if file != nil && len(parseErrs) == 0 {
		typeInfo, typeDiags := checker.Check(file)
		for _, d := range typeDiags {
			lspDiags = append(lspDiags, diagFromHaira(d, text))
		}

		h.mu.Lock()
		h.typeInfos[uri] = typeInfo
		h.mu.Unlock()
	}

	// Publish diagnostics
	h.server.SendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lspDiags,
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// diagFromHaira converts a Haira diagnostic to an LSP diagnostic.
func diagFromHaira(d hairaerr.Diagnostic, text string) Diagnostic {
	severity := 1 // Error
	switch d.Level {
	case hairaerr.Warning:
		severity = 2
	case hairaerr.Info:
		severity = 3
	}

	start := offsetToPosition(text, d.Span.Start)
	end := offsetToPosition(text, d.Span.End)

	return Diagnostic{
		Range:    Range{Start: start, End: end},
		Severity: severity,
		Message:  d.Message,
	}
}

// offsetToPosition converts a byte offset to a Position.
func offsetToPosition(text string, offset int) Position {
	line := 0
	col := 0
	for i := 0; i < len(text) && i < offset; i++ {
		if text[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return Position{Line: line, Character: col}
}

// positionToOffset converts a Position to a byte offset.
func positionToOffset(text string, pos Position) int {
	line := 0
	for i := 0; i < len(text); i++ {
		if line == pos.Line {
			return i + pos.Character
		}
		if text[i] == '\n' {
			line++
		}
	}
	return len(text)
}

// getWordPrefix returns the identifier prefix before the cursor position.
func getWordPrefix(text string, offset int) string {
	if offset > len(text) {
		offset = len(text)
	}
	end := offset
	start := end
	for start > 0 && isIdentChar(text[start-1]) {
		start--
	}
	return text[start:end]
}

// getWordAt returns the full identifier at the given offset.
func getWordAt(text string, offset int) string {
	if offset >= len(text) {
		return ""
	}
	start := offset
	for start > 0 && isIdentChar(text[start-1]) {
		start--
	}
	end := offset
	for end < len(text) && isIdentChar(text[end]) {
		end++
	}
	if start == end {
		return ""
	}
	return text[start:end]
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// spanToLocation converts a Haira span to an LSP Location.
func spanToLocation(uri, text string, span ast.Span) *Location {
	return &Location{
		URI: uri,
		Range: Range{
			Start: offsetToPosition(text, span.Start),
			End:   offsetToPosition(text, span.End),
		},
	}
}

// itemHoverInfo returns hover info for a top-level item.
func itemHoverInfo(item ast.Item) string {
	switch it := item.Node.(type) {
	case ast.FunctionDef:
		return fmt.Sprintf("```haira\nfn %s\n```", funcSignature(it))
	case ast.TypeDef:
		fields := make([]string, len(it.Fields))
		for i, f := range it.Fields {
			ty := "any"
			if f.Ty != nil {
				ty = astTypeString(f.Ty.Node)
			}
			fields[i] = fmt.Sprintf("  %s: %s", f.Name.Node, ty)
		}
		return fmt.Sprintf("```haira\ntype %s {\n%s\n}\n```", it.Name.Node, strings.Join(fields, "\n"))
	case ast.EnumDef:
		variants := make([]string, len(it.Variants))
		for i, v := range it.Variants {
			variants[i] = "  " + v.Name.Node
		}
		return fmt.Sprintf("```haira\nenum %s {\n%s\n}\n```", it.Name.Node, strings.Join(variants, "\n"))
	case ast.ToolDecl:
		return fmt.Sprintf("```haira\ntool %s\n```\n%s", toolSignature(it), it.Description)
	case ast.AgentDecl:
		return fmt.Sprintf("```haira\nagent %s\n```", it.Name.Node)
	case ast.WorkflowDecl:
		return fmt.Sprintf("```haira\nworkflow %s\n```", workflowSignature(it))
	case ast.ProviderDecl:
		return fmt.Sprintf("```haira\nprovider %s\n```", it.Name.Node)
	case ast.TestDecl:
		return fmt.Sprintf("```haira\ntest %q\n```", it.Name.Node)
	}
	return ""
}

func funcSignature(fn ast.FunctionDef) string {
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		ty := "any"
		if p.Ty != nil {
			ty = astTypeString(p.Ty.Node)
		}
		params[i] = p.Name.Node + ": " + ty
	}
	sig := fn.Name.Node + "(" + strings.Join(params, ", ") + ")"
	if fn.ReturnTy != nil {
		sig += " -> " + astTypeString(fn.ReturnTy.Node)
	}
	return sig
}

func toolSignature(tool ast.ToolDecl) string {
	params := make([]string, len(tool.Params))
	for i, p := range tool.Params {
		ty := "any"
		if p.Ty != nil {
			ty = astTypeString(p.Ty.Node)
		}
		params[i] = p.Name.Node + ": " + ty
	}
	sig := tool.Name.Node + "(" + strings.Join(params, ", ") + ")"
	if tool.ReturnTy != nil {
		sig += " -> " + astTypeString(tool.ReturnTy.Node)
	}
	return sig
}

func workflowSignature(wf ast.WorkflowDecl) string {
	params := make([]string, len(wf.Params))
	for i, p := range wf.Params {
		ty := "any"
		if p.Ty != nil {
			ty = astTypeString(p.Ty.Node)
		}
		params[i] = p.Name.Node + ": " + ty
	}
	sig := wf.Name.Node + "(" + strings.Join(params, ", ") + ")"
	if wf.ReturnTy != nil {
		sig += " -> " + astTypeString(wf.ReturnTy.Node)
	}
	return sig
}

func astTypeString(ty ast.Type) string {
	switch t := ty.(type) {
	case ast.NamedType:
		return t.Name
	case ast.ListType:
		return "[" + astTypeString(t.Elem.Node) + "]"
	case ast.MapType:
		return "{" + astTypeString(t.Key.Node) + ": " + astTypeString(t.Value.Node) + "}"
	case ast.OptionType:
		return astTypeString(t.Inner.Node) + "?"
	case ast.FunctionType:
		params := make([]string, len(t.Params))
		for i, p := range t.Params {
			params[i] = astTypeString(p.Node)
		}
		return "fn(" + strings.Join(params, ", ") + ") -> " + astTypeString(t.Ret.Node)
	}
	return "any"
}

// URIToPath converts a file:// URI to a file path.
func URIToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		parsed, err := url.Parse(uri)
		if err == nil {
			return parsed.Path
		}
	}
	return uri
}

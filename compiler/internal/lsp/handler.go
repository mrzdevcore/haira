package lsp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/haira-lang/haira/internal/ast"
	"github.com/haira-lang/haira/internal/checker"
	hairaerr "github.com/haira-lang/haira/internal/errors"
	"github.com/haira-lang/haira/internal/formatter"
	"github.com/haira-lang/haira/internal/lexer"
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
			DefinitionProvider:     true,
			DocumentSymbolProvider: true,
			SignatureHelpProvider: &SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
			DocumentFormattingProvider: true,
		},
		ServerInfo: &ServerInfo{
			Name:    "haira-lsp",
			Version: "0.2.0",
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
	for _, mod := range []string{"io", "http", "json", "time", "env", "postgres", "slack", "excel", "strings", "math", "conv", "fs", "regex", "vector", "github", "gitlab", "sqlite"} {
		if strings.HasPrefix(mod, prefix) {
			items = append(items, CompletionItem{
				Label: mod,
				Kind:  9, // Module
			})
		}
	}

	// Stdlib functions — expanded
	for name, detail := range stdlibFuncsMap {
		if strings.HasPrefix(name, prefix) {
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   3, // Function
				Detail: detail,
			})
		}
	}

	// Context-aware: agent field suggestions inside agent block
	line := getLineAt(text, offset)
	trimmedLine := strings.TrimSpace(line)
	if isInsideAgentBlock(text, offset) && (trimmedLine == "" || !strings.Contains(trimmedLine, ":")) {
		for _, field := range agentFieldCompletions {
			if strings.HasPrefix(field.Label, prefix) {
				items = append(items, field)
			}
		}
	}

	// Context-aware: agent method suggestions after "."
	if offset > 0 && offset <= len(text) && text[offset-1] == '.' {
		dotPrefix := getWordAt(text, offset-2)
		if file != nil && isAgentName(file, dotPrefix) {
			items = append(items,
				CompletionItem{Label: "ask", Kind: 2, Detail: "Send a message and get a response"},
				CompletionItem{Label: "run", Kind: 2, Detail: "Run agent and return AgentResult"},
				CompletionItem{Label: "stream", Kind: 2, Detail: "Stream agent response via SSE"},
			)
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

// DocumentSymbol handles textDocument/documentSymbol.
func (h *Handler) DocumentSymbol(params json.RawMessage) []DocumentSymbol {
	var p DocumentSymbolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return []DocumentSymbol{}
	}

	h.mu.RLock()
	text, ok := h.documents[p.TextDocument.URI]
	file := h.asts[p.TextDocument.URI]
	h.mu.RUnlock()

	if !ok || file == nil {
		return []DocumentSymbol{}
	}

	symbols := []DocumentSymbol{}

	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.FunctionDef:
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         "fn " + funcSignature(it),
				Kind:           SymbolKindFunction,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
			})
		case ast.MethodDef:
			symbols = append(symbols, DocumentSymbol{
				Name:           it.TypeName.Node + "." + it.Name.Node,
				Detail:         "method",
				Kind:           SymbolKindMethod,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
			})
		case ast.TypeDef:
			children := make([]DocumentSymbol, len(it.Fields))
			for i, f := range it.Fields {
				ty := "any"
				if f.Ty != nil {
					ty = astTypeString(f.Ty.Node)
				}
				children[i] = DocumentSymbol{
					Name:           f.Name.Node,
					Detail:         ty,
					Kind:           SymbolKindField,
					Range:          spanToRange(text, f.Span),
					SelectionRange: spanToRange(text, f.Name.Span),
				}
			}
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         fmt.Sprintf("struct (%d fields)", len(it.Fields)),
				Kind:           SymbolKindStruct,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
				Children:       children,
			})
		case ast.EnumDef:
			children := make([]DocumentSymbol, len(it.Variants))
			for i, v := range it.Variants {
				children[i] = DocumentSymbol{
					Name:           v.Name.Node,
					Kind:           SymbolKindEnumMember,
					Range:          spanToRange(text, v.Span),
					SelectionRange: spanToRange(text, v.Name.Span),
				}
			}
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         fmt.Sprintf("enum (%d variants)", len(it.Variants)),
				Kind:           SymbolKindEnum,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
				Children:       children,
			})
		case ast.ToolDecl:
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         "tool " + toolSignature(it),
				Kind:           SymbolKindFunction,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
			})
		case ast.AgentDecl:
			children := make([]DocumentSymbol, 0, len(it.Fields))
			for _, f := range it.Fields {
				children = append(children, DocumentSymbol{
					Name:           f.Key.Node,
					Kind:           SymbolKindProperty,
					Range:          spanToRange(text, f.Key.Span),
					SelectionRange: spanToRange(text, f.Key.Span),
				})
			}
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         "agent",
				Kind:           SymbolKindClass,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
				Children:       children,
			})
		case ast.WorkflowDecl:
			detail := "workflow " + workflowSignature(it)
			if it.Trigger != nil {
				detail = "@" + it.Trigger.Name.Node + " " + detail
			}
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         detail,
				Kind:           SymbolKindFunction,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
			})
		case ast.ProviderDecl:
			providerType := ""
			for _, f := range it.Fields {
				if f.Key.Node == "type" {
					providerType = exprStringValue(f.Value)
					break
				}
			}
			detail := "provider"
			if providerType != "" {
				detail += " (" + providerType + ")"
			}
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         detail,
				Kind:           SymbolKindModule,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
			})
		case ast.TestDecl:
			symbols = append(symbols, DocumentSymbol{
				Name:           it.Name.Node,
				Detail:         "test",
				Kind:           SymbolKindEvent,
				Range:          spanToRange(text, item.Span),
				SelectionRange: spanToRange(text, it.Name.Span),
			})
		}
	}

	return symbols
}

// SignatureHelp handles textDocument/signatureHelp.
func (h *Handler) SignatureHelp(params json.RawMessage) *SignatureHelp {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil
	}

	h.mu.RLock()
	text, ok := h.documents[p.TextDocument.URI]
	file := h.asts[p.TextDocument.URI]
	h.mu.RUnlock()

	if !ok || file == nil {
		return nil
	}

	offset := positionToOffset(text, p.Position)

	// Walk backwards from cursor to find the opening '(' and count commas
	depth := 0
	commas := 0
	parenStart := -1
	for i := offset - 1; i >= 0; i-- {
		switch text[i] {
		case ')':
			depth++
		case '(':
			if depth > 0 {
				depth--
			} else {
				parenStart = i
			}
		case ',':
			if depth == 0 {
				commas++
			}
		}
		if parenStart >= 0 {
			break
		}
	}

	if parenStart < 0 {
		return nil
	}

	// Get the function name before the '('
	funcName := getWordAt(text, parenStart-1)
	if funcName == "" {
		return nil
	}

	// Look up the function/tool/workflow in the AST
	for _, item := range file.Items {
		switch it := item.Node.(type) {
		case ast.FunctionDef:
			if it.Name.Node == funcName {
				return buildSignatureHelp(funcSignature(it), it.Params, commas)
			}
		case ast.ToolDecl:
			if it.Name.Node == funcName {
				return buildSignatureHelp(toolSignature(it), it.Params, commas)
			}
		case ast.WorkflowDecl:
			if it.Name.Node == funcName {
				return buildSignatureHelp(workflowSignature(it), it.Params, commas)
			}
		case ast.MethodDef:
			if it.Name.Node == funcName {
				sig := methodSignature(it)
				return buildSignatureHelp(sig, it.Params, commas)
			}
		}
	}

	// Check stdlib signatures
	if sig, ok := stdlibSignatures[funcName]; ok {
		return &SignatureHelp{
			Signatures: []SignatureInformation{
				{Label: sig},
			},
			ActiveSignature: 0,
			ActiveParameter: commas,
		}
	}

	return nil
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

	// Type check (only if parse succeeded)
	var typeInfo *checker.TypeInfo
	if file != nil && len(parseErrs) == 0 {
		var typeDiags []hairaerr.Diagnostic
		typeInfo, typeDiags = checker.Check(file)
		for _, d := range typeDiags {
			lspDiags = append(lspDiags, diagFromHaira(d, text))
		}
	}

	// Update AST and type info atomically under a single lock
	h.mu.Lock()
	if file != nil {
		h.asts[uri] = file
	}
	if typeInfo != nil {
		h.typeInfos[uri] = typeInfo
	}
	h.mu.Unlock()

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
// LSP positions use UTF-16 code units for the character offset.
// For ASCII-only text, UTF-16 offset == byte offset.
// For multi-byte characters, we count UTF-16 code units properly.
func positionToOffset(text string, pos Position) int {
	line := 0
	i := 0
	// Find the start of the target line
	for i < len(text) {
		if line == pos.Line {
			break
		}
		if text[i] == '\n' {
			line++
		}
		i++
	}
	if line != pos.Line {
		return len(text)
	}
	// Now advance by pos.Character UTF-16 code units
	utf16Col := 0
	for i < len(text) && text[i] != '\n' && utf16Col < pos.Character {
		r := rune(text[i])
		size := 1
		if r >= 0x80 {
			r, size = decodeRune(text[i:])
		}
		// Characters in the BMP are 1 UTF-16 unit; supplementary are 2
		if r >= 0x10000 {
			utf16Col += 2
		} else {
			utf16Col++
		}
		i += size
	}
	return i
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
		desc := ""
		if it.Description != "" {
			desc = "\n\n" + it.Description
		}
		return fmt.Sprintf("```haira\ntool %s\n```%s", toolSignature(it), desc)
	case ast.AgentDecl:
		return agentHoverInfo(it)
	case ast.WorkflowDecl:
		return workflowHoverInfo(it)
	case ast.ProviderDecl:
		return providerHoverInfo(it)
	case ast.TestDecl:
		return fmt.Sprintf("```haira\ntest %q\n```", it.Name.Node)
	}
	return ""
}

func agentHoverInfo(agent ast.AgentDecl) string {
	lines := []string{fmt.Sprintf("```haira\nagent %s\n```", agent.Name.Node)}

	for _, f := range agent.Fields {
		switch f.Key.Node {
		case "provider":
			if v := exprStringValue(f.Value); v != "" {
				lines = append(lines, "**Provider:** "+v)
			} else {
				lines = append(lines, "**Provider:** "+exprIdentValue(f.Value))
			}
		case "system":
			if v := exprStringValue(f.Value); v != "" {
				// Truncate long system prompts
				if len(v) > 120 {
					v = v[:120] + "…"
				}
				lines = append(lines, "**System:** "+v)
			}
		case "tools":
			lines = append(lines, "**Tools:** "+exprListString(f.Value))
		case "scope":
			if v := exprStringValue(f.Value); v != "" {
				lines = append(lines, "**Scope:** "+v)
			}
		case "memory":
			lines = append(lines, "**Memory:** "+exprRawString(f.Value))
		case "temperature":
			lines = append(lines, "**Temperature:** "+exprRawString(f.Value))
		case "model":
			if v := exprStringValue(f.Value); v != "" {
				lines = append(lines, "**Model:** "+v)
			}
		}
	}

	return strings.Join(lines, "\n\n")
}

func workflowHoverInfo(wf ast.WorkflowDecl) string {
	header := "workflow " + workflowSignature(wf)
	if wf.Trigger != nil {
		header = "@" + wf.Trigger.Name.Node + " " + header
	}
	lines := []string{fmt.Sprintf("```haira\n%s\n```", header)}

	if wf.Description != "" {
		lines = append(lines, wf.Description)
	}

	// List decorators
	for _, d := range wf.Decorators {
		lines = append(lines, "**@"+d.Name.Node+"**")
	}

	return strings.Join(lines, "\n\n")
}

func providerHoverInfo(prov ast.ProviderDecl) string {
	provType := ""
	model := ""
	for _, f := range prov.Fields {
		switch f.Key.Node {
		case "type":
			provType = exprStringValue(f.Value)
		case "model":
			model = exprStringValue(f.Value)
		}
	}
	lines := []string{fmt.Sprintf("```haira\nprovider %s\n```", prov.Name.Node)}
	if provType != "" {
		lines = append(lines, "**Type:** "+provType)
	}
	if model != "" {
		lines = append(lines, "**Model:** "+model)
	}
	return strings.Join(lines, "\n\n")
}

// exprIdentValue extracts an identifier name from an expression.
func exprIdentValue(expr ast.Expr) string {
	if ident, ok := expr.Node.(ast.IdentExpr); ok {
		return ident.Name
	}
	return ""
}

// exprListString returns a comma-separated string of identifiers in a list expression.
func exprListString(expr ast.Expr) string {
	if list, ok := expr.Node.(ast.ListExpr); ok {
		names := make([]string, 0, len(list.Elems))
		for _, elem := range list.Elems {
			if ident, ok := elem.Node.(ast.IdentExpr); ok {
				names = append(names, ident.Name)
			}
		}
		return "[" + strings.Join(names, ", ") + "]"
	}
	return exprRawString(expr)
}

// exprRawString returns a basic string representation of an expression.
func exprRawString(expr ast.Expr) string {
	switch e := expr.Node.(type) {
	case ast.LiteralExpr:
		switch lit := e.Lit.(type) {
		case ast.StringLit:
			return lit.Value
		case ast.IntLit:
			return fmt.Sprintf("%d", lit.Value)
		case ast.FloatLit:
			return fmt.Sprintf("%g", lit.Value)
		case ast.BoolLit:
			if lit.Value {
				return "true"
			}
			return "false"
		}
	case ast.IdentExpr:
		return e.Name
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

// spanToRange converts a Haira span to an LSP Range.
func spanToRange(text string, span ast.Span) Range {
	return Range{
		Start: offsetToPosition(text, span.Start),
		End:   offsetToPosition(text, span.End),
	}
}

func buildSignatureHelp(sig string, params []ast.Param, activeParam int) *SignatureHelp {
	paramInfos := make([]ParameterInformation, len(params))
	for i, p := range params {
		ty := "any"
		if p.Ty != nil {
			ty = astTypeString(p.Ty.Node)
		}
		paramInfos[i] = ParameterInformation{
			Label: p.Name.Node + ": " + ty,
		}
	}
	return &SignatureHelp{
		Signatures: []SignatureInformation{
			{
				Label:      sig,
				Parameters: paramInfos,
			},
		},
		ActiveSignature: 0,
		ActiveParameter: activeParam,
	}
}

func methodSignature(m ast.MethodDef) string {
	params := make([]string, len(m.Params))
	for i, p := range m.Params {
		ty := "any"
		if p.Ty != nil {
			ty = astTypeString(p.Ty.Node)
		}
		params[i] = p.Name.Node + ": " + ty
	}
	sig := m.TypeName.Node + "." + m.Name.Node + "(" + strings.Join(params, ", ") + ")"
	if m.ReturnTy != nil {
		sig += " -> " + astTypeString(m.ReturnTy.Node)
	}
	return sig
}

// exprStringValue extracts a string value from a string literal expression.
func exprStringValue(expr ast.Expr) string {
	if litExpr, ok := expr.Node.(ast.LiteralExpr); ok {
		if str, ok := litExpr.Lit.(ast.StringLit); ok {
			return str.Value
		}
	}
	return ""
}

// stdlibFuncsMap holds all documented stdlib functions for completion.
var stdlibFuncsMap = map[string]string{
	// io
	"io.println": "Print a value with newline",
	"io.print":   "Print a value",
	"io.readln":  "Read a line from stdin",
	"io.sprintf": "Format a string",
	// http
	"http.get":    "HTTP GET request",
	"http.post":   "HTTP POST request",
	"http.put":    "HTTP PUT request",
	"http.delete": "HTTP DELETE request",
	// json
	"json.parse":     "Parse JSON string to value",
	"json.stringify": "Convert value to JSON string",
	// time
	"time.now":    "Current timestamp",
	"time.sleep":  "Sleep for duration",
	"time.parse":  "Parse time string",
	"time.format": "Format time to string",
	// strings
	"strings.upper":   "Convert to uppercase",
	"strings.lower":   "Convert to lowercase",
	"strings.trim":    "Trim whitespace",
	"strings.replace": "Replace substring",
	// math
	"math.abs":   "Absolute value",
	"math.min":   "Minimum of values",
	"math.max":   "Maximum of values",
	"math.floor": "Floor (round down)",
	"math.ceil":  "Ceiling (round up)",
	// conv
	"conv.to_int":    "Convert to integer",
	"conv.to_float":  "Convert to float",
	"conv.to_string": "Convert to string",
	// builtins
	"len":      "Get length of a collection",
	"keys":     "Get keys of a map",
	"values":   "Get values of a map",
	"join":     "Join list elements with separator",
	"split":    "Split string by separator",
	"contains": "Check if string contains substring",
	"append":   "Append items to list",
	"push":     "Push item to list",
	"pop":      "Pop item from list",
	"env":      "Read environment variable",
}

// agentFieldCompletions suggests fields inside agent {} blocks.
var agentFieldCompletions = []CompletionItem{
	{Label: "provider", Kind: 7, Detail: "LLM provider reference"},
	{Label: "system", Kind: 7, Detail: "System prompt"},
	{Label: "tools", Kind: 7, Detail: "List of tools available to the agent"},
	{Label: "model", Kind: 7, Detail: "Model name override"},
	{Label: "temperature", Kind: 7, Detail: "Temperature (0.0 - 1.0)"},
	{Label: "memory", Kind: 7, Detail: "Memory mode: conversation, summary, none"},
	{Label: "handoffs", Kind: 7, Detail: "List of agents for handoff"},
	{Label: "scope", Kind: 7, Detail: "Topic restriction for guardrails"},
	{Label: "scope_deny", Kind: 7, Detail: "Message returned for off-topic requests"},
	{Label: "output", Kind: 7, Detail: "Structured output type"},
}

// isAgentName checks if a name is a declared agent in the file.
func isAgentName(file *ast.SourceFile, name string) bool {
	for _, item := range file.Items {
		if agent, ok := item.Node.(ast.AgentDecl); ok {
			if agent.Name.Node == name {
				return true
			}
		}
	}
	return false
}

// isInsideAgentBlock checks if cursor is inside an agent { } block.
func isInsideAgentBlock(text string, offset int) bool {
	// Simple heuristic: walk backwards to find "agent <Name> {" before a matching "}"
	depth := 0
	for i := offset - 1; i >= 0; i-- {
		switch text[i] {
		case '}':
			depth++
		case '{':
			if depth > 0 {
				depth--
			} else {
				// Found our opening brace — check if preceded by "agent <Name>"
				before := strings.TrimSpace(text[:i])
				// Look for "agent <word>" pattern
				parts := strings.Fields(before)
				if len(parts) >= 2 && parts[len(parts)-2] == "agent" {
					return true
				}
				return false
			}
		}
	}
	return false
}

// getLineAt returns the line of text at the given offset.
func getLineAt(text string, offset int) string {
	if offset > len(text) {
		offset = len(text)
	}
	start := offset
	for start > 0 && text[start-1] != '\n' {
		start--
	}
	end := offset
	for end < len(text) && text[end] != '\n' {
		end++
	}
	return text[start:end]
}

// stdlibSignatures holds signatures for common stdlib functions.
var stdlibSignatures = map[string]string{
	"println": "io.println(value: any)",
	"print":   "io.print(value: any)",
	"readln":  "io.readln(prompt: string) -> string",
	"sprintf": "io.sprintf(format: string, args: ...any) -> string",
	"len":     "len(collection: any) -> int",
	"keys":    "keys(map: {string: any}) -> [string]",
	"values":  "values(map: {string: any}) -> [any]",
	"join":    "join(list: [string], sep: string) -> string",
	"split":   "split(s: string, sep: string) -> [string]",
	"contains": "contains(s: string, substr: string) -> bool",
	"append":  "append(list: [any], items: ...any) -> [any]",
	"push":    "push(list: [any], item: any)",
	"pop":     "pop(list: [any]) -> any",
	"env":     "env(key: string) -> string",
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

// decodeRune decodes a single UTF-8 rune from a string slice.
func decodeRune(s string) (rune, int) {
	return utf8.DecodeRuneInString(s)
}

// escapeMarkdown escapes characters that have special meaning in markdown.
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		`|`, `\|`,
		`<`, `&lt;`,
		`>`, `&gt;`,
	)
	return replacer.Replace(s)
}

// Formatting handles textDocument/formatting.
func (h *Handler) Formatting(params json.RawMessage) []TextEdit {
	var p DocumentFormattingParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil
	}

	h.mu.RLock()
	text, ok := h.documents[p.TextDocument.URI]
	h.mu.RUnlock()

	if !ok {
		return nil
	}

	// Lex and parse
	tokens := lexer.New(text).AllTokens()
	file, errs := parser.Parse(text)
	if len(errs) > 0 || file == nil {
		return nil // Don't format files with parse errors
	}

	formatted := formatter.Format(text, tokens, file)
	if formatted == text {
		return nil // No changes needed
	}

	// Return a single edit replacing the entire document
	lines := strings.Count(text, "\n")
	lastLineLen := len(text) - strings.LastIndex(text, "\n") - 1
	if lastLineLen < 0 {
		lastLineLen = len(text)
	}

	return []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: lines, Character: lastLineLen},
			},
			NewText: formatted,
		},
	}
}

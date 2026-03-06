// Package lsp implements a Language Server Protocol server for Haira.
// It uses JSON-RPC 2.0 over stdin/stdout with no external dependencies.
package lsp

// --- JSON-RPC 2.0 types ---

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"` // number or string; nil for notifications
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
// Note: Result uses a pointer to json.RawMessage so we can distinguish
// "result: null" from omitting "result" entirely.
type Response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      interface{}    `json:"id"`
	Result  interface{}    `json:"result"`
	Error   *ResponseError `json:"error,omitempty"`
}

// ResponseError is a JSON-RPC 2.0 error.
type ResponseError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Notification is a JSON-RPC 2.0 notification (no id, no response expected).
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// --- LSP protocol types ---

// InitializeParams is the params for initialize request.
type InitializeParams struct {
	ProcessID    *int               `json:"processId"`
	RootURI      string             `json:"rootUri"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

type ClientCapabilities struct {
	TextDocument *TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Completion *CompletionClientCapabilities `json:"completion,omitempty"`
	Hover      *HoverClientCapabilities      `json:"hover,omitempty"`
}

type CompletionClientCapabilities struct{}
type HoverClientCapabilities struct{}

// InitializeResult is the result for initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ServerCapabilities struct {
	TextDocumentSync       int                   `json:"textDocumentSync"` // 1 = Full
	HoverProvider          bool                  `json:"hoverProvider,omitempty"`
	CompletionProvider     *CompletionOptions    `json:"completionProvider,omitempty"`
	DefinitionProvider     bool                  `json:"definitionProvider,omitempty"`
	DocumentSymbolProvider bool                  `json:"documentSymbolProvider,omitempty"`
	SignatureHelpProvider  *SignatureHelpOptions  `json:"signatureHelpProvider,omitempty"`
	DocumentFormattingProvider bool              `json:"documentFormattingProvider,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// TextDocumentItem represents an opened document.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenTextDocumentParams is for textDocument/didOpen.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams is for textDocument/didChange.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"` // full content (TextDocumentSync = Full)
}

// DidCloseTextDocumentParams is for textDocument/didClose.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DidSaveTextDocumentParams is for textDocument/didSave.
type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

// TextDocumentIdentifier identifies a document by URI.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// TextDocumentPositionParams is used for hover and goto definition.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Position is a zero-based line/character position.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a range in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location is a URI + range.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic represents a compiler error or warning.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=Error, 2=Warning, 3=Info, 4=Hint
	Message  string `json:"message"`
}

// PublishDiagnosticsParams is for textDocument/publishDiagnostics.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Hover is the result for textDocument/hover.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "plaintext" or "markdown"
	Value string `json:"value"`
}

// CompletionParams is the params for textDocument/completion.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"` // 1=Text, 2=Method, 3=Function, 6=Variable, 7=Class, 14=Keyword, 9=Module
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

// CompletionList is the result for textDocument/completion.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// DocumentSymbol represents a symbol in a document (for outline/breadcrumb).
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// LSP SymbolKind constants.
const (
	SymbolKindFunction    = 12
	SymbolKindStruct      = 23
	SymbolKindEnum        = 10
	SymbolKindEnumMember  = 22
	SymbolKindField       = 8
	SymbolKindMethod      = 6
	SymbolKindVariable    = 13
	SymbolKindConstant    = 14
	SymbolKindClass       = 5
	SymbolKindModule      = 2
	SymbolKindProperty    = 7
	SymbolKindEvent       = 24
)

// DocumentSymbolParams is for textDocument/documentSymbol.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SignatureHelp is the result for textDocument/signatureHelp.
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

// SignatureInformation represents a callable signature.
type SignatureInformation struct {
	Label      string                 `json:"label"`
	Parameters []ParameterInformation `json:"parameters,omitempty"`
}

// ParameterInformation represents a parameter of a callable.
type ParameterInformation struct {
	Label string `json:"label"`
}

// DocumentFormattingParams is for textDocument/formatting.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions describes formatting options.
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// TextEdit represents a change to a text document.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

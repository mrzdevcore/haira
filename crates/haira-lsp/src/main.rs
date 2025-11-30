//! Haira Language Server
//!
//! A Language Server Protocol implementation for the Haira programming language.

use dashmap::DashMap;
use ropey::Rope;
use tower_lsp::jsonrpc::Result;
use tower_lsp::lsp_types::*;
use tower_lsp::{Client, LanguageServer, LspService, Server};

mod analysis;
mod completion;
mod diagnostics;
mod hover;
mod symbols;

use diagnostics::collect_diagnostics;

/// Document state stored by the server.
#[derive(Debug)]
struct Document {
    /// The document content as a rope for efficient editing.
    content: Rope,
    /// The document version.
    version: i32,
}

/// The Haira language server.
struct HairaLanguageServer {
    /// The LSP client for sending notifications.
    client: Client,
    /// Open documents.
    documents: DashMap<Url, Document>,
}

impl HairaLanguageServer {
    fn new(client: Client) -> Self {
        Self {
            client,
            documents: DashMap::new(),
        }
    }

    /// Get the content of a document.
    fn get_document_content(&self, uri: &Url) -> Option<String> {
        self.documents.get(uri).map(|doc| doc.content.to_string())
    }

    /// Analyze a document and publish diagnostics.
    async fn analyze_document(&self, uri: &Url) {
        let content = match self.get_document_content(uri) {
            Some(c) => c,
            None => return,
        };

        let diagnostics = collect_diagnostics(&content);
        self.client
            .publish_diagnostics(uri.clone(), diagnostics, None)
            .await;
    }
}

#[tower_lsp::async_trait]
impl LanguageServer for HairaLanguageServer {
    async fn initialize(&self, _: InitializeParams) -> Result<InitializeResult> {
        Ok(InitializeResult {
            capabilities: ServerCapabilities {
                text_document_sync: Some(TextDocumentSyncCapability::Options(
                    TextDocumentSyncOptions {
                        open_close: Some(true),
                        change: Some(TextDocumentSyncKind::FULL),
                        save: Some(TextDocumentSyncSaveOptions::SaveOptions(SaveOptions {
                            include_text: Some(true),
                        })),
                        ..Default::default()
                    },
                )),
                completion_provider: Some(CompletionOptions {
                    trigger_characters: Some(vec![".".to_string(), ":".to_string()]),
                    resolve_provider: Some(false),
                    ..Default::default()
                }),
                hover_provider: Some(HoverProviderCapability::Simple(true)),
                document_symbol_provider: Some(OneOf::Left(true)),
                definition_provider: Some(OneOf::Left(true)),
                references_provider: Some(OneOf::Left(true)),
                document_formatting_provider: Some(OneOf::Left(true)),
                ..Default::default()
            },
            server_info: Some(ServerInfo {
                name: "haira-lsp".to_string(),
                version: Some(env!("CARGO_PKG_VERSION").to_string()),
            }),
        })
    }

    async fn initialized(&self, _: InitializedParams) {
        self.client
            .log_message(MessageType::INFO, "Haira language server initialized")
            .await;
    }

    async fn shutdown(&self) -> Result<()> {
        Ok(())
    }

    async fn did_open(&self, params: DidOpenTextDocumentParams) {
        let uri = params.text_document.uri;
        let content = params.text_document.text;
        let version = params.text_document.version;

        self.documents.insert(
            uri.clone(),
            Document {
                content: Rope::from_str(&content),
                version,
            },
        );

        self.analyze_document(&uri).await;
    }

    async fn did_change(&self, params: DidChangeTextDocumentParams) {
        let uri = params.text_document.uri;

        if let Some(mut doc) = self.documents.get_mut(&uri) {
            // Full sync - replace entire content
            if let Some(change) = params.content_changes.into_iter().next() {
                doc.content = Rope::from_str(&change.text);
                doc.version = params.text_document.version;
            }
        }

        self.analyze_document(&uri).await;
    }

    async fn did_save(&self, params: DidSaveTextDocumentParams) {
        if let Some(text) = params.text {
            if let Some(mut doc) = self.documents.get_mut(&params.text_document.uri) {
                doc.content = Rope::from_str(&text);
            }
        }
        self.analyze_document(&params.text_document.uri).await;
    }

    async fn did_close(&self, params: DidCloseTextDocumentParams) {
        self.documents.remove(&params.text_document.uri);
        // Clear diagnostics
        self.client
            .publish_diagnostics(params.text_document.uri, vec![], None)
            .await;
    }

    async fn completion(&self, params: CompletionParams) -> Result<Option<CompletionResponse>> {
        let uri = &params.text_document_position.text_document.uri;
        let position = params.text_document_position.position;

        let content = match self.get_document_content(uri) {
            Some(c) => c,
            None => return Ok(None),
        };

        let completions = completion::get_completions(&content, position);
        Ok(Some(CompletionResponse::Array(completions)))
    }

    async fn hover(&self, params: HoverParams) -> Result<Option<Hover>> {
        let uri = &params.text_document_position_params.text_document.uri;
        let position = params.text_document_position_params.position;

        let content = match self.get_document_content(uri) {
            Some(c) => c,
            None => return Ok(None),
        };

        Ok(hover::get_hover(&content, position))
    }

    async fn document_symbol(
        &self,
        params: DocumentSymbolParams,
    ) -> Result<Option<DocumentSymbolResponse>> {
        let uri = &params.text_document.uri;

        let content = match self.get_document_content(uri) {
            Some(c) => c,
            None => return Ok(None),
        };

        let symbols = symbols::get_document_symbols(&content);
        Ok(Some(DocumentSymbolResponse::Flat(symbols)))
    }

    async fn goto_definition(
        &self,
        params: GotoDefinitionParams,
    ) -> Result<Option<GotoDefinitionResponse>> {
        let uri = &params.text_document_position_params.text_document.uri;
        let position = params.text_document_position_params.position;

        let content = match self.get_document_content(uri) {
            Some(c) => c,
            None => return Ok(None),
        };

        if let Some(location) = analysis::find_definition(&content, position, uri.clone()) {
            Ok(Some(GotoDefinitionResponse::Scalar(location)))
        } else {
            Ok(None)
        }
    }

    async fn references(&self, params: ReferenceParams) -> Result<Option<Vec<Location>>> {
        let uri = &params.text_document_position.text_document.uri;
        let position = params.text_document_position.position;

        let content = match self.get_document_content(uri) {
            Some(c) => c,
            None => return Ok(None),
        };

        let references = analysis::find_references(&content, position, uri.clone());
        if references.is_empty() {
            Ok(None)
        } else {
            Ok(Some(references))
        }
    }

    async fn formatting(&self, params: DocumentFormattingParams) -> Result<Option<Vec<TextEdit>>> {
        let uri = &params.text_document.uri;

        let _content = match self.get_document_content(uri) {
            Some(c) => c,
            None => return Ok(None),
        };

        // TODO: Implement formatting
        Ok(None)
    }
}

#[tokio::main]
async fn main() {
    env_logger::init();

    let stdin = tokio::io::stdin();
    let stdout = tokio::io::stdout();

    let (service, socket) = LspService::new(HairaLanguageServer::new);
    Server::new(stdin, stdout, socket).serve(service).await;
}

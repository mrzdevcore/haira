//! Local AI backend for Haira using llama.cpp.
//!
//! This crate provides a self-contained local AI experience where Haira
//! manages its own llama-server process and model downloads.

mod client;
mod error;
mod model;
pub mod paths;
mod server;

pub use client::LlamaCppClient;
pub use error::LocalAIError;
pub use model::{ModelInfo, ModelManager};
pub use server::LlamaCppServer;

/// Default port for the local llama-server instance.
pub const DEFAULT_PORT: u16 = 11435;

/// Default model name for Haira code generation.
pub const DEFAULT_MODEL_NAME: &str = "haira-coder-3b";

/// Default model filename.
pub const DEFAULT_MODEL_FILENAME: &str = "haira-coder-3b.gguf";

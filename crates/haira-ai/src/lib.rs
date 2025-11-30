//! # Haira AI Integration
//!
//! This crate provides integration with Claude AI for interpreting developer
//! intent and generating code implementations.
//!
//! ## Architecture
//!
//! ```text
//! ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
//! │  Unresolved     │ --> │    AI Engine    │ --> │  CIR Function   │
//! │  Function Call  │     │  (Claude API)   │     │  (Generated)    │
//! └─────────────────┘     └─────────────────┘     └─────────────────┘
//!                               │
//!                         ┌─────┴─────┐
//!                         │   Cache   │
//!                         └───────────┘
//! ```
//!
//! ## Usage
//!
//! ```ignore
//! use haira_ai::{AIEngine, AIConfig};
//!
//! let config = AIConfig::from_env();
//! let engine = AIEngine::new(config);
//!
//! let request = AIRequest { ... };
//! let response = engine.interpret(request).await?;
//! ```

mod cache;
mod client;
mod config;
mod engine;
mod ollama;
mod prompt;

pub use cache::AICache;
pub use client::ClaudeClient;
pub use config::AIConfig;
pub use engine::{AIBackend, AIEngine, AIError};
pub use ollama::{OllamaClient, OllamaError, DEFAULT_OLLAMA_MODEL, DEFAULT_OLLAMA_URL};

// Re-export CIR types for convenience
pub use haira_cir::{
    AIRequest, AIResponse, CIRFunction, CIROperation, CIRType, CIRValue, InterpretationContext,
    TypeDefinition,
};

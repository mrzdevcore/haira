//! # Haira AI Integration
//!
//! This crate provides AI integration for interpreting developer intent
//! and generating code implementations.
//!
//! ## Architecture
//!
//! ```text
//! ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
//! │  Unresolved     │ --> │    AI Engine    │ --> │  CIR Function   │
//! │  Function Call  │     │  (Local/Ollama) │     │  (Generated)    │
//! └─────────────────┘     └─────────────────┘     └─────────────────┘
//!                               │
//!                         ┌─────┴─────┐
//!                         │   Cache   │
//!                         └───────────┘
//! ```
//!
//! ## AI Backends
//!
//! - **Local AI** (primary) - Uses llama.cpp with local models
//! - **Ollama** (fallback) - Uses Ollama server
//!
//! ## Usage
//!
//! ```ignore
//! use haira_ai::{AIEngine, AIConfig};
//!
//! let config = AIConfig::default();
//! let engine = AIEngine::with_local_ai(config, None);
//!
//! let func = engine.interpret("get_active_users", context).await?;
//! ```

mod cache;
mod config;
mod engine;
pub mod hif;
mod ollama;
mod prompt;

pub use cache::AICache;
pub use config::AIConfig;
pub use engine::{AIBackend, AIEngine, AIError};
pub use ollama::{OllamaClient, OllamaError, DEFAULT_OLLAMA_MODEL, DEFAULT_OLLAMA_URL};

// Re-export local AI types
pub use haira_local_ai::{
    paths as local_ai_paths, LlamaCppClient, LlamaCppServer, LocalAIError, ModelInfo, ModelManager,
    DEFAULT_MODEL_FILENAME, DEFAULT_MODEL_NAME, DEFAULT_PORT as DEFAULT_LOCAL_AI_PORT,
};

// Re-export CIR types for convenience
pub use haira_cir::{
    AIRequest, AIResponse, CIRFunction, CIROperation, CIRType, CIRValue, InterpretationContext,
    TypeDefinition,
};

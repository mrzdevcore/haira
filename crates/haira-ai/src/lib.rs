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

mod client;
mod engine;
mod cache;
mod prompt;
mod config;

pub use client::ClaudeClient;
pub use engine::AIEngine;
pub use cache::AICache;
pub use config::AIConfig;

// Re-export CIR types for convenience
pub use haira_cir::{
    AIRequest, AIResponse, CIRFunction, CIROperation, CIRType, CIRValue,
    InterpretationContext, TypeDefinition,
};

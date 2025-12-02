//! Error types for local AI operations.

use thiserror::Error;

/// Errors that can occur during local AI operations.
#[derive(Debug, Error)]
pub enum LocalAIError {
    /// HTTP request failed.
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),

    /// JSON serialization/deserialization failed.
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),

    /// Server returned an error response.
    #[error("API error: {0}")]
    Api(String),

    /// Server is not running or not reachable.
    #[error("Server not running at {0}. Start it with: haira serve")]
    ServerNotRunning(String),

    /// Server failed to start.
    #[error("Failed to start server: {0}")]
    ServerStartFailed(String),

    /// Server process died unexpectedly.
    #[error("Server process died: {0}")]
    ServerDied(String),

    /// Model not found locally.
    #[error("Model '{0}' not found. Download it with: haira model pull")]
    ModelNotFound(String),

    /// Model download failed.
    #[error("Failed to download model: {0}")]
    DownloadFailed(String),

    /// Checksum verification failed.
    #[error("Model checksum mismatch. Expected: {expected}, got: {actual}")]
    ChecksumMismatch { expected: String, actual: String },

    /// llama-server binary not found.
    #[error("llama-server binary not found at {0}")]
    ServerBinaryNotFound(String),

    /// I/O error.
    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    /// Failed to create Haira data directory.
    #[error("Failed to create data directory: {0}")]
    DataDirCreationFailed(String),

    /// Timeout waiting for server to start.
    #[error("Timeout waiting for server to become ready")]
    ServerStartTimeout,
}

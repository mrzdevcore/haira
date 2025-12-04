//! AI configuration.

use std::path::PathBuf;

/// Configuration for the AI engine.
#[derive(Debug, Clone)]
pub struct AIConfig {
    /// Cache directory
    pub cache_dir: PathBuf,
    /// Whether to use cache
    pub use_cache: bool,
    /// Minimum confidence to accept (0.0 - 1.0)
    pub min_confidence: f64,
    /// Ollama model name (for Ollama backend)
    pub ollama_model: Option<String>,
    /// Local AI model filename (for Local AI backend)
    pub local_model: Option<String>,
}

impl Default for AIConfig {
    fn default() -> Self {
        Self {
            cache_dir: PathBuf::from(".haira-cache/ai"),
            // Disabled by default - HIF caching is used at the build level instead
            use_cache: false,
            min_confidence: 0.5,
            ollama_model: None,
            local_model: None,
        }
    }
}

impl AIConfig {
    /// Create config from environment variables.
    pub fn from_env() -> Self {
        let cache_dir = std::env::var("HAIRA_CACHE_DIR")
            .map(PathBuf::from)
            .unwrap_or_else(|_| PathBuf::from(".haira-cache/ai"));

        // Disabled by default - HIF caching is used at the build level instead
        let use_cache = std::env::var("HAIRA_AI_CACHE")
            .map(|v| v == "1" || v.to_lowercase() == "true")
            .unwrap_or(false);

        let min_confidence = std::env::var("HAIRA_AI_MIN_CONFIDENCE")
            .ok()
            .and_then(|v| v.parse().ok())
            .unwrap_or(0.5);

        let ollama_model = std::env::var("HAIRA_OLLAMA_MODEL").ok();
        let local_model = std::env::var("HAIRA_LOCAL_MODEL").ok();

        Self {
            cache_dir,
            use_cache,
            min_confidence,
            ollama_model,
            local_model,
        }
    }

    /// Create a builder for configuration.
    pub fn builder() -> AIConfigBuilder {
        AIConfigBuilder::default()
    }
}

/// Builder for AI configuration.
#[derive(Debug, Default)]
pub struct AIConfigBuilder {
    config: AIConfig,
}

impl AIConfigBuilder {
    pub fn cache_dir(mut self, path: impl Into<PathBuf>) -> Self {
        self.config.cache_dir = path.into();
        self
    }

    pub fn use_cache(mut self, use_cache: bool) -> Self {
        self.config.use_cache = use_cache;
        self
    }

    pub fn min_confidence(mut self, confidence: f64) -> Self {
        self.config.min_confidence = confidence;
        self
    }

    pub fn ollama_model(mut self, model: impl Into<String>) -> Self {
        self.config.ollama_model = Some(model.into());
        self
    }

    pub fn local_model(mut self, model: impl Into<String>) -> Self {
        self.config.local_model = Some(model.into());
        self
    }

    pub fn build(self) -> AIConfig {
        self.config
    }
}

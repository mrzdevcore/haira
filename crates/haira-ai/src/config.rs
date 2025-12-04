//! AI configuration.

use std::path::PathBuf;

/// Configuration for the AI engine.
#[derive(Debug, Clone)]
pub struct AIConfig {
    /// Anthropic API key
    pub api_key: String,
    /// Model to use (default: claude-3-5-sonnet-20241022)
    pub model: String,
    /// Maximum tokens in response
    pub max_tokens: u32,
    /// Cache directory
    pub cache_dir: PathBuf,
    /// Whether to use cache
    pub use_cache: bool,
    /// Minimum confidence to accept (0.0 - 1.0)
    pub min_confidence: f64,
}

impl Default for AIConfig {
    fn default() -> Self {
        Self {
            api_key: String::new(),
            model: "claude-3-5-sonnet-20241022".to_string(),
            max_tokens: 4096,
            cache_dir: PathBuf::from(".haira-cache/ai"),
            // Disabled by default - HIF caching is used at the build level instead
            use_cache: false,
            min_confidence: 0.5,
        }
    }
}

impl AIConfig {
    /// Create config from environment variables.
    pub fn from_env() -> Self {
        let api_key = std::env::var("ANTHROPIC_API_KEY")
            .or_else(|_| std::env::var("CLAUDE_API_KEY"))
            .unwrap_or_default();

        let model = std::env::var("HAIRA_AI_MODEL")
            .unwrap_or_else(|_| "claude-3-5-sonnet-20241022".to_string());

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

        Self {
            api_key,
            model,
            max_tokens: 4096,
            cache_dir,
            use_cache,
            min_confidence,
        }
    }

    /// Check if the config is valid (has API key).
    pub fn is_valid(&self) -> bool {
        !self.api_key.is_empty()
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
    pub fn api_key(mut self, key: impl Into<String>) -> Self {
        self.config.api_key = key.into();
        self
    }

    pub fn model(mut self, model: impl Into<String>) -> Self {
        self.config.model = model.into();
        self
    }

    pub fn max_tokens(mut self, tokens: u32) -> Self {
        self.config.max_tokens = tokens;
        self
    }

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

    pub fn build(self) -> AIConfig {
        self.config
    }
}

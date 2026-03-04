//! Provider configuration and endpoint resolution.
//!
//! A provider represents an LLM backend (OpenAI, Anthropic, etc.) with its
//! API key, model, and endpoint. Haira supports many backends through a
//! unified OpenAI-compatible chat completions interface.

/// Configuration for an LLM provider.
#[derive(Debug, Clone)]
pub struct Provider {
    pub name: String,
    pub api_key: String,
    pub endpoint: String,
    pub model: String,
    pub api_version: Option<String>,
    pub temperature: Option<f64>,
    pub max_tokens: Option<u32>,
}

impl Provider {
    /// Create a provider from a map of field key-value pairs.
    ///
    /// Recognized keys: api_key, endpoint, model, api_version, temperature, max_tokens.
    /// The endpoint is resolved automatically based on known backend names if not
    /// explicitly set.
    pub fn from_fields(name: &str, fields: &[(String, String)]) -> Self {
        let mut api_key = String::new();
        let mut endpoint = String::new();
        let mut model = String::new();
        let mut api_version = None;
        let mut temperature = None;
        let mut max_tokens = None;

        for (key, val) in fields {
            match key.as_str() {
                "api_key" => api_key = val.clone(),
                "endpoint" => endpoint = val.clone(),
                "model" => model = val.clone(),
                "api_version" => api_version = Some(val.clone()),
                "temperature" => temperature = val.parse().ok(),
                "max_tokens" => max_tokens = val.parse().ok(),
                _ => {}
            }
        }

        // Auto-resolve endpoint from provider name if not explicitly set.
        if endpoint.is_empty() {
            endpoint = resolve_endpoint(name);
        }

        Provider {
            name: name.to_string(),
            api_key,
            endpoint,
            model,
            api_version,
            temperature,
            max_tokens,
        }
    }
}

/// Resolve the chat completions endpoint from a known backend name.
///
/// Returns the base URL for the chat completions API. For unknown backends,
/// returns OpenAI's endpoint as the default (most backends are OpenAI-compatible).
pub fn resolve_endpoint(backend: &str) -> String {
    match backend.to_lowercase().as_str() {
        "openai" => "https://api.openai.com/v1".to_string(),
        "anthropic" => "https://api.anthropic.com/v1".to_string(),
        "groq" => "https://api.groq.com/openai/v1".to_string(),
        "ollama" => "http://localhost:11434/v1".to_string(),
        "together" => "https://api.together.xyz/v1".to_string(),
        "mistral" => "https://api.mistral.ai/v1".to_string(),
        "deepseek" => "https://api.deepseek.com/v1".to_string(),
        "fireworks" => "https://api.fireworks.ai/inference/v1".to_string(),
        "openrouter" => "https://openrouter.ai/api/v1".to_string(),
        "xai" => "https://api.x.ai/v1".to_string(),
        "cerebras" => "https://api.cerebras.ai/v1".to_string(),
        // Azure uses a different URL pattern — users must provide endpoint explicitly.
        "azure_openai" | "azure" => String::new(),
        _ => "https://api.openai.com/v1".to_string(),
    }
}

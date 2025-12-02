//! Ollama API client for local LLM inference.

use serde::{Deserialize, Serialize};
use thiserror::Error;

/// Default Ollama server URL.
pub const DEFAULT_OLLAMA_URL: &str = "http://localhost:11434";

/// Default model for code generation.
pub const DEFAULT_OLLAMA_MODEL: &str = "deepseek-coder-v2";

/// Ollama API client.
pub struct OllamaClient {
    client: reqwest::Client,
    base_url: String,
    model: String,
}

/// Errors from the Ollama client.
#[derive(Debug, Error)]
pub enum OllamaError {
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),
    #[error("Ollama API error: {0}")]
    Api(String),
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),
    #[error("Ollama server not running at {0}. Start it with: ollama serve")]
    ServerNotRunning(String),
    #[error("Model '{0}' not found. Pull it with: ollama pull {0}")]
    ModelNotFound(String),
}

/// Request to Ollama generate API.
#[derive(Debug, Serialize)]
struct OllamaRequest {
    model: String,
    prompt: String,
    system: Option<String>,
    stream: bool,
    options: Option<OllamaOptions>,
}

/// Ollama generation options.
#[derive(Debug, Serialize)]
struct OllamaOptions {
    temperature: f32,
    num_predict: i32,
}

/// Response from Ollama generate API.
#[derive(Debug, Deserialize)]
struct OllamaResponse {
    response: String,
    #[allow(dead_code)]
    done: bool,
    #[serde(default)]
    error: Option<String>,
}

/// Response from Ollama tags API (list models).
#[derive(Debug, Deserialize)]
struct OllamaTagsResponse {
    models: Vec<OllamaModel>,
}

#[derive(Debug, Deserialize)]
struct OllamaModel {
    name: String,
}

impl OllamaClient {
    /// Create a new Ollama client with default settings.
    pub fn new() -> Self {
        Self {
            client: reqwest::Client::new(),
            base_url: DEFAULT_OLLAMA_URL.to_string(),
            model: DEFAULT_OLLAMA_MODEL.to_string(),
        }
    }

    /// Create a new Ollama client with custom URL and model.
    pub fn with_config(base_url: impl Into<String>, model: impl Into<String>) -> Self {
        Self {
            client: reqwest::Client::new(),
            base_url: base_url.into(),
            model: model.into(),
        }
    }

    /// Set the model to use.
    pub fn with_model(mut self, model: impl Into<String>) -> Self {
        self.model = model.into();
        self
    }

    /// Set the base URL.
    pub fn with_url(mut self, url: impl Into<String>) -> Self {
        self.base_url = url.into();
        self
    }

    /// Check if Ollama server is running and model is available.
    pub async fn check_availability(&self) -> Result<(), OllamaError> {
        // Check if server is running
        let tags_url = format!("{}/api/tags", self.base_url);
        let response = self
            .client
            .get(&tags_url)
            .send()
            .await
            .map_err(|_| OllamaError::ServerNotRunning(self.base_url.clone()))?;

        if !response.status().is_success() {
            return Err(OllamaError::ServerNotRunning(self.base_url.clone()));
        }

        // Check if model is available
        let tags: OllamaTagsResponse = response.json().await?;
        let model_base = self.model.split(':').next().unwrap_or(&self.model);

        let model_found = tags
            .models
            .iter()
            .any(|m| m.name == self.model || m.name.starts_with(&format!("{}:", model_base)));

        if !model_found {
            return Err(OllamaError::ModelNotFound(self.model.clone()));
        }

        Ok(())
    }

    /// Send a prompt to Ollama and get a response.
    pub async fn complete(&self, system: &str, user_message: &str) -> Result<String, OllamaError> {
        let request = OllamaRequest {
            model: self.model.clone(),
            prompt: user_message.to_string(),
            system: Some(system.to_string()),
            stream: false,
            options: Some(OllamaOptions {
                temperature: 0.1, // Low temperature for deterministic code generation
                num_predict: 4096,
            }),
        };

        let url = format!("{}/api/generate", self.base_url);

        let response = self
            .client
            .post(&url)
            .json(&request)
            .send()
            .await
            .map_err(|e| {
                if e.is_connect() {
                    OllamaError::ServerNotRunning(self.base_url.clone())
                } else {
                    OllamaError::Http(e)
                }
            })?;

        if !response.status().is_success() {
            let status = response.status();
            let text = response.text().await.unwrap_or_default();
            return Err(OllamaError::Api(format!("{}: {}", status, text)));
        }

        let response: OllamaResponse = response.json().await?;

        if let Some(error) = response.error {
            return Err(OllamaError::Api(error));
        }

        Ok(response.response)
    }

    /// Get the current model name.
    pub fn model(&self) -> &str {
        &self.model
    }

    /// Get the base URL.
    pub fn base_url(&self) -> &str {
        &self.base_url
    }
}

impl Default for OllamaClient {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_client() {
        let client = OllamaClient::new();
        assert_eq!(client.base_url(), DEFAULT_OLLAMA_URL);
        assert_eq!(client.model(), DEFAULT_OLLAMA_MODEL);
    }

    #[test]
    fn test_custom_config() {
        let client = OllamaClient::with_config("http://localhost:8080", "codellama:7b");
        assert_eq!(client.base_url(), "http://localhost:8080");
        assert_eq!(client.model(), "codellama:7b");
    }

    #[test]
    fn test_builder_pattern() {
        let client = OllamaClient::new()
            .with_url("http://myserver:11434")
            .with_model("qwen2.5-coder:7b");
        assert_eq!(client.base_url(), "http://myserver:11434");
        assert_eq!(client.model(), "qwen2.5-coder:7b");
    }
}

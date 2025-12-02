//! HTTP client for llama-server's OpenAI-compatible API.

use serde::{Deserialize, Serialize};

use crate::error::LocalAIError;
use crate::DEFAULT_PORT;

/// Client for communicating with llama-server.
pub struct LlamaCppClient {
    client: reqwest::Client,
    base_url: String,
}

/// OpenAI-compatible chat completion request.
#[derive(Debug, Serialize)]
struct ChatCompletionRequest {
    messages: Vec<ChatMessage>,
    temperature: f32,
    max_tokens: i32,
    stream: bool,
}

/// Chat message in OpenAI format.
#[derive(Debug, Serialize)]
struct ChatMessage {
    role: String,
    content: String,
}

/// OpenAI-compatible chat completion response.
#[derive(Debug, Deserialize)]
struct ChatCompletionResponse {
    choices: Vec<Choice>,
    #[allow(dead_code)]
    usage: Option<Usage>,
}

#[derive(Debug, Deserialize)]
struct Choice {
    message: ResponseMessage,
    #[allow(dead_code)]
    finish_reason: Option<String>,
}

#[derive(Debug, Deserialize)]
struct ResponseMessage {
    content: String,
}

#[derive(Debug, Deserialize)]
struct Usage {
    #[allow(dead_code)]
    prompt_tokens: u32,
    #[allow(dead_code)]
    completion_tokens: u32,
    #[allow(dead_code)]
    total_tokens: u32,
}

impl LlamaCppClient {
    /// Create a new client with default URL (localhost:11435).
    pub fn new() -> Self {
        Self::with_url(format!("http://127.0.0.1:{}", DEFAULT_PORT))
    }

    /// Create a new client with a custom URL.
    pub fn with_url(base_url: impl Into<String>) -> Self {
        Self {
            client: reqwest::Client::new(),
            base_url: base_url.into(),
        }
    }

    /// Create a new client with a custom port on localhost.
    pub fn with_port(port: u16) -> Self {
        Self::with_url(format!("http://127.0.0.1:{}", port))
    }

    /// Get the base URL.
    pub fn base_url(&self) -> &str {
        &self.base_url
    }

    /// Check if the server is running and healthy.
    pub async fn check_health(&self) -> Result<(), LocalAIError> {
        let url = format!("{}/health", self.base_url);

        let response = self
            .client
            .get(&url)
            .timeout(std::time::Duration::from_secs(5))
            .send()
            .await
            .map_err(|e| {
                if e.is_connect() || e.is_timeout() {
                    LocalAIError::ServerNotRunning(self.base_url.clone())
                } else {
                    LocalAIError::Http(e)
                }
            })?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(LocalAIError::ServerNotRunning(self.base_url.clone()))
        }
    }

    /// Send a completion request to the server.
    ///
    /// Uses the OpenAI-compatible `/v1/chat/completions` endpoint.
    pub async fn complete(&self, system: &str, user_message: &str) -> Result<String, LocalAIError> {
        let request = ChatCompletionRequest {
            messages: vec![
                ChatMessage {
                    role: "system".to_string(),
                    content: system.to_string(),
                },
                ChatMessage {
                    role: "user".to_string(),
                    content: user_message.to_string(),
                },
            ],
            temperature: 0.1, // Low temperature for deterministic code generation
            max_tokens: 2048,
            stream: false,
        };

        let url = format!("{}/v1/chat/completions", self.base_url);

        let response = self
            .client
            .post(&url)
            .json(&request)
            .send()
            .await
            .map_err(|e| {
                if e.is_connect() {
                    LocalAIError::ServerNotRunning(self.base_url.clone())
                } else {
                    LocalAIError::Http(e)
                }
            })?;

        if !response.status().is_success() {
            let status = response.status();
            let text = response.text().await.unwrap_or_default();
            return Err(LocalAIError::Api(format!("{}: {}", status, text)));
        }

        let completion: ChatCompletionResponse = response.json().await?;

        completion
            .choices
            .into_iter()
            .next()
            .map(|c| c.message.content)
            .ok_or_else(|| LocalAIError::Api("No completion returned".to_string()))
    }
}

impl Default for LlamaCppClient {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_url() {
        let client = LlamaCppClient::new();
        assert_eq!(client.base_url(), "http://127.0.0.1:11435");
    }

    #[test]
    fn test_custom_url() {
        let client = LlamaCppClient::with_url("http://192.168.1.100:8080");
        assert_eq!(client.base_url(), "http://192.168.1.100:8080");
    }

    #[test]
    fn test_custom_port() {
        let client = LlamaCppClient::with_port(9000);
        assert_eq!(client.base_url(), "http://127.0.0.1:9000");
    }
}

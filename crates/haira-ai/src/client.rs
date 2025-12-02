//! Claude API client.

use reqwest::header::{HeaderMap, HeaderValue, CONTENT_TYPE};
use serde::{Deserialize, Serialize};
use thiserror::Error;

use crate::AIConfig;

/// Claude API client.
pub struct ClaudeClient {
    client: reqwest::Client,
    config: AIConfig,
}

/// Errors from the Claude client.
#[derive(Debug, Error)]
pub enum ClientError {
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),
    #[error("API error: {status} - {message}")]
    Api { status: u16, message: String },
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),
    #[error("Missing API key")]
    MissingApiKey,
}

/// Request to Claude API.
#[derive(Debug, Serialize)]
struct ClaudeRequest {
    model: String,
    max_tokens: u32,
    messages: Vec<Message>,
    system: Option<String>,
}

/// A message in the conversation.
#[derive(Debug, Serialize, Deserialize)]
struct Message {
    role: String,
    content: String,
}

/// Response from Claude API.
#[derive(Debug, Deserialize)]
struct ClaudeResponse {
    content: Vec<ContentBlock>,
    #[allow(dead_code)]
    stop_reason: Option<String>,
    #[allow(dead_code)]
    usage: Option<Usage>,
}

#[derive(Debug, Deserialize)]
struct ContentBlock {
    #[serde(rename = "type")]
    #[allow(dead_code)]
    block_type: String,
    text: Option<String>,
}

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct Usage {
    input_tokens: u32,
    output_tokens: u32,
}

/// Error response from Claude API.
#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: ApiError,
}

#[derive(Debug, Deserialize)]
struct ApiError {
    message: String,
    #[serde(rename = "type")]
    #[allow(dead_code)]
    error_type: String,
}

impl ClaudeClient {
    /// Create a new Claude client.
    pub fn new(config: AIConfig) -> Result<Self, ClientError> {
        if config.api_key.is_empty() {
            return Err(ClientError::MissingApiKey);
        }

        let mut headers = HeaderMap::new();
        headers.insert(CONTENT_TYPE, HeaderValue::from_static("application/json"));
        headers.insert("x-api-key", HeaderValue::from_str(&config.api_key).unwrap());
        headers.insert("anthropic-version", HeaderValue::from_static("2023-06-01"));

        let client = reqwest::Client::builder()
            .default_headers(headers)
            .build()?;

        Ok(Self { client, config })
    }

    /// Send a message to Claude and get a response.
    pub async fn complete(&self, system: &str, user_message: &str) -> Result<String, ClientError> {
        let request = ClaudeRequest {
            model: self.config.model.clone(),
            max_tokens: self.config.max_tokens,
            messages: vec![Message {
                role: "user".to_string(),
                content: user_message.to_string(),
            }],
            system: Some(system.to_string()),
        };

        let response = self
            .client
            .post("https://api.anthropic.com/v1/messages")
            .json(&request)
            .send()
            .await?;

        let status = response.status();

        if !status.is_success() {
            let error: ErrorResponse = response.json().await?;
            return Err(ClientError::Api {
                status: status.as_u16(),
                message: error.error.message,
            });
        }

        let response: ClaudeResponse = response.json().await?;

        // Extract text from response
        let text = response
            .content
            .iter()
            .filter_map(|block| block.text.as_ref())
            .map(|s| s.as_str())
            .collect::<Vec<_>>()
            .join("");

        Ok(text)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_missing_api_key() {
        let config = AIConfig::default();
        let result = ClaudeClient::new(config);
        assert!(matches!(result, Err(ClientError::MissingApiKey)));
    }
}

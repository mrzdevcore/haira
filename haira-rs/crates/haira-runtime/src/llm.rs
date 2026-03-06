//! OpenAI-compatible chat completions client.
//!
//! Sends synchronous HTTP requests via `ureq`. All LLM providers that support
//! the OpenAI chat completions format can be used through this module.

use serde::{Deserialize, Serialize};

// ---------------------------------------------------------------------------
// Request types
// ---------------------------------------------------------------------------

#[derive(Debug, Serialize)]
pub struct ChatRequest {
    pub model: String,
    pub messages: Vec<ChatMessage>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub temperature: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub max_tokens: Option<u32>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub tools: Vec<serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatMessage {
    pub role: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub content: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_calls: Option<Vec<ToolCall>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tool_call_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ToolCall {
    pub id: String,
    #[serde(rename = "type")]
    pub call_type: String,
    pub function: FunctionCall,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FunctionCall {
    pub name: String,
    pub arguments: String,
}

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

#[derive(Debug, Deserialize)]
pub struct ChatResponse {
    pub choices: Vec<Choice>,
}

#[derive(Debug, Deserialize)]
pub struct Choice {
    pub message: ChatMessage,
    pub finish_reason: Option<String>,
}

// ---------------------------------------------------------------------------
// Client
// ---------------------------------------------------------------------------

/// Send a chat completion request to an OpenAI-compatible endpoint.
///
/// The endpoint should be the base URL (e.g. `https://api.openai.com/v1`).
/// This function appends `/chat/completions` automatically.
pub fn chat_completion(
    endpoint: &str,
    api_key: &str,
    api_version: Option<&str>,
    request: &ChatRequest,
) -> Result<ChatResponse, String> {
    let url = if endpoint.contains("/chat/completions") {
        endpoint.to_string()
    } else {
        let base = endpoint.trim_end_matches('/');
        format!("{base}/chat/completions")
    };

    let mut req = ureq::post(&url)
        .set("Content-Type", "application/json");

    if !api_key.is_empty() {
        req = req.set("Authorization", &format!("Bearer {api_key}"));
    }

    // Azure uses api-key header and query param for version.
    if let Some(ver) = api_version {
        let url_with_version = if url.contains('?') {
            format!("{url}&api-version={ver}")
        } else {
            format!("{url}?api-version={ver}")
        };
        req = ureq::post(&url_with_version)
            .set("Content-Type", "application/json")
            .set("api-key", api_key);
    }

    let body = serde_json::to_string(request)
        .map_err(|e| format!("failed to serialize request: {e}"))?;

    let response = req
        .send_string(&body)
        .map_err(|e| format!("HTTP request failed: {e}"))?;

    let response_body = response
        .into_string()
        .map_err(|e| format!("failed to read response body: {e}"))?;

    serde_json::from_str::<ChatResponse>(&response_body)
        .map_err(|e| format!("failed to parse response: {e}\nBody: {response_body}"))
}

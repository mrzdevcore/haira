//! AI Engine - the main entry point for intent interpretation.

use thiserror::Error;
use tracing::{debug, info, warn};

use crate::cache::AICache;
use crate::client::{ClaudeClient, ClientError};
use crate::config::AIConfig;
use crate::prompt::{self, SYSTEM_PROMPT};
use haira_cir::{AIRequest, AIResponse, CIRFunction, InterpretationContext};

/// AI Engine for interpreting developer intent.
pub struct AIEngine {
    config: AIConfig,
    client: Option<ClaudeClient>,
    cache: AICache,
}

/// Errors from the AI engine.
#[derive(Debug, Error)]
pub enum AIError {
    #[error("client error: {0}")]
    Client(#[from] ClientError),
    #[error("cache error: {0}")]
    Cache(#[from] crate::cache::CacheError),
    #[error("parse error: {0}")]
    Parse(#[from] serde_json::Error),
    #[error("validation error: {0}")]
    Validation(String),
    #[error("low confidence: {confidence} (minimum: {minimum})")]
    LowConfidence { confidence: f64, minimum: f64 },
    #[error("AI interpretation failed: {0}")]
    InterpretationFailed(String),
    #[error("missing API key - set ANTHROPIC_API_KEY environment variable")]
    MissingApiKey,
}

impl AIEngine {
    /// Create a new AI engine.
    pub fn new(config: AIConfig) -> Self {
        let client = if config.api_key.is_empty() {
            None
        } else {
            ClaudeClient::new(config.clone()).ok()
        };

        let cache = AICache::new(config.cache_dir.clone());

        Self {
            config,
            client,
            cache,
        }
    }

    /// Interpret a function call and generate CIR.
    pub async fn interpret(
        &mut self,
        function_name: &str,
        context: InterpretationContext,
    ) -> Result<CIRFunction, AIError> {
        info!("Interpreting function: {}", function_name);

        // 1. Try to match a simple pattern (no AI needed)
        if let Some((pattern, type_name, field)) = prompt::parse_function_name(function_name) {
            debug!("Matched pattern: {} for type {}", pattern, type_name);

            // Check if the type exists in context
            let type_exists = context
                .types_in_scope
                .iter()
                .any(|t| t.name == type_name);

            if type_exists {
                if let Some(func) = prompt::build_simple_pattern_prompt(
                    &pattern,
                    &type_name,
                    field.as_deref(),
                ) {
                    info!("Generated from pattern (no AI): {}", function_name);
                    return Ok(func);
                }
            }
        }

        // 2. Check cache
        let context_json = serde_json::to_string(&context)?;
        let cache_key = AICache::cache_key(function_name, &context_json);

        if self.config.use_cache {
            if let Some(func) = self.cache.get(&cache_key) {
                info!("Cache hit for: {}", function_name);
                return Ok(func);
            }
        }

        // 3. Call Claude API
        let client = self.client.as_ref().ok_or(AIError::MissingApiKey)?;

        let user_prompt = prompt::build_user_prompt(function_name, &context);

        debug!("Calling Claude API...");
        let response_text = client.complete(SYSTEM_PROMPT, &user_prompt).await?;

        // 4. Parse response
        let response: AIResponse = self.parse_response(&response_text)?;

        if !response.success {
            return Err(AIError::InterpretationFailed(
                response.error.unwrap_or_else(|| "Unknown error".to_string()),
            ));
        }

        // 5. Check confidence
        if response.confidence < self.config.min_confidence {
            warn!(
                "Low confidence for {}: {} (minimum: {})",
                function_name, response.confidence, self.config.min_confidence
            );
            return Err(AIError::LowConfidence {
                confidence: response.confidence,
                minimum: self.config.min_confidence,
            });
        }

        let func = response
            .interpretation
            .ok_or_else(|| AIError::InterpretationFailed("No interpretation returned".to_string()))?;

        // 6. Validate CIR
        if let Err(errors) = haira_cir::validate(&func) {
            let error_msg = errors
                .iter()
                .map(|e| e.to_string())
                .collect::<Vec<_>>()
                .join(", ");
            return Err(AIError::Validation(error_msg));
        }

        // 7. Cache result
        if self.config.use_cache {
            self.cache.set(&cache_key, &func)?;
            info!("Cached result for: {}", function_name);
        }

        info!(
            "Successfully interpreted {} with confidence {}",
            function_name, response.confidence
        );

        Ok(func)
    }

    /// Parse AI response, handling potential JSON issues.
    fn parse_response(&self, text: &str) -> Result<AIResponse, serde_json::Error> {
        // Try to extract JSON from the response
        let json_text = if let Some(start) = text.find('{') {
            if let Some(end) = text.rfind('}') {
                &text[start..=end]
            } else {
                text
            }
        } else {
            text
        };

        serde_json::from_str(json_text)
    }

    /// Check if a function name matches a known pattern.
    pub fn matches_pattern(&self, function_name: &str) -> bool {
        prompt::parse_function_name(function_name).is_some()
    }

    /// Get confidence level description.
    pub fn confidence_level(confidence: f64) -> &'static str {
        if confidence >= 0.9 {
            "high"
        } else if confidence >= 0.7 {
            "medium"
        } else if confidence >= 0.5 {
            "low"
        } else {
            "failed"
        }
    }

    /// Clear the cache.
    pub fn clear_cache(&mut self) -> Result<(), AIError> {
        self.cache.clear()?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use haira_cir::{CallSiteInfo, TypeDefinition, FieldDefinition};

    fn test_context() -> InterpretationContext {
        InterpretationContext {
            types_in_scope: vec![TypeDefinition {
                name: "User".to_string(),
                fields: vec![
                    FieldDefinition {
                        name: "id".to_string(),
                        ty: "int".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "name".to_string(),
                        ty: "string".to_string(),
                        optional: false,
                        default: None,
                    },
                    FieldDefinition {
                        name: "active".to_string(),
                        ty: "bool".to_string(),
                        optional: false,
                        default: None,
                    },
                ],
            }],
            call_site: CallSiteInfo {
                file: "main.haira".to_string(),
                line: 10,
                arguments: vec![],
                expected_return: None,
            },
            project_schema: Default::default(),
        }
    }

    #[test]
    fn test_pattern_matching() {
        let config = AIConfig::default();
        let engine = AIEngine::new(config);

        assert!(engine.matches_pattern("get_users"));
        assert!(engine.matches_pattern("get_user_by_id"));
        assert!(engine.matches_pattern("get_active_users"));
        assert!(engine.matches_pattern("save_user"));
        assert!(engine.matches_pattern("delete_user"));
        assert!(!engine.matches_pattern("do_something_complex"));
    }

    #[tokio::test]
    async fn test_simple_pattern_no_ai() {
        let config = AIConfig::builder()
            .use_cache(false)
            .build();
        let mut engine = AIEngine::new(config);
        let context = test_context();

        // This should work without AI because it matches a pattern
        let result = engine.interpret("get_users", context).await;

        assert!(result.is_ok());
        let func = result.unwrap();
        assert_eq!(func.name, "get_users");
    }

    #[test]
    fn test_confidence_levels() {
        assert_eq!(AIEngine::confidence_level(0.95), "high");
        assert_eq!(AIEngine::confidence_level(0.85), "medium");
        assert_eq!(AIEngine::confidence_level(0.6), "low");
        assert_eq!(AIEngine::confidence_level(0.3), "failed");
    }
}

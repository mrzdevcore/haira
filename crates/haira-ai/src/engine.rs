//! AI Engine - the main entry point for intent interpretation.

use thiserror::Error;
use tracing::{debug, info, warn};

use crate::cache::AICache;
use crate::client::{ClaudeClient, ClientError};
use crate::config::AIConfig;
use crate::ollama::{OllamaClient, OllamaError};
use crate::prompt::{self, SYSTEM_PROMPT};
use haira_cir::{AIResponse, CIRFunction, InterpretationContext};
use haira_local_ai::{LlamaCppServer, LocalAIError};

/// AI backend type.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum AIBackend {
    /// Use Claude API (requires ANTHROPIC_API_KEY)
    Claude,
    /// Use local Ollama server
    Ollama,
    /// Use local llama.cpp server (self-managed)
    LocalAI,
}

/// AI Engine for interpreting developer intent.
pub struct AIEngine {
    config: AIConfig,
    claude_client: Option<ClaudeClient>,
    ollama_client: Option<OllamaClient>,
    local_ai_server: Option<LlamaCppServer>,
    backend: AIBackend,
    cache: AICache,
}

/// Errors from the AI engine.
#[derive(Debug, Error)]
pub enum AIError {
    #[error("client error: {0}")]
    Client(#[from] ClientError),
    #[error("ollama error: {0}")]
    Ollama(#[from] OllamaError),
    #[error("local AI error: {0}")]
    LocalAI(#[from] LocalAIError),
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
    #[error("no AI backend available")]
    NoBackend,
}

impl AIEngine {
    /// Create a new AI engine with Claude backend.
    pub fn new(config: AIConfig) -> Self {
        let claude_client = if config.api_key.is_empty() {
            None
        } else {
            ClaudeClient::new(config.clone()).ok()
        };

        let cache = AICache::new(config.cache_dir.clone());

        Self {
            config,
            claude_client,
            ollama_client: None,
            local_ai_server: None,
            backend: AIBackend::Claude,
            cache,
        }
    }

    /// Create a new AI engine with Ollama backend.
    pub fn with_ollama(config: AIConfig, ollama_model: Option<&str>) -> Self {
        let ollama_client = if let Some(model) = ollama_model {
            OllamaClient::new().with_model(model)
        } else {
            OllamaClient::new()
        };

        let cache = AICache::new(config.cache_dir.clone());

        Self {
            config,
            claude_client: None,
            ollama_client: Some(ollama_client),
            local_ai_server: None,
            backend: AIBackend::Ollama,
            cache,
        }
    }

    /// Create a new AI engine with local llama.cpp backend.
    ///
    /// The model filename should be the name of a GGUF file in ~/.haira/models/
    pub fn with_local_ai(config: AIConfig, model_filename: Option<&str>) -> Self {
        let filename = model_filename
            .unwrap_or(haira_local_ai::DEFAULT_MODEL_FILENAME)
            .to_string();

        let server = LlamaCppServer::new(filename);
        let cache = AICache::new(config.cache_dir.clone());

        Self {
            config,
            claude_client: None,
            ollama_client: None,
            local_ai_server: Some(server),
            backend: AIBackend::LocalAI,
            cache,
        }
    }

    /// Set the AI backend to use.
    pub fn set_backend(&mut self, backend: AIBackend) {
        self.backend = backend;
    }

    /// Get the current backend.
    pub fn backend(&self) -> AIBackend {
        self.backend
    }

    /// Start the local AI server (only for LocalAI backend).
    ///
    /// This starts the llama-server process and waits for it to become ready.
    pub async fn start_local_server(&mut self) -> Result<(), AIError> {
        if self.backend != AIBackend::LocalAI {
            return Ok(()); // No-op for other backends
        }

        let server = self.local_ai_server.as_mut().ok_or(AIError::NoBackend)?;

        // Start the server process
        server.start()?;

        // Wait for it to become ready (up to 60 seconds for model loading)
        server
            .wait_ready(std::time::Duration::from_secs(60))
            .await?;

        Ok(())
    }

    /// Stop the local AI server (only for LocalAI backend).
    pub fn stop_local_server(&mut self) -> Result<(), AIError> {
        if let Some(ref mut server) = self.local_ai_server {
            server.stop()?;
        }
        Ok(())
    }

    /// Check if the local AI server is running.
    pub fn is_local_server_running(&mut self) -> bool {
        if let Some(ref mut server) = self.local_ai_server {
            server.is_running()
        } else {
            false
        }
    }

    /// Check if the current backend is available.
    pub async fn check_availability(&self) -> Result<(), AIError> {
        match self.backend {
            AIBackend::Claude => {
                if self.claude_client.is_none() {
                    return Err(AIError::MissingApiKey);
                }
                Ok(())
            }
            AIBackend::Ollama => {
                let client = self.ollama_client.as_ref().ok_or(AIError::NoBackend)?;
                client.check_availability().await?;
                Ok(())
            }
            AIBackend::LocalAI => {
                let server = self.local_ai_server.as_ref().ok_or(AIError::NoBackend)?;
                // Check that the server binary and model exist
                if !server.binary_exists() {
                    return Err(AIError::LocalAI(LocalAIError::ServerBinaryNotFound(
                        haira_local_ai::paths::llama_server_path()
                            .display()
                            .to_string(),
                    )));
                }
                if !server.model_exists() {
                    return Err(AIError::LocalAI(LocalAIError::ModelNotFound(
                        "Model not found. Run: haira model pull".to_string(),
                    )));
                }
                Ok(())
            }
        }
    }

    /// Complete a prompt using the configured backend.
    async fn complete(&self, system: &str, user_message: &str) -> Result<String, AIError> {
        match self.backend {
            AIBackend::Claude => {
                let client = self.claude_client.as_ref().ok_or(AIError::MissingApiKey)?;
                Ok(client.complete(system, user_message).await?)
            }
            AIBackend::Ollama => {
                let client = self.ollama_client.as_ref().ok_or(AIError::NoBackend)?;
                Ok(client.complete(system, user_message).await?)
            }
            AIBackend::LocalAI => {
                let server = self.local_ai_server.as_ref().ok_or(AIError::NoBackend)?;
                let client = server.client();
                Ok(client.complete(system, user_message).await?)
            }
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
            let type_exists = context.types_in_scope.iter().any(|t| t.name == type_name);

            if type_exists {
                if let Some(func) =
                    prompt::build_simple_pattern_prompt(&pattern, &type_name, field.as_deref())
                {
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

        // 3. Call AI backend
        let user_prompt = prompt::build_user_prompt(function_name, &context);

        debug!("Calling {:?} backend...", self.backend);
        let response_text = self.complete(SYSTEM_PROMPT, &user_prompt).await?;

        // 4. Parse response
        let response: AIResponse = self.parse_response(&response_text)?;

        if !response.success {
            return Err(AIError::InterpretationFailed(
                response
                    .error
                    .unwrap_or_else(|| "Unknown error".to_string()),
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

        let func = response.interpretation.ok_or_else(|| {
            AIError::InterpretationFailed("No interpretation returned".to_string())
        })?;

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

    /// Interpret an explicit AI intent block.
    ///
    /// This is called when the user explicitly defines what they want using
    /// the `ai` block syntax:
    ///
    /// ```haira
    /// ai summarize_activity(user: User) -> ActivitySummary {
    ///     Summarize the user's activity over the last 30 days.
    ///     Group by activity type and find most common.
    /// }
    /// ```
    pub async fn interpret_intent(
        &mut self,
        function_name: Option<&str>,
        intent: &str,
        params: &[(String, String)], // (name, type) pairs
        return_type: Option<&str>,
        context: InterpretationContext,
    ) -> Result<CIRFunction, AIError> {
        let name_for_log = function_name.unwrap_or("<anonymous>");
        info!("Interpreting explicit intent for: {}", name_for_log);
        debug!("Intent: {}", intent);

        // 1. Build cache key from intent + signature + context
        let cache_key = self.intent_cache_key(function_name, intent, params, return_type, &context);

        // 2. Check cache
        if self.config.use_cache {
            if let Some(func) = self.cache.get(&cache_key) {
                info!("Cache hit for intent: {}", name_for_log);
                return Ok(func);
            }
        }

        // 3. Call AI backend
        let user_prompt =
            prompt::build_intent_prompt(function_name, intent, params, return_type, &context);

        debug!("Calling {:?} backend for intent...", self.backend);
        let response_text = self.complete(SYSTEM_PROMPT, &user_prompt).await?;

        // 4. Parse response
        let response: AIResponse = self.parse_response(&response_text)?;

        if !response.success {
            return Err(AIError::InterpretationFailed(
                response
                    .error
                    .unwrap_or_else(|| "Unknown error".to_string()),
            ));
        }

        // 5. Check confidence (can be more lenient for explicit intents)
        let min_confidence = self.config.min_confidence * 0.8; // 20% more lenient for explicit
        if response.confidence < min_confidence {
            warn!(
                "Low confidence for intent {}: {} (minimum: {})",
                name_for_log, response.confidence, min_confidence
            );
            return Err(AIError::LowConfidence {
                confidence: response.confidence,
                minimum: min_confidence,
            });
        }

        let mut func = response.interpretation.ok_or_else(|| {
            AIError::InterpretationFailed("No interpretation returned".to_string())
        })?;

        // 6. Override function name if provided
        if let Some(name) = function_name {
            func.name = name.to_string();
        }

        // 6.5. Fix recursive calls - LLMs often confuse variable names with function names
        Self::fix_recursive_calls(&mut func);

        // 7. Validate CIR
        if let Err(errors) = haira_cir::validate(&func) {
            let error_msg = errors
                .iter()
                .map(|e| e.to_string())
                .collect::<Vec<_>>()
                .join(", ");
            return Err(AIError::Validation(error_msg));
        }

        // 8. Cache result
        if self.config.use_cache {
            self.cache.set(&cache_key, &func)?;
            info!("Cached intent result for: {}", name_for_log);
        }

        info!(
            "Successfully interpreted intent {} with confidence {}",
            name_for_log, response.confidence
        );

        Ok(func)
    }

    /// Build a cache key for an explicit intent block.
    fn intent_cache_key(
        &self,
        function_name: Option<&str>,
        intent: &str,
        params: &[(String, String)],
        return_type: Option<&str>,
        context: &InterpretationContext,
    ) -> String {
        use std::collections::hash_map::DefaultHasher;
        use std::hash::{Hash, Hasher};

        let mut hasher = DefaultHasher::new();

        // Hash the intent text
        intent.hash(&mut hasher);

        // Hash the function signature
        function_name.hash(&mut hasher);
        for (name, ty) in params {
            name.hash(&mut hasher);
            ty.hash(&mut hasher);
        }
        return_type.hash(&mut hasher);

        // Hash relevant context (types in scope)
        for ty in &context.types_in_scope {
            ty.name.hash(&mut hasher);
            for field in &ty.fields {
                field.name.hash(&mut hasher);
                field.ty.hash(&mut hasher);
            }
        }

        format!("intent_{:x}", hasher.finish())
    }

    /// Parse AI response, handling potential JSON issues.
    fn parse_response(&self, text: &str) -> Result<AIResponse, serde_json::Error> {
        debug!("Raw AI response ({} chars):\n{}", text.len(), text);

        // Clean up common LLM artifacts
        let cleaned = Self::clean_llm_output(text);

        // Try to extract the first complete JSON object from the response
        // This handles cases where the model repeats the JSON multiple times
        let json_text = if let Some(start) = cleaned.find('{') {
            // Find the matching closing brace by counting braces
            let chars: Vec<char> = cleaned[start..].chars().collect();
            let mut depth = 0;
            let mut end_offset = 0;
            for (i, ch) in chars.iter().enumerate() {
                match ch {
                    '{' => depth += 1,
                    '}' => {
                        depth -= 1;
                        if depth == 0 {
                            end_offset = i;
                            break;
                        }
                    }
                    _ => {}
                }
            }
            if end_offset > 0 {
                &cleaned[start..=start + end_offset]
            } else if let Some(end) = cleaned.rfind('}') {
                &cleaned[start..=end]
            } else {
                &cleaned
            }
        } else {
            &cleaned
        };

        debug!("Extracted JSON:\n{}", json_text);

        // Normalize the JSON to handle common model variations
        let normalized = self.normalize_cir_json(json_text);
        debug!("Normalized JSON:\n{}", normalized);

        let result = serde_json::from_str(&normalized);
        if let Err(ref e) = result {
            warn!("JSON parse error: {}", e);
            warn!("Failed to parse JSON:\n{}", normalized);
        }

        result
    }

    /// Clean up common LLM output artifacts.
    fn clean_llm_output(text: &str) -> String {
        let mut cleaned = text.to_string();

        // Remove common special tokens from various models
        let tokens_to_remove = [
            "<｜begin▁of▁sentence｜>",
            "<｜end▁of▁sentence｜>",
            "<|endoftext|>",
            "<|im_end|>",
            "<|im_start|>",
            "</s>",
            "<s>",
        ];

        for token in tokens_to_remove {
            cleaned = cleaned.replace(token, "");
        }

        // Fix common JSON errors
        loop {
            let new_cleaned = cleaned
                // Fix trailing commas before closing braces/brackets
                .replace(",}", "}")
                .replace(",]", "]")
                .replace(", }", "}")
                .replace(", ]", "]")
                // Fix double commas
                .replace(",,", ",")
                .replace(", ,", ",");
            if new_cleaned == cleaned {
                break;
            }
            cleaned = new_cleaned;
        }

        // Fix invalid nested brace patterns like "then": { {"kind": ...} }
        // This should become "then": [{"kind": ...}]
        // The pattern is: ": {" followed by whitespace and another "{"
        // Keep applying until no more changes
        loop {
            let fixed = Self::fix_nested_brace_objects(&cleaned);
            if fixed == cleaned {
                break;
            }
            cleaned = fixed;
        }

        cleaned
    }

    /// Fix invalid JSON patterns where LLM generates { {...} } instead of [...]
    fn fix_nested_brace_objects(json: &str) -> String {
        let chars: Vec<char> = json.chars().collect();
        let mut result = String::with_capacity(json.len());
        let mut i = 0;

        while i < chars.len() {
            // Look for pattern: ": {" followed by whitespace/newlines and "{"
            if i + 2 < chars.len() && chars[i] == ':' {
                // Check for ": {" pattern
                let mut j = i + 1;

                // Skip whitespace after colon
                while j < chars.len() && chars[j].is_whitespace() {
                    j += 1;
                }

                if j < chars.len() && chars[j] == '{' {
                    let brace_start = j;
                    j += 1;

                    // Skip whitespace after first brace
                    while j < chars.len() && chars[j].is_whitespace() {
                        j += 1;
                    }

                    // Check if next non-whitespace is another '{'
                    if j < chars.len() && chars[j] == '{' {
                        // Found the problematic pattern!
                        // We need to convert { {...} } to [{...}]

                        // Copy up to and including the colon
                        result.push(chars[i]);
                        i += 1;

                        // Copy whitespace
                        while i < brace_start {
                            result.push(chars[i]);
                            i += 1;
                        }

                        // Replace opening brace with bracket
                        result.push('[');
                        i = brace_start + 1;

                        // Skip to where the inner content starts
                        while i < chars.len() && chars[i].is_whitespace() {
                            result.push(chars[i]);
                            i += 1;
                        }

                        // Now copy content, tracking braces to find the outer closing one
                        let mut inner_depth = 0;
                        while i < chars.len() {
                            let c = chars[i];
                            if c == '{' {
                                inner_depth += 1;
                                result.push(c);
                            } else if c == '}' {
                                if inner_depth > 0 {
                                    inner_depth -= 1;
                                    result.push(c);
                                } else {
                                    // This is the outer closing brace - replace with bracket
                                    result.push(']');
                                    i += 1;
                                    break;
                                }
                            } else {
                                result.push(c);
                            }
                            i += 1;
                        }
                        continue;
                    }
                }
            }

            result.push(chars[i]);
            i += 1;
        }

        result
    }

    /// Normalize CIR JSON to handle common variations from different models.
    fn normalize_cir_json(&self, json: &str) -> String {
        // Parse as generic JSON value for manipulation
        let Ok(mut value) = serde_json::from_str::<serde_json::Value>(json) else {
            return json.to_string();
        };

        // Normalize the interpretation if present
        if let Some(interp) = value.get_mut("interpretation") {
            Self::normalize_function(interp);
        }

        serde_json::to_string(&value).unwrap_or_else(|_| json.to_string())
    }

    /// Normalize a CIR function object.
    fn normalize_function(func: &mut serde_json::Value) {
        // Normalize body operations
        if let Some(body) = func.get_mut("body") {
            if let Some(arr) = body.as_array_mut() {
                // First pass: normalize each operation
                for op in arr.iter_mut() {
                    Self::normalize_operation(op);
                }

                // Second pass: fix return statements that use literals when they should use results
                Self::fix_return_values(arr);
            }
        }
    }

    /// Fix return statements that use literal values when they should reference results.
    fn fix_return_values(body: &mut [serde_json::Value]) {
        // Find the last result variable before return
        let mut last_result: Option<String> = None;

        for op in body.iter_mut() {
            if let Some(obj) = op.as_object_mut() {
                let kind = obj.get("kind").and_then(|k| k.as_str());

                match kind {
                    Some("binary_op") | Some("call") | Some("get_field") | Some("format")
                    | Some("concat") | Some("literal") | Some("var") => {
                        // Track the result variable
                        if let Some(result) = obj.get("result").and_then(|r| r.as_str()) {
                            last_result = Some(result.to_string());
                        }
                    }
                    Some("return") => {
                        // If return has a literal bool/int and we have a previous result,
                        // the model probably meant to return the result
                        if let Some(ref result_var) = last_result {
                            if let Some(value) = obj.get("value") {
                                // Check if it's a simple literal that should be the result
                                let should_replace = match value {
                                    serde_json::Value::Bool(_) => true,
                                    serde_json::Value::String(s) => {
                                        // Don't replace if it's already a variable reference
                                        // Check if the string could be a result variable
                                        s == "true" || s == "false" || s == "result"
                                    }
                                    _ => false,
                                };

                                if should_replace {
                                    obj.insert(
                                        "value".to_string(),
                                        serde_json::Value::String(result_var.clone()),
                                    );
                                }
                            }
                        }
                    }
                    _ => {}
                }
            }
        }
    }

    /// Normalize a CIR operation.
    fn normalize_operation(op: &mut serde_json::Value) {
        if let Some(obj) = op.as_object_mut() {
            // Handle nested operation format where operation type is a key
            // e.g., {"binary_op": {"left": ..., "right": ..., "op": ...}, "eq": true}
            // should become {"kind": "binary_op", "left": ..., "right": ..., "op": ...}
            if !obj.contains_key("kind") && !obj.contains_key("op") {
                // Check for known operation type keys
                let op_types = [
                    "binary_op",
                    "unary_op",
                    "call",
                    "get_field",
                    "get_index",
                    "set_field",
                    "literal",
                    "var",
                    "return",
                    "if",
                    "loop",
                    "map",
                    "filter",
                    "reduce",
                    "sort",
                    "construct",
                    "format",
                    "concat",
                ];

                for op_type in op_types {
                    if let Some(inner) = obj.remove(op_type) {
                        // Flatten the structure: take fields from inner and add kind
                        if let Some(inner_obj) = inner.as_object() {
                            for (k, v) in inner_obj {
                                if !obj.contains_key(k) {
                                    obj.insert(k.clone(), v.clone());
                                }
                            }
                        }
                        obj.insert("kind".to_string(), serde_json::json!(op_type));
                        break;
                    }
                }
            }

            // Convert "op" to "kind" if present, BUT only if it's not a binary_op
            // (binary_op uses "op" for the operator, not the kind)
            let current_kind = obj
                .get("kind")
                .and_then(|k| k.as_str())
                .map(|s| s.to_string());
            if current_kind.as_deref() != Some("binary_op") {
                if let Some(op_val) = obj.remove("op") {
                    if !obj.contains_key("kind") {
                        obj.insert("kind".to_string(), op_val);
                    } else {
                        // Put it back since we already have a kind
                        obj.insert("op".to_string(), op_val);
                    }
                }
            }

            // Convert "operator" to "op" for binary_op (model sometimes uses this)
            if let Some(operator_val) = obj.remove("operator") {
                if !obj.contains_key("op") {
                    obj.insert("op".to_string(), operator_val);
                }
            }

            // Get the operation kind for args conversion
            let kind = obj
                .get("kind")
                .and_then(|k| k.as_str())
                .map(|s| s.to_string());

            // For call operations, ensure function is a plain string (not a {"ref": ...} object)
            if kind.as_deref() == Some("call") {
                if let Some(func) = obj.get_mut("function") {
                    // If function is {"ref": "name"}, extract the name as a plain string
                    if let Some(func_obj) = func.as_object() {
                        if let Some(ref_val) = func_obj.get("ref").and_then(|v| v.as_str()) {
                            *func = serde_json::Value::String(ref_val.to_string());
                        }
                    }
                }
            }

            // For binary_op, ensure left/right are properly structured
            if kind.as_deref() == Some("binary_op") {
                // If left/right are just strings (variable refs), wrap them in CIRValue format
                Self::normalize_binary_op_operand(obj, "left");
                Self::normalize_binary_op_operand(obj, "right");

                // Normalize operator symbols to CIR names
                if let Some(op_val) = obj.get_mut("op") {
                    if let Some(op_str) = op_val.as_str() {
                        let normalized = match op_str {
                            ">" => "gt",
                            "<" => "lt",
                            ">=" => "ge",
                            "<=" => "le",
                            "==" | "=" => "eq",
                            "!=" | "<>" => "ne",
                            "+" => "add",
                            "-" => "sub",
                            "*" => "mul",
                            "/" => "div",
                            "%" => "mod",
                            "&&" | "and" => "and",
                            "||" | "or" => "or",
                            _ => op_str, // Already normalized
                        };
                        *op_val = serde_json::Value::String(normalized.to_string());
                    }
                }

                // Ensure result field exists
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_cmp".to_string()),
                    );
                }
            }

            // Convert "args" array to named fields based on operation kind
            if let Some(args) = obj.remove("args") {
                if let Some(args_arr) = args.as_array() {
                    Self::convert_args_to_fields(obj, kind.as_deref(), args_arr);
                }
            }

            // Normalize nested operations in value fields
            if let Some(value) = obj.get_mut("value") {
                Self::normalize_value(value);
            }

            // Normalize nested values in "values" map (for Format)
            if let Some(values) = obj.get_mut("values") {
                if let Some(values_obj) = values.as_object_mut() {
                    for (_, v) in values_obj.iter_mut() {
                        Self::normalize_value(v);
                    }
                }
            }

            // Normalize condition, then_ops, else_ops for If
            for key in [
                "condition",
                "then_ops",
                "else_ops",
                "body",
                "transform",
                "predicate",
                "reducer",
                "key",
            ] {
                if let Some(nested) = obj.get_mut(key) {
                    if let Some(arr) = nested.as_array_mut() {
                        for item in arr {
                            Self::normalize_operation(item);
                        }
                    }
                }
            }

            // Normalize left/right for BinaryOp
            if let Some(left) = obj.get_mut("left") {
                Self::normalize_value(left);
            }
            if let Some(right) = obj.get_mut("right") {
                Self::normalize_value(right);
            }

            // Normalize parts for Concat
            if let Some(parts) = obj.get_mut("parts") {
                if let Some(arr) = parts.as_array_mut() {
                    for item in arr {
                        Self::normalize_value(item);
                    }
                }
            }

            // Normalize operation-specific fields (construct, filter, sort, etc.)
            Self::normalize_operation_fields(obj);
        }
    }

    /// Normalize a binary_op operand - model may return bare strings instead of proper values.
    fn normalize_binary_op_operand(
        obj: &mut serde_json::Map<String, serde_json::Value>,
        field: &str,
    ) {
        if let Some(val) = obj.get_mut(field) {
            // If it's a string that looks like a number, parse it
            if let Some(s) = val.as_str() {
                if let Ok(n) = s.parse::<i64>() {
                    *val = serde_json::Value::Number(n.into());
                }
                // Otherwise leave as string (variable reference)
            }
            // Normalize nested operations
            Self::normalize_value(val);
        }
    }

    /// Convert args array to named fields based on operation kind.
    fn convert_args_to_fields(
        obj: &mut serde_json::Map<String, serde_json::Value>,
        kind: Option<&str>,
        args: &[serde_json::Value],
    ) {
        match kind {
            Some("format") => {
                // format(template, ...values) -> { template, values: {0: v0, 1: v1, ...}, result }
                if let Some(template) = args.first() {
                    obj.insert("template".to_string(), template.clone());
                }
                if args.len() > 1 {
                    let mut values = serde_json::Map::new();
                    for (i, v) in args.iter().skip(1).enumerate() {
                        let mut normalized = v.clone();
                        Self::normalize_value(&mut normalized);
                        values.insert(i.to_string(), normalized);
                    }
                    obj.insert("values".to_string(), serde_json::Value::Object(values));
                } else {
                    obj.insert(
                        "values".to_string(),
                        serde_json::Value::Object(serde_json::Map::new()),
                    );
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_fmt".to_string()),
                    );
                }
            }
            Some("get_field") => {
                // get_field(source, field) or get_field(source) where source is "obj.field"
                if args.len() >= 2 {
                    obj.insert("source".to_string(), args[0].clone());
                    obj.insert("field".to_string(), args[1].clone());
                    if !obj.contains_key("result") {
                        obj.insert(
                            "result".to_string(),
                            serde_json::Value::String("_field".to_string()),
                        );
                    }
                } else if let Some(source) = args.first() {
                    // Try to split "source.field" format
                    if let Some(s) = source.as_str() {
                        if let Some((src, field)) = s.rsplit_once('.') {
                            obj.insert(
                                "source".to_string(),
                                serde_json::Value::String(src.to_string()),
                            );
                            obj.insert(
                                "field".to_string(),
                                serde_json::Value::String(field.to_string()),
                            );
                            if !obj.contains_key("result") {
                                obj.insert(
                                    "result".to_string(),
                                    serde_json::Value::String("_field".to_string()),
                                );
                            }
                        } else {
                            // Just a variable reference - convert to "var" operation
                            obj.remove("kind");
                            obj.insert("kind".to_string(), serde_json::json!("var"));
                            obj.insert("name".to_string(), source.clone());
                            if !obj.contains_key("result") {
                                obj.insert(
                                    "result".to_string(),
                                    serde_json::Value::String("_var".to_string()),
                                );
                            }
                        }
                    } else {
                        obj.insert("source".to_string(), source.clone());
                    }
                }
            }
            Some("concat") => {
                // concat(parts...) -> { parts, result }
                let mut parts = Vec::new();
                for arg in args {
                    let mut normalized = arg.clone();
                    Self::normalize_value(&mut normalized);
                    parts.push(normalized);
                }
                obj.insert("parts".to_string(), serde_json::Value::Array(parts));
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_concat".to_string()),
                    );
                }
            }
            Some("binary_op") => {
                // binary_op(op, left, right) -> { op, left, right, result }
                if args.len() >= 3 {
                    obj.insert("op".to_string(), args[0].clone());
                    let mut left = args[1].clone();
                    let mut right = args[2].clone();
                    Self::normalize_value(&mut left);
                    Self::normalize_value(&mut right);
                    obj.insert("left".to_string(), left);
                    obj.insert("right".to_string(), right);
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_binop".to_string()),
                    );
                }
            }
            Some("call") => {
                // call(function, ...args) -> { function, args, result }
                if let Some(func) = args.first() {
                    obj.insert("function".to_string(), func.clone());
                }
                let call_args: Vec<_> = args.iter().skip(1).cloned().collect();
                obj.insert("args".to_string(), serde_json::Value::Array(call_args));
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_call".to_string()),
                    );
                }
            }
            Some("return") => {
                // return(value) -> { value }
                if let Some(value) = args.first() {
                    let mut normalized = value.clone();
                    Self::normalize_value(&mut normalized);
                    obj.insert("value".to_string(), normalized);
                }
            }
            Some("var") => {
                // var(name) -> { name, result }
                if let Some(name) = args.first() {
                    obj.insert("name".to_string(), name.clone());
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_var".to_string()),
                    );
                }
            }
            Some("literal") => {
                // literal(value) -> { value, result }
                if let Some(value) = args.first() {
                    obj.insert("value".to_string(), value.clone());
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_lit".to_string()),
                    );
                }
            }
            _ => {
                // Unknown operation - keep args as-is for debugging
                debug!("Unknown operation kind {:?} with args, keeping as-is", kind);
            }
        }
    }

    /// Normalize operations that may have incomplete or non-standard fields.
    fn normalize_operation_fields(obj: &mut serde_json::Map<String, serde_json::Value>) {
        let kind = obj
            .get("kind")
            .and_then(|k| k.as_str())
            .map(|s| s.to_string());

        match kind.as_deref() {
            Some("construct") => {
                // Normalize type_name to type
                if let Some(type_name) = obj.remove("type_name") {
                    if !obj.contains_key("type") {
                        obj.insert("type".to_string(), type_name);
                    }
                }
                // Ensure fields exists (even if empty)
                if !obj.contains_key("fields") {
                    obj.insert(
                        "fields".to_string(),
                        serde_json::Value::Object(serde_json::Map::new()),
                    );
                }
                // Ensure result exists
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_obj".to_string()),
                    );
                }
            }
            Some("return") => {
                // Ensure return has a value
                if !obj.contains_key("value") {
                    // Default to returning none/unit
                    obj.insert("value".to_string(), serde_json::Value::Null);
                }
            }
            Some("if") => {
                // Normalize condition - model may return object instead of array
                if let Some(cond) = obj.get("condition") {
                    if cond.is_object() && !cond.is_array() {
                        let cond_clone = cond.clone();
                        obj.insert(
                            "condition".to_string(),
                            serde_json::Value::Array(vec![cond_clone]),
                        );
                    }
                }
                if !obj.contains_key("condition") {
                    obj.insert("condition".to_string(), serde_json::Value::Array(vec![]));
                }
                // Normalize condition operations
                if let Some(cond) = obj.get_mut("condition") {
                    if let Some(arr) = cond.as_array_mut() {
                        for item in arr.iter_mut() {
                            Self::normalize_operation(item);
                        }
                    }
                }
                // Ensure then_ops and else_ops exist (may be named "then" and "else")
                if let Some(then_val) = obj.remove("then") {
                    if !obj.contains_key("then_ops") {
                        obj.insert("then_ops".to_string(), then_val);
                    }
                }
                if let Some(else_val) = obj.remove("else") {
                    if !obj.contains_key("else_ops") {
                        obj.insert("else_ops".to_string(), else_val);
                    }
                }
                if !obj.contains_key("then_ops") {
                    obj.insert("then_ops".to_string(), serde_json::Value::Array(vec![]));
                }
                if !obj.contains_key("else_ops") {
                    obj.insert("else_ops".to_string(), serde_json::Value::Array(vec![]));
                }
                // Wrap single objects in arrays for then_ops/else_ops
                for key in ["then_ops", "else_ops"] {
                    if let Some(ops) = obj.get(key) {
                        if ops.is_object() && !ops.is_array() {
                            let ops_clone = ops.clone();
                            obj.insert(key.to_string(), serde_json::Value::Array(vec![ops_clone]));
                        }
                    }
                }
                // Normalize operations inside then_ops and else_ops
                for key in ["then_ops", "else_ops"] {
                    if let Some(ops) = obj.get_mut(key) {
                        if let Some(arr) = ops.as_array_mut() {
                            for item in arr.iter_mut() {
                                Self::normalize_operation(item);
                            }
                        }
                    }
                }
                // Ensure result exists
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_if".to_string()),
                    );
                }
            }
            Some("filter") => {
                // Normalize source field if it's an object like {"ref": "n"}
                Self::normalize_source_field(obj);
                if !obj.contains_key("source") {
                    obj.insert(
                        "source".to_string(),
                        serde_json::Value::String("_input".to_string()),
                    );
                }
                if !obj.contains_key("element_var") {
                    obj.insert(
                        "element_var".to_string(),
                        serde_json::Value::String("item".to_string()),
                    );
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_filtered".to_string()),
                    );
                }
                // Convert predicate object to array if needed
                if let Some(pred) = obj.get("predicate") {
                    if pred.is_object() && !pred.is_array() {
                        let pred_clone = pred.clone();
                        obj.insert(
                            "predicate".to_string(),
                            serde_json::Value::Array(vec![pred_clone]),
                        );
                    }
                }
                if !obj.contains_key("predicate") {
                    obj.insert("predicate".to_string(), serde_json::Value::Array(vec![]));
                }
                // Normalize operations inside predicate array
                if let Some(pred) = obj.get_mut("predicate") {
                    if let Some(arr) = pred.as_array_mut() {
                        for item in arr.iter_mut() {
                            Self::normalize_operation(item);
                        }
                    }
                }
            }
            Some("reduce") => {
                // Normalize source field if it's an object like {"ref": "n"}
                Self::normalize_source_field(obj);
                if !obj.contains_key("source") {
                    obj.insert(
                        "source".to_string(),
                        serde_json::Value::String("_input".to_string()),
                    );
                }
                if !obj.contains_key("element_var") {
                    obj.insert(
                        "element_var".to_string(),
                        serde_json::Value::String("item".to_string()),
                    );
                }
                if !obj.contains_key("accumulator_var") {
                    obj.insert(
                        "accumulator_var".to_string(),
                        serde_json::Value::String("acc".to_string()),
                    );
                }
                // Normalize "accumulator" to "initial"
                if let Some(acc) = obj.remove("accumulator") {
                    if !obj.contains_key("initial") {
                        obj.insert("initial".to_string(), acc);
                    }
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_reduced".to_string()),
                    );
                }
                // Convert reducer object to array if needed
                if let Some(reducer) = obj.get("reducer") {
                    if reducer.is_object() && !reducer.is_array() {
                        let reducer_clone = reducer.clone();
                        obj.insert(
                            "reducer".to_string(),
                            serde_json::Value::Array(vec![reducer_clone]),
                        );
                    }
                }
                if !obj.contains_key("reducer") {
                    obj.insert("reducer".to_string(), serde_json::Value::Array(vec![]));
                }
                // Normalize operations inside reducer array
                if let Some(reducer) = obj.get_mut("reducer") {
                    if let Some(arr) = reducer.as_array_mut() {
                        for item in arr.iter_mut() {
                            Self::normalize_operation(item);
                        }
                    }
                }
            }
            Some("get_field") => {
                // Model may return various formats:
                // {"op": "get_field", "field": "name", "value": true}
                // {"op": "get_field", "object": "$0", "value": "active"}
                // {"op": "get_field", "source": {"op": "var", "name": "user"}, "field": "active"}
                // {"op": "get_field", "field": 0, ...} - numeric field index
                // We need: {"kind": "get_field", "source": "...", "field": "...", "result": "..."}

                // Normalize "object" to "source"
                if let Some(object) = obj.remove("object") {
                    if !obj.contains_key("source") {
                        obj.insert("source".to_string(), object);
                    }
                }

                // If source is an object (like {"op": "var", "name": "user"}), extract the name
                if let Some(source) = obj.get("source") {
                    if let Some(source_obj) = source.as_object() {
                        // Check if it's a var operation
                        let is_var = source_obj.get("op").and_then(|v| v.as_str()) == Some("var")
                            || source_obj.get("kind").and_then(|v| v.as_str()) == Some("var");
                        if is_var {
                            if let Some(name) = source_obj.get("name").and_then(|n| n.as_str()) {
                                obj.insert(
                                    "source".to_string(),
                                    serde_json::Value::String(name.to_string()),
                                );
                            }
                        }
                    }
                }

                // Normalize "value" to "field" (when value is a string field name)
                if let Some(value) = obj.get("value") {
                    if value.is_string() && !obj.contains_key("field") {
                        obj.insert("field".to_string(), value.clone());
                    }
                }

                // If field is a number (index), convert to string
                if let Some(field) = obj.get("field") {
                    if let Some(n) = field.as_i64() {
                        obj.insert(
                            "field".to_string(),
                            serde_json::Value::String(n.to_string()),
                        );
                    }
                }

                // If "field" is present but "source" is not, use default source
                if obj.contains_key("field") && !obj.contains_key("source") {
                    obj.insert(
                        "source".to_string(),
                        serde_json::Value::String("item".to_string()),
                    );
                }

                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_field".to_string()),
                    );
                }
            }
            Some("sort") => {
                // Normalize sort key format
                if !obj.contains_key("source") {
                    obj.insert(
                        "source".to_string(),
                        serde_json::Value::String("_input".to_string()),
                    );
                }
                if !obj.contains_key("element_var") {
                    obj.insert(
                        "element_var".to_string(),
                        serde_json::Value::String("item".to_string()),
                    );
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_sorted".to_string()),
                    );
                }
                // Convert key object to array if needed
                if let Some(key) = obj.get("key") {
                    if key.is_object() && !key.is_array() {
                        let key_clone = key.clone();
                        obj.insert("key".to_string(), serde_json::Value::Array(vec![key_clone]));
                    }
                }
                if !obj.contains_key("key") {
                    obj.insert("key".to_string(), serde_json::Value::Array(vec![]));
                }
                // Normalize operations inside key array
                if let Some(key) = obj.get_mut("key") {
                    if let Some(arr) = key.as_array_mut() {
                        for item in arr.iter_mut() {
                            Self::normalize_operation(item);
                        }
                    }
                }
            }
            Some("loop") => {
                // Model may generate C-style loop with condition/step/body
                // instead of foreach-style loop with source/element_var/body
                // Convert: {condition: {...}, step: {...}, body: [...]}
                // To: {source: "_range", element_var: "loop_index", body: [...], result: "_loop"}

                // If condition exists but source doesn't, it's a C-style loop
                if obj.contains_key("condition") && !obj.contains_key("source") {
                    // Extract loop bounds from condition if possible
                    // For now, use a placeholder range source
                    obj.insert(
                        "source".to_string(),
                        serde_json::Value::String("_range".to_string()),
                    );

                    // Remove C-style specific fields
                    obj.remove("condition");
                    obj.remove("step");
                }

                // Normalize source field if it's an object like {"ref": "n"}
                Self::normalize_source_field(obj);

                // Ensure required fields exist
                if !obj.contains_key("source") {
                    obj.insert(
                        "source".to_string(),
                        serde_json::Value::String("_input".to_string()),
                    );
                }
                if !obj.contains_key("element_var") {
                    obj.insert(
                        "element_var".to_string(),
                        serde_json::Value::String("loop_index".to_string()),
                    );
                }
                if !obj.contains_key("result") {
                    obj.insert(
                        "result".to_string(),
                        serde_json::Value::String("_loop".to_string()),
                    );
                }

                // Normalize operations inside body array
                if let Some(body) = obj.get_mut("body") {
                    if let Some(arr) = body.as_array_mut() {
                        for item in arr.iter_mut() {
                            Self::normalize_operation(item);
                        }
                    }
                }
                if !obj.contains_key("body") {
                    obj.insert("body".to_string(), serde_json::Value::Array(vec![]));
                }
            }
            _ => {}
        }
    }

    /// Normalize source field - convert {"ref": "name"} to just "name"
    fn normalize_source_field(obj: &mut serde_json::Map<String, serde_json::Value>) {
        if let Some(source) = obj.get("source") {
            if let Some(source_obj) = source.as_object() {
                // Check if it's {"ref": "name"}
                if let Some(ref_val) = source_obj.get("ref").and_then(|v| v.as_str()) {
                    obj.insert(
                        "source".to_string(),
                        serde_json::Value::String(ref_val.to_string()),
                    );
                }
                // Check if it's {"kind": "var", "name": "x"}
                else if source_obj.get("kind").and_then(|k| k.as_str()) == Some("var") {
                    if let Some(name) = source_obj.get("name").and_then(|n| n.as_str()) {
                        obj.insert(
                            "source".to_string(),
                            serde_json::Value::String(name.to_string()),
                        );
                    }
                }
            }
        }
    }

    /// Normalize a CIR value - wrap raw literals in proper format.
    fn normalize_value(value: &mut serde_json::Value) {
        // If value is a string that looks like a number (including negative), convert it
        if let Some(s) = value.as_str() {
            // Try to parse as integer (including negative numbers like "-42")
            if let Ok(n) = s.parse::<i64>() {
                *value = serde_json::Value::Number(n.into());
                return;
            }
            // Try to parse as float
            if let Ok(f) = s.parse::<f64>() {
                if let Some(num) = serde_json::Number::from_f64(f) {
                    *value = serde_json::Value::Number(num);
                    return;
                }
            }
        }

        // If value is an object with "kind"/"op", it's an operation - normalize it
        if let Some(obj) = value.as_object_mut() {
            // First normalize the operation itself
            if obj.contains_key("op") || obj.contains_key("kind") {
                Self::normalize_operation(value);
            }

            // After normalization, check for simplifications
            if let Some(obj) = value.as_object() {
                let kind = obj.get("kind").and_then(|k| k.as_str());

                match kind {
                    Some("var") => {
                        if let Some(name) = obj.get("name").and_then(|n| n.as_str()) {
                            // Check if the name is actually a number
                            if let Ok(n) = name.parse::<i64>() {
                                *value = serde_json::Value::Number(n.into());
                                return;
                            }
                            // Convert {"kind": "var", "name": "x"} to just "x"
                            *value = serde_json::Value::String(name.to_string());
                            return;
                        }
                    }
                    Some("literal") => {
                        // Convert {"kind": "literal", "value": X} to just X
                        if let Some(inner_value) = obj.get("value") {
                            *value = inner_value.clone();
                            return;
                        }
                    }
                    _ => {}
                }
            }
        }
        // Raw values (int, string, bool) are fine as-is since CIRValue uses untagged
    }

    /// Fix recursive calls where the LLM confused variable names with function names.
    ///
    /// Common mistake: LLM generates `{"kind": "call", "function": "_n1", "args": []}`
    /// when it should be `{"kind": "call", "function": "factorial", "args": [{"ref": "_n1"}]}`
    fn fix_recursive_calls(func: &mut haira_cir::CIRFunction) {
        let func_name = func.name.clone();
        for op in &mut func.body {
            Self::fix_recursive_calls_in_op(op, &func_name);
        }
    }

    fn fix_recursive_calls_in_op(op: &mut haira_cir::CIROperation, func_name: &str) {
        use haira_cir::CIROperation;

        match op {
            CIROperation::Call { function, args, .. } => {
                // If function name starts with _ (temp var) and args is empty,
                // this is likely a confused recursive call
                if function.starts_with('_') && args.is_empty() {
                    // The function field probably contains what should be an arg
                    let var_name = function.clone();
                    *function = func_name.to_string();
                    args.push(haira_cir::CIRValue::Ref(var_name));
                }
            }
            CIROperation::If {
                condition,
                then_ops,
                else_ops,
                ..
            } => {
                for inner in condition.iter_mut() {
                    Self::fix_recursive_calls_in_op(inner, func_name);
                }
                for inner in then_ops.iter_mut() {
                    Self::fix_recursive_calls_in_op(inner, func_name);
                }
                for inner in else_ops.iter_mut() {
                    Self::fix_recursive_calls_in_op(inner, func_name);
                }
            }
            CIROperation::Loop { body, .. } => {
                for inner in body.iter_mut() {
                    Self::fix_recursive_calls_in_op(inner, func_name);
                }
            }
            CIROperation::Map { transform, .. } => {
                for inner in transform.iter_mut() {
                    Self::fix_recursive_calls_in_op(inner, func_name);
                }
            }
            CIROperation::Filter { predicate, .. } => {
                for inner in predicate.iter_mut() {
                    Self::fix_recursive_calls_in_op(inner, func_name);
                }
            }
            CIROperation::Reduce { reducer, .. } => {
                for inner in reducer.iter_mut() {
                    Self::fix_recursive_calls_in_op(inner, func_name);
                }
            }
            _ => {}
        }
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

    /// Infer types for struct fields that don't have explicit type annotations.
    ///
    /// Given a struct definition like `User { name, age, email }`, this uses AI
    /// to infer the most likely types for each field based on the field names.
    ///
    /// Returns a map of field name -> type string (e.g., "string", "int", "float", "bool").
    pub async fn infer_struct_field_types(
        &self,
        struct_name: &str,
        field_names: &[String],
    ) -> Result<std::collections::HashMap<String, String>, AIError> {
        use std::collections::HashMap;

        // Build a simple prompt for type inference
        let fields_list = field_names.join(", ");
        let prompt = format!(
            r#"Infer the most likely types for each field in the struct below.

Struct: {struct_name} {{ {fields_list} }}

For each field, determine the most appropriate type from: string, int, float, bool

Output ONLY a JSON object mapping field names to types. Example:
{{"name": "string", "age": "int", "is_active": "bool"}}

Be smart about common field names:
- "name", "email", "title", "description", etc. -> string
- "age", "count", "id", "quantity", etc. -> int
- "price", "amount", "rate", etc. -> float
- "is_*", "has_*", "active", "enabled", etc. -> bool

Output only the JSON object, nothing else:"#
        );

        let system = "You are a type inference assistant. Given field names, infer their types. Output only valid JSON.";

        let response = self.complete(system, &prompt).await?;

        // Parse the response as JSON
        let cleaned = Self::clean_llm_output(&response);

        // Find the JSON object in the response
        let json_text = if let Some(start) = cleaned.find('{') {
            if let Some(end) = cleaned.rfind('}') {
                &cleaned[start..=end]
            } else {
                &cleaned
            }
        } else {
            &cleaned
        };

        let types: HashMap<String, String> = serde_json::from_str(json_text).map_err(|e| {
            AIError::InterpretationFailed(format!("Failed to parse type inference response: {}", e))
        })?;

        // Validate that we got types for all fields
        for field in field_names {
            if !types.contains_key(field) {
                warn!(
                    "AI did not infer type for field '{}', defaulting to 'any'",
                    field
                );
            }
        }

        Ok(types)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use haira_cir::{CallSiteInfo, FieldDefinition, TypeDefinition};

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
        let config = AIConfig::builder().use_cache(false).build();
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

    #[test]
    fn test_ollama_backend() {
        let config = AIConfig::default();
        let engine = AIEngine::with_ollama(config, Some("codellama:7b"));
        assert_eq!(engine.backend(), AIBackend::Ollama);
    }
}

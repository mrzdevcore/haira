//! Agent with state-machine tool-calling loop.
//!
//! The agent sends prompts to an LLM via the chat completions API and handles
//! tool calling through a state-machine pattern. When the LLM requests a tool
//! call, the agent yields control back to the interpreter via `AgentStep::NeedTool`.
//! The interpreter executes the tool body and feeds the result back via
//! `continue_with_tool_result()`.

use crate::llm::{self, ChatMessage, ChatRequest, ToolCall};
use crate::memory::{MemoryConfig, Message, SessionStore, ToolCallInfo};
use crate::provider::Provider;
use crate::tool::ToolRegistry;

/// Result of one step in the agent's execution loop.
#[derive(Debug)]
pub enum AgentStep {
    /// The LLM needs a tool to be executed.
    NeedTool {
        tool_name: String,
        args_json: String,
        tool_call_id: String,
    },
    /// The agent has finished and produced a final text response.
    Done(String),
    /// An error occurred.
    Error(String),
}

/// An agent instance that can interact with an LLM.
#[derive(Debug, Clone)]
pub struct Agent {
    pub name: String,
    pub provider: Provider,
    pub system_prompt: String,
    pub tools: ToolRegistry,
    pub memory: SessionStore,
    pub memory_config: MemoryConfig,
    pub temperature: Option<f64>,
    pub max_tokens: Option<u32>,
    /// Pending tool calls from the last LLM response (for multi-tool-call support).
    pending_tool_calls: Vec<ToolCall>,
    /// Index into pending_tool_calls for the next tool to process.
    pending_index: usize,
    /// Accumulated messages for the current conversation turn.
    current_messages: Vec<ChatMessage>,
    /// The current session ID being used.
    current_session: String,
}

impl Agent {
    pub fn new(
        name: String,
        provider: Provider,
        system_prompt: String,
        tools: ToolRegistry,
        memory_config: MemoryConfig,
        temperature: Option<f64>,
        max_tokens: Option<u32>,
    ) -> Self {
        let max_turns = match &memory_config {
            MemoryConfig::Conversation { max_turns } => *max_turns,
            MemoryConfig::None => 0,
        };
        Self {
            name,
            provider,
            system_prompt,
            tools,
            memory: SessionStore::new(max_turns),
            memory_config,
            temperature,
            max_tokens,
            pending_tool_calls: Vec::new(),
            pending_index: 0,
            current_messages: Vec::new(),
            current_session: String::new(),
        }
    }

    /// Start a new interaction with the agent.
    ///
    /// Sends the prompt to the LLM and returns the first step:
    /// either `Done` with the response text, or `NeedTool` if the LLM
    /// wants to call a tool.
    pub fn start(&mut self, prompt: &str, session_id: &str) -> AgentStep {
        self.current_session = session_id.to_string();
        self.pending_tool_calls.clear();
        self.pending_index = 0;

        // Build messages: system + history + user prompt.
        let mut messages = Vec::new();

        // System prompt.
        if !self.system_prompt.is_empty() {
            messages.push(ChatMessage {
                role: "system".to_string(),
                content: Some(self.system_prompt.clone()),
                tool_calls: None,
                tool_call_id: None,
            });
        }

        // Conversation history.
        for msg in self.memory.get_history(session_id) {
            messages.push(memory_msg_to_chat_msg(msg));
        }

        // User message.
        messages.push(ChatMessage {
            role: "user".to_string(),
            content: Some(prompt.to_string()),
            tool_calls: None,
            tool_call_id: None,
        });

        // Save the user message to memory.
        self.memory.add_message(
            session_id,
            Message {
                role: "user".to_string(),
                content: prompt.to_string(),
                tool_calls: Vec::new(),
                tool_call_id: None,
            },
        );

        self.current_messages = messages;
        self.call_llm()
    }

    /// Continue after a tool execution result.
    ///
    /// If there are more pending tool calls, yields the next one.
    /// Otherwise, sends all tool results back to the LLM for the next response.
    pub fn continue_with_tool_result(
        &mut self,
        tool_call_id: &str,
        result: &str,
    ) -> AgentStep {
        // Add the tool result message.
        self.current_messages.push(ChatMessage {
            role: "tool".to_string(),
            content: Some(result.to_string()),
            tool_calls: None,
            tool_call_id: Some(tool_call_id.to_string()),
        });

        // Save tool result to memory.
        self.memory.add_message(
            &self.current_session,
            Message {
                role: "tool".to_string(),
                content: result.to_string(),
                tool_calls: Vec::new(),
                tool_call_id: Some(tool_call_id.to_string()),
            },
        );

        // Check if there are more pending tool calls.
        if self.pending_index < self.pending_tool_calls.len() {
            let tc = &self.pending_tool_calls[self.pending_index];
            let step = AgentStep::NeedTool {
                tool_name: tc.function.name.clone(),
                args_json: tc.function.arguments.clone(),
                tool_call_id: tc.id.clone(),
            };
            self.pending_index += 1;
            return step;
        }

        // All tool calls processed — send results back to LLM.
        self.call_llm()
    }

    /// Make an LLM API call and process the response.
    fn call_llm(&mut self) -> AgentStep {
        let tool_specs = self.tools.to_tool_specs();
        let temperature = self.temperature.or(self.provider.temperature);
        let max_tokens = self.max_tokens.or(self.provider.max_tokens);

        let request = ChatRequest {
            model: self.provider.model.clone(),
            messages: self.current_messages.clone(),
            temperature,
            max_tokens,
            tools: tool_specs,
        };

        let response = llm::chat_completion(
            &self.provider.endpoint,
            &self.provider.api_key,
            self.provider.api_version.as_deref(),
            &request,
        );

        match response {
            Err(e) => AgentStep::Error(e),
            Ok(resp) => {
                let choice = match resp.choices.first() {
                    Some(c) => c,
                    None => return AgentStep::Error("no choices in response".to_string()),
                };

                let msg = &choice.message;

                // Add assistant message to conversation.
                self.current_messages.push(msg.clone());

                // Save to memory.
                let tool_call_infos: Vec<ToolCallInfo> = msg
                    .tool_calls
                    .as_ref()
                    .map(|tcs| {
                        tcs.iter()
                            .map(|tc| ToolCallInfo {
                                id: tc.id.clone(),
                                name: tc.function.name.clone(),
                                arguments: tc.function.arguments.clone(),
                            })
                            .collect()
                    })
                    .unwrap_or_default();

                self.memory.add_message(
                    &self.current_session,
                    Message {
                        role: "assistant".to_string(),
                        content: msg.content.clone().unwrap_or_default(),
                        tool_calls: tool_call_infos,
                        tool_call_id: None,
                    },
                );

                // Check for tool calls.
                if let Some(tool_calls) = &msg.tool_calls {
                    if !tool_calls.is_empty() {
                        self.pending_tool_calls = tool_calls.clone();
                        self.pending_index = 1; // We'll yield the first one now.

                        let first = &tool_calls[0];
                        return AgentStep::NeedTool {
                            tool_name: first.function.name.clone(),
                            args_json: first.function.arguments.clone(),
                            tool_call_id: first.id.clone(),
                        };
                    }
                }

                // No tool calls — return the text response.
                let content = msg.content.clone().unwrap_or_default();
                AgentStep::Done(content)
            }
        }
    }
}

/// Convert a memory Message to a ChatMessage for the API.
fn memory_msg_to_chat_msg(msg: &Message) -> ChatMessage {
    let tool_calls = if msg.tool_calls.is_empty() {
        None
    } else {
        Some(
            msg.tool_calls
                .iter()
                .map(|tc| ToolCall {
                    id: tc.id.clone(),
                    call_type: "function".to_string(),
                    function: llm::FunctionCall {
                        name: tc.name.clone(),
                        arguments: tc.arguments.clone(),
                    },
                })
                .collect(),
        )
    };

    ChatMessage {
        role: msg.role.clone(),
        content: Some(msg.content.clone()),
        tool_calls,
        tool_call_id: msg.tool_call_id.clone(),
    }
}

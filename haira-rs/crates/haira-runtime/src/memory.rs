//! Per-session conversation memory.
//!
//! Stores conversation history per session ID, with configurable max_turns
//! trimming. Messages include role, content, and optional tool call metadata.

use std::collections::HashMap;

/// A single message in a conversation.
#[derive(Debug, Clone)]
pub struct Message {
    pub role: String,
    pub content: String,
    /// Tool calls requested by the assistant (if any).
    pub tool_calls: Vec<ToolCallInfo>,
    /// Tool call ID this message is responding to (for tool role messages).
    pub tool_call_id: Option<String>,
}

/// Information about a tool call in an assistant message.
#[derive(Debug, Clone)]
pub struct ToolCallInfo {
    pub id: String,
    pub name: String,
    pub arguments: String,
}

/// Memory configuration for an agent.
#[derive(Debug, Clone)]
pub enum MemoryConfig {
    /// No memory — each call is stateless.
    None,
    /// Conversation memory with a maximum number of turns.
    Conversation { max_turns: usize },
}

impl Default for MemoryConfig {
    fn default() -> Self {
        MemoryConfig::None
    }
}

/// Per-session conversation history store.
#[derive(Debug, Clone, Default)]
pub struct SessionStore {
    sessions: HashMap<String, Vec<Message>>,
    max_turns: usize,
}

impl SessionStore {
    pub fn new(max_turns: usize) -> Self {
        Self {
            sessions: HashMap::new(),
            max_turns,
        }
    }

    /// Get the conversation history for a session.
    pub fn get_history(&self, session_id: &str) -> &[Message] {
        self.sessions
            .get(session_id)
            .map(|v| v.as_slice())
            .unwrap_or(&[])
    }

    /// Add a message to a session's history, trimming if necessary.
    pub fn add_message(&mut self, session_id: &str, msg: Message) {
        let history = self.sessions.entry(session_id.to_string()).or_default();
        history.push(msg);

        // Trim to max_turns * 2 messages (each turn = user + assistant).
        if self.max_turns > 0 {
            let max_messages = self.max_turns * 2;
            if history.len() > max_messages {
                let excess = history.len() - max_messages;
                history.drain(..excess);
            }
        }
    }

    /// Add multiple messages (e.g., tool call + tool response pairs).
    pub fn add_messages(&mut self, session_id: &str, msgs: Vec<Message>) {
        for msg in msgs {
            self.add_message(session_id, msg);
        }
    }
}

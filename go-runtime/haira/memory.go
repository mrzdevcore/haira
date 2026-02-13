package haira

import "sync"

// MemoryConfig configures agent memory behavior.
type MemoryConfig struct {
	Kind     string // "conversation", "summary", "none"
	MaxTurns int
}

// Message represents a single message in a conversation.
type Message struct {
	Role       string            `json:"role"` // "system", "user", "assistant", "tool"
	Content    string            `json:"content"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCallRequest `json:"tool_calls,omitempty"`
}

// ToolCallRequest represents an LLM request to call a tool.
type ToolCallRequest struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // always "function"
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction holds the function name and JSON arguments.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// SessionStore manages conversation histories per session.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]Message
	maxTurns int
}

// NewSessionStore creates a new session store with a max turn limit.
func NewSessionStore(maxTurns int) *SessionStore {
	return &SessionStore{
		sessions: make(map[string][]Message),
		maxTurns: maxTurns,
	}
}

// GetHistory returns the conversation history for a session.
func (s *SessionStore) GetHistory(sessionID string) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}
	result := make([]Message, len(msgs))
	copy(result, msgs)
	return result
}

// AddMessage appends a message to a session's history, trimming if needed.
func (s *SessionStore) AddMessage(sessionID string, msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = append(s.sessions[sessionID], msg)
	// Trim to max turns (each turn = user + assistant = 2 messages)
	maxMessages := s.maxTurns * 2
	if maxMessages > 0 && len(s.sessions[sessionID]) > maxMessages {
		s.sessions[sessionID] = s.sessions[sessionID][len(s.sessions[sessionID])-maxMessages:]
	}
}

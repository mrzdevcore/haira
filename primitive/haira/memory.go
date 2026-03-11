package haira

import (
	"fmt"
	"strings"
	"sync"
)

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

// SessionContext tracks key facts about a session for long-term recall.
// This is injected into the system prompt so the LLM always knows what
// happened earlier, even after messages are compacted.
type SessionContext struct {
	FilesRead    []string // file paths read during the session
	FilesWritten []string // file paths written/edited
	CommandsRun  []string // commands executed
	KeyFacts     []string // important observations (errors, decisions)
	StoredFiles  []string // persistent file references (id:name) — survives compaction
}

// AddFileRead records a file read (deduplicates).
func (sc *SessionContext) AddFileRead(path string) {
	for _, p := range sc.FilesRead {
		if p == path {
			return
		}
	}
	sc.FilesRead = append(sc.FilesRead, path)
}

// AddFileWritten records a file write (deduplicates).
func (sc *SessionContext) AddFileWritten(path string) {
	for _, p := range sc.FilesWritten {
		if p == path {
			return
		}
	}
	sc.FilesWritten = append(sc.FilesWritten, path)
}

// AddCommand records a command execution (keeps last 20).
func (sc *SessionContext) AddCommand(cmd string) {
	// Truncate long commands
	if len(cmd) > 80 {
		cmd = cmd[:77] + "..."
	}
	sc.CommandsRun = append(sc.CommandsRun, cmd)
	if len(sc.CommandsRun) > 20 {
		sc.CommandsRun = sc.CommandsRun[len(sc.CommandsRun)-20:]
	}
}

// AddKeyFact records an important observation (keeps last 10).
func (sc *SessionContext) AddKeyFact(fact string) {
	sc.KeyFacts = append(sc.KeyFacts, fact)
	if len(sc.KeyFacts) > 10 {
		sc.KeyFacts = sc.KeyFacts[len(sc.KeyFacts)-10:]
	}
}

// AddStoredFile records a persistent file reference (deduplicates by ID).
func (sc *SessionContext) AddStoredFile(id, name string) {
	ref := id + ":" + name
	for _, f := range sc.StoredFiles {
		if f == ref {
			return
		}
	}
	sc.StoredFiles = append(sc.StoredFiles, ref)
}

// String returns a compact summary for injection into the system prompt.
func (sc *SessionContext) String() string {
	if len(sc.FilesRead) == 0 && len(sc.FilesWritten) == 0 &&
		len(sc.CommandsRun) == 0 && len(sc.KeyFacts) == 0 &&
		len(sc.StoredFiles) == 0 {
		return ""
	}

	var parts []string
	if len(sc.StoredFiles) > 0 {
		parts = append(parts, "Stored files (persistent, use list_session_files/get_artifact/restore_file to access): "+strings.Join(sc.StoredFiles, ", "))
	}
	if len(sc.FilesRead) > 0 {
		parts = append(parts, "Files read: "+strings.Join(sc.FilesRead, ", "))
	}
	if len(sc.FilesWritten) > 0 {
		parts = append(parts, "Files modified: "+strings.Join(sc.FilesWritten, ", "))
	}
	if len(sc.CommandsRun) > 0 {
		parts = append(parts, "Commands run: "+strings.Join(sc.CommandsRun, "; "))
	}
	if len(sc.KeyFacts) > 0 {
		parts = append(parts, "Notes: "+strings.Join(sc.KeyFacts, "; "))
	}
	return "\n\nSESSION CONTEXT (what happened earlier in this conversation):\n" +
		strings.Join(parts, "\n")
}

// recentWindowSize is the number of recent messages kept in full detail.
// Older messages are compacted into single-line summaries.
const recentWindowSize = 16 // ~8 turns (user+assistant pairs)

// SessionStore manages conversation histories per session.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string][]Message
	contexts map[string]*SessionContext
	maxTurns int
	disabled bool // when true (memory: none), no history is stored
}

// NewSessionStore creates a new session store with a max turn limit.
func NewSessionStore(maxTurns int) *SessionStore {
	return &SessionStore{
		sessions: make(map[string][]Message),
		contexts: make(map[string]*SessionContext),
		maxTurns: maxTurns,
	}
}

// GetContext returns the session context for a session (creates if needed).
func (s *SessionStore) GetContext(sessionID string) *SessionContext {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, ok := s.contexts[sessionID]
	if !ok {
		ctx = &SessionContext{}
		s.contexts[sessionID] = ctx
	}
	return ctx
}

// GetHistory returns the conversation history for a session.
// Recent messages are returned in full; older messages are compacted
// into summary lines to save context window space.
func (s *SessionStore) GetHistory(sessionID string) []Message {
	if s.disabled {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}

	// If history fits in the recent window, return as-is
	if len(msgs) <= recentWindowSize {
		result := make([]Message, len(msgs))
		copy(result, msgs)
		return result
	}

	// Compact older messages into summaries
	cutoff := len(msgs) - recentWindowSize
	var compacted []string
	for i := 0; i < cutoff; i++ {
		msg := msgs[i]
		summary := compactMessage(msg)
		if summary != "" {
			compacted = append(compacted, summary)
		}
	}

	// Build result: one summary message + recent messages in full
	var result []Message
	if len(compacted) > 0 {
		result = append(result, Message{
			Role:    "system",
			Content: "Summary of earlier conversation:\n" + strings.Join(compacted, "\n"),
		})
	}

	recent := msgs[cutoff:]
	for _, msg := range recent {
		result = append(result, msg)
	}

	return result
}

// compactMessage converts a message into a one-line summary.
func compactMessage(msg Message) string {
	switch msg.Role {
	case "user":
		content := msg.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		return fmt.Sprintf("- User: %s", content)
	case "assistant":
		content := msg.Content
		// Extract just the tool completion prefix if present
		if strings.HasPrefix(content, "[Completed:") {
			end := strings.Index(content, "]\n")
			if end > 0 {
				toolPart := content[:end+1]
				textPart := content[end+2:]
				if len(textPart) > 100 {
					textPart = textPart[:97] + "..."
				}
				return fmt.Sprintf("- Assistant: %s %s", toolPart, textPart)
			}
		}
		if len(content) > 150 {
			content = content[:147] + "..."
		}
		return fmt.Sprintf("- Assistant: %s", content)
	default:
		return ""
	}
}

const maxSessions = 1000

// AddMessage appends a message to a session's history, trimming if needed.
func (s *SessionStore) AddMessage(sessionID string, msg Message) {
	if s.disabled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = append(s.sessions[sessionID], msg)
	// Hard cap: keep at most maxTurns*2 raw messages
	maxMessages := s.maxTurns * 2
	if maxMessages > 0 && len(s.sessions[sessionID]) > maxMessages {
		s.sessions[sessionID] = s.sessions[sessionID][len(s.sessions[sessionID])-maxMessages:]
	}
	// Evict a random old session if we exceed the limit
	if len(s.sessions) > maxSessions {
		for id := range s.sessions {
			if id != sessionID {
				delete(s.sessions, id)
				delete(s.contexts, id)
				break
			}
		}
	}
}

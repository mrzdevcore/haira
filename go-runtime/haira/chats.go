package haira

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"
)

// ChatSession is the compact summary returned by the list endpoint.
type ChatSession struct {
	ID           string    `json:"id"`
	WorkflowName string    `json:"workflow_name"`
	WorkflowPath string    `json:"workflow_path"`
	Title        string    `json:"title"`
	Owner        string    `json:"owner,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
}

// ChatSessionDetail includes the full message history.
type ChatSessionDetail struct {
	ChatSession
	Messages []ChatMessage `json:"messages"`
}

// ChatMessage is a single message in a chat session.
type ChatMessage struct {
	Role      string          `json:"role"` // "user" or "assistant"
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
	UIEvents  json.RawMessage `json:"ui_events,omitempty"`
}

type chatStore struct {
	mu          sync.RWMutex
	sessions    map[string]*ChatSessionDetail
	order       []string // session IDs, newest first
	maxSessions int
}

var globalChatStore = &chatStore{
	sessions:    make(map[string]*ChatSessionDetail),
	maxSessions: 100,
}

func init() {
	if v := os.Getenv("HAIRA_MAX_CHAT_SESSIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			globalChatStore.maxSessions = n
		}
	}
}

// EnsureSession creates a chat session if it doesn't already exist.
func (s *chatStore) EnsureSession(id, wfName, wfPath, owner string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; exists {
		return
	}

	now := time.Now()
	s.sessions[id] = &ChatSessionDetail{
		ChatSession: ChatSession{
			ID:           id,
			WorkflowName: wfName,
			WorkflowPath: wfPath,
			Owner:        owner,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		Messages: []ChatMessage{},
	}
	// Prepend to order (newest first)
	s.order = append([]string{id}, s.order...)
	s.evictLocked()
}

// AddMessage appends a message to a session. Auto-sets title from first user message.
func (s *chatStore) AddMessage(sessionID, role, content string) {
	s.mu.Lock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return
	}

	now := time.Now()
	sess.Messages = append(sess.Messages, ChatMessage{
		Role:      role,
		Content:   content,
		Timestamp: now,
	})
	sess.MessageCount = len(sess.Messages)
	sess.UpdatedAt = now

	// Auto-title from first user message
	if sess.Title == "" && role == "user" && content != "" {
		title := content
		if len(title) > 80 {
			title = title[:80] + "..."
		}
		sess.Title = title
	}

	// Move to front of order
	s.moveToFrontLocked(sessionID)
	s.mu.Unlock()

	s.persist()
}

// AddMessageWithUI appends a message with associated UI render events.
func (s *chatStore) AddMessageWithUI(sessionID, role, content string, uiEvents []json.RawMessage) {
	if len(uiEvents) == 0 {
		s.AddMessage(sessionID, role, content)
		return
	}

	s.mu.Lock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return
	}

	eventsJSON, _ := json.Marshal(uiEvents)

	now := time.Now()
	sess.Messages = append(sess.Messages, ChatMessage{
		Role:      role,
		Content:   content,
		Timestamp: now,
		UIEvents:  eventsJSON,
	})
	sess.MessageCount = len(sess.Messages)
	sess.UpdatedAt = now

	s.moveToFrontLocked(sessionID)
	s.mu.Unlock()

	s.persist()
}

// GetSession returns the full session with messages, or nil.
func (s *chatStore) GetSession(id string) *ChatSessionDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

// ListSessions returns summaries, newest first. Filters are optional (empty = no filter).
func (s *chatStore) ListSessions(wfPath, owner string) []ChatSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ChatSession, 0, len(s.order))
	for _, id := range s.order {
		sess := s.sessions[id]
		if sess == nil {
			continue
		}
		if wfPath != "" && sess.WorkflowPath != wfPath {
			continue
		}
		if owner != "" && sess.Owner != owner {
			continue
		}
		result = append(result, sess.ChatSession)
	}
	return result
}

// DeleteSession removes a session by ID.
func (s *chatStore) DeleteSession(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	for i, sid := range s.order {
		if sid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	s.mu.Unlock()

	s.persist()
}

func (s *chatStore) moveToFrontLocked(id string) {
	for i, sid := range s.order {
		if sid == id {
			if i == 0 {
				return
			}
			s.order = append(s.order[:i], s.order[i+1:]...)
			s.order = append([]string{id}, s.order...)
			return
		}
	}
}

func (s *chatStore) evictLocked() {
	for len(s.order) > s.maxSessions {
		oldID := s.order[len(s.order)-1]
		delete(s.sessions, oldID)
		s.order = s.order[:len(s.order)-1]
	}
}

// --- File persistence ---

const chatsFile = ".haira-chats.json"

func (s *chatStore) persist() {
	s.mu.RLock()
	// Build ordered list for serialization
	sessions := make([]*ChatSessionDetail, 0, len(s.order))
	for _, id := range s.order {
		if sess := s.sessions[id]; sess != nil {
			sessions = append(sessions, sess)
		}
	}
	data, err := json.MarshalIndent(sessions, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return
	}
	os.WriteFile(chatsFile, data, 0644)
}

func (s *chatStore) load() {
	data, err := os.ReadFile(chatsFile)
	if err != nil {
		return
	}
	var sessions []*ChatSessionDetail
	if json.Unmarshal(data, &sessions) != nil {
		return
	}
	s.mu.Lock()
	s.sessions = make(map[string]*ChatSessionDetail, len(sessions))
	s.order = make([]string, 0, len(sessions))
	for _, sess := range sessions {
		s.sessions[sess.ID] = sess
		s.order = append(s.order, sess.ID)
	}
	s.mu.Unlock()
}

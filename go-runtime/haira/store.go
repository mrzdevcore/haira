package haira

import (
	"encoding/json"
	"os"
	"time"
)

// Store is the unified storage interface for the Haira runtime.
// Implementations: SQLiteStore (default, embedded) and PostgresStore (production).
type Store interface {
	// Lifecycle
	Init() error
	Close() error

	// Chat sessions
	EnsureSession(id, wfName, wfPath, owner string) error
	AddMessage(sessionID, role, content string, uiEvents []json.RawMessage) error
	GetSession(id string) (*ChatSessionDetail, error)
	ListSessions(wfPath, owner string) ([]ChatSession, error)
	DeleteSession(id string) error

	// Runs
	CreateRun(run *Run) error
	UpdateRun(run *Run) error
	GetRun(id string) (*Run, error)
	ListRuns(wfPath string) ([]RunSummary, error)
}

// --- Data types (shared by all backends) ---

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

// --- Global store ---

var globalStore Store

// InitStore initializes the global store based on environment.
// If HAIRA_DATABASE_URL is set, uses Postgres. Otherwise uses SQLite.
func InitStore() error {
	dbURL := os.Getenv("HAIRA_DATABASE_URL")
	if dbURL != "" {
		globalStore = NewPostgresStore(dbURL)
	} else {
		globalStore = NewSQLiteStore(".haira.db")
	}
	return globalStore.Init()
}

// CloseStore closes the global store.
func CloseStore() error {
	if globalStore != nil {
		return globalStore.Close()
	}
	return nil
}

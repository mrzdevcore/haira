package haira

import (
	"encoding/json"
	"os"
	"time"
)

// Store is the unified storage interface for the Haira runtime.
// Implementations: SQLiteStore (default), D1Store (Cloudflare Workers), and PostgresStore (production).
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

	// Observability
	SaveGeneration(gen LLMGeneration) error
	SaveToolExec(exec ToolExec) error
	LoadGenerations() ([]LLMGeneration, error)
	LoadToolExecs() ([]ToolExec, error)
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

// --- Store backend registry ---

// storeBackends maps backend names to factory functions.
// Backends register themselves via init() in their respective files.
var storeBackends = map[string]func(string) Store{}

// RegisterStoreBackend registers a store backend factory.
func RegisterStoreBackend(name string, factory func(string) Store) {
	storeBackends[name] = factory
}

// --- Global store ---

var globalStore Store
var storeURL string // set programmatically via store.database() in Haira code

// SetStoreURL sets the database URL for the runtime store.
// Must be called before InitStore (i.e., before http.Server()).
func SetStoreURL(url string) {
	storeURL = url
}

// InitStore initializes the global store based on environment.
// Priority: programmatic URL (store.database()) > HAIRA_DATABASE_URL > D1 (workers) > SQLite (native default).
func InitStore() error {
	dbURL := storeURL
	if dbURL == "" {
		dbURL = os.Getenv("HAIRA_DATABASE_URL")
	}
	if dbURL != "" {
		if factory, ok := storeBackends["postgres"]; ok {
			globalStore = factory(dbURL)
		}
	} else if factory, ok := storeBackends["d1"]; ok {
		// D1 backend registered → workers target, use Cloudflare D1
		globalStore = factory("DB") // "DB" = binding name in wrangler.toml
	} else {
		if factory, ok := storeBackends["sqlite"]; ok {
			globalStore = factory(".haira.db")
		}
	}
	if globalStore == nil {
		return nil
	}
	if err := globalStore.Init(); err != nil {
		return err
	}
	// Mark any stale "running" runs as "failed" — these are leftovers from
	// a previous server process that was killed or crashed.
	cleanupStaleRuns()
	ObserveLoadFromStore()
	return nil
}

func cleanupStaleRuns() {
	runs, err := globalStore.ListRuns("")
	if err != nil {
		return
	}
	for _, r := range runs {
		if r.Status == RunStatusRunning {
			full, err := globalStore.GetRun(r.ID)
			if err != nil || full == nil {
				continue
			}
			full.Status = RunStatusFailed
			full.Error = "interrupted: server restarted"
			now := time.Now()
			full.FinishedAt = &now
			globalStore.UpdateRun(full)
		}
	}
}

// CloseStore closes the global store.
func CloseStore() error {
	if globalStore != nil {
		return globalStore.Close()
	}
	return nil
}

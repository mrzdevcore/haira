package arp

import (
	"encoding/json"
	"net/http"
)

// SessionStore is an optional interface for chat session persistence.
// Implementing this enables session history, sidebar, and reconnection.
// The ARP server works without it (stateless mode).
type SessionStore interface {
	EnsureSession(id, workflowName, workflowPath, owner string) error
	AddMessage(sessionID, role, content string, uiEvents []json.RawMessage) error
	GetSession(id string) (*ChatSessionDetail, error)
	ListSessions(workflowPath, owner string) ([]ChatSession, error)
	DeleteSession(id string) error
}

// ChatSession is the compact summary returned by the list endpoint.
type ChatSession struct {
	ID           string `json:"id"`
	WorkflowName string `json:"workflow_name"`
	WorkflowPath string `json:"workflow_path"`
	Title        string `json:"title"`
	Owner        string `json:"owner,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	MessageCount int    `json:"message_count"`
}

// ChatSessionDetail includes the full message history.
type ChatSessionDetail struct {
	ChatSession
	Messages []ChatMessageEntry `json:"messages"`
}

// ChatMessageEntry is a single message in a chat session.
type ChatMessageEntry struct {
	Role      string          `json:"role"` // "user" or "assistant"
	Content   string          `json:"content"`
	Timestamp string          `json:"timestamp"`
	UIEvents  json.RawMessage `json:"ui_events,omitempty"`
}

// RegisterSessionRoutes adds chat session management API routes to a ServeMux:
//
//	GET    /_api/chats?workflow=<path>&owner=<owner>  — list sessions
//	GET    /_api/chats/<id>                           — get session detail
//	DELETE /_api/chats/<id>                           — delete session
func RegisterSessionRoutes(mux *http.ServeMux, store SessionStore) {
	mux.HandleFunc("/_api/chats", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			wfPath := r.URL.Query().Get("workflow")
			owner := r.URL.Query().Get("owner")
			sessions, err := store.ListSessions(wfPath, owner)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(rw).Encode(sessions)
		default:
			http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/_api/chats/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		id := r.URL.Path[len("/_api/chats/"):]
		if id == "" {
			http.Error(rw, "missing session id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			session, err := store.GetSession(id)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			if session == nil {
				http.Error(rw, "not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(rw).Encode(session)

		case http.MethodDelete:
			if err := store.DeleteSession(id); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(rw).Encode(map[string]bool{"ok": true})

		default:
			http.Error(rw, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

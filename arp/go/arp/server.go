package arp

import (
	"encoding/json"
	"net/http"
)

// ServerConfig configures an ARP HTTP server.
type ServerConfig struct {
	// Capabilities to advertise to clients. If zero value, uses DefaultCapabilities().
	Capabilities ArpHello

	// InputHandler handles user input from both WebSocket and SSE transports.
	// Required for the server to function.
	InputHandler InputHandler

	// Persist provides optional session persistence callbacks.
	// If nil, the server operates in stateless mode.
	Persist *PersistenceCallbacks

	// SessionStore enables the /_api/chats endpoints for session management.
	// If nil, session routes are not registered.
	SessionStore SessionStore

	// CORSOrigin sets the Access-Control-Allow-Origin header. Default: "*".
	CORSOrigin string
}

// ArpServer is an HTTP server with ARP protocol support.
type ArpServer struct {
	Mux    *http.ServeMux
	config ServerConfig
}

// NewServer creates an ARP-enabled HTTP server.
// It registers:
//
//	/_arp/v1   — WebSocket endpoint
//	/_api/arp  — Capability discovery (JSON)
//
// Additional routes can be added to the returned ArpServer.Mux.
func NewServer(config ServerConfig) *ArpServer {
	if config.Capabilities.V == 0 {
		config.Capabilities = DefaultCapabilities()
	}

	s := &ArpServer{
		Mux:    http.NewServeMux(),
		config: config,
	}

	// ARP WebSocket endpoint
	s.Mux.HandleFunc("/_arp/v1", WebSocketHandler(config.Capabilities, config.InputHandler, config.Persist))

	// Capability discovery endpoint
	s.Mux.HandleFunc("/_api/arp", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(config.Capabilities)
	})

	// Session routes (if store is provided)
	if config.SessionStore != nil {
		RegisterSessionRoutes(s.Mux, config.SessionStore)
	}

	return s
}

// ListenAndServe starts the ARP server on the given address.
// The address should be in the form ":8080" or "localhost:8080".
func (s *ArpServer) ListenAndServe(addr string) error {
	handler := CORSMiddleware(s.Mux, s.config.CORSOrigin)
	return http.ListenAndServe(addr, handler)
}

// Handler returns the server's HTTP handler (with CORS middleware).
// Use this to mount the ARP server as part of a larger application.
func (s *ArpServer) Handler() http.Handler {
	return CORSMiddleware(s.Mux, s.config.CORSOrigin)
}

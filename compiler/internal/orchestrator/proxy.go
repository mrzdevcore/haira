package orchestrator

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ReverseProxy routes requests by deployment name prefix to backend ports.
type ReverseProxy struct {
	mu     sync.RWMutex
	routes map[string]int // name → port
}

// NewReverseProxy creates a new reverse proxy router.
func NewReverseProxy() *ReverseProxy {
	return &ReverseProxy{
		routes: make(map[string]int),
	}
}

// AddRoute registers a deployment route.
func (rp *ReverseProxy) AddRoute(name string, port int) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.routes[name] = port
}

// RemoveRoute deregisters a deployment route.
func (rp *ReverseProxy) RemoveRoute(name string) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	delete(rp.routes, name)
}

// ServeHTTP implements http.Handler. Routes /<name>/... → localhost:<port>/...
func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse: /my-project/api/chat → name="my-project", rest="/api/chat"
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		// Root path — redirect to console UI
		http.Redirect(w, r, "/_ui/", http.StatusTemporaryRedirect)
		return
	}

	name := parts[0]
	rest := "/"
	if len(parts) > 1 {
		rest = "/" + parts[1]
	}

	rp.mu.RLock()
	port, exists := rp.routes[name]
	rp.mu.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf(`{"error": "deployment %q not found"}`, name), http.StatusNotFound)
		return
	}

	// Redirect bare root to /_ui/ (the Haira server's web UI)
	if rest == "/" {
		http.Redirect(w, r, "/"+name+"/_ui/", http.StatusTemporaryRedirect)
		return
	}

	// Check for WebSocket upgrade
	if isWebSocketUpgrade(r) {
		rp.handleWebSocket(w, r, port, rest)
		return
	}

	target, _ := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = rest
			req.URL.RawQuery = r.URL.RawQuery
			req.Host = target.Host
			req.Header.Set("X-Forwarded-For", r.RemoteAddr)
			req.Header.Set("X-Forwarded-Host", r.Host)
			req.Header.Set("X-Forwarded-Prefix", "/"+name)
		},
		// Stream SSE without buffering
		FlushInterval: -1,
	}

	proxy.ServeHTTP(w, r)
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func (rp *ReverseProxy) handleWebSocket(w http.ResponseWriter, r *http.Request, port int, path string) {
	backendAddr := fmt.Sprintf("localhost:%d", port)
	backendConn, err := net.DialTimeout("tcp", backendAddr, 5*time.Second)
	if err != nil {
		http.Error(w, `{"error": "backend unavailable"}`, http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, `{"error": "websocket not supported"}`, http.StatusInternalServerError)
		return
	}

	clientConn, clientBuf, err := hj.Hijack()
	if err != nil {
		http.Error(w, `{"error": "hijack failed"}`, http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Forward the original request to backend with rewritten path
	r.URL.Path = path
	r.Write(backendConn)

	// Flush any buffered data
	if clientBuf.Reader.Buffered() > 0 {
		buffered := make([]byte, clientBuf.Reader.Buffered())
		clientBuf.Read(buffered)
		backendConn.Write(buffered)
	}

	// Bidirectional copy
	done := make(chan struct{})
	go func() {
		io.Copy(backendConn, clientConn)
		close(done)
	}()
	io.Copy(clientConn, backendConn)
	<-done
}

// Routes returns a copy of the current route map.
func (rp *ReverseProxy) Routes() map[string]int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	out := make(map[string]int, len(rp.routes))
	for k, v := range rp.routes {
		out[k] = v
	}
	return out
}

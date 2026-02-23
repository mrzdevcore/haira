// Package webui implements the `haira webui` command.
// It serves the Haira UI SDK (embedded in the compiler binary) and proxies
// all other requests to a running Haira backend (which is a pure ARP server
// with no UI of its own).
//
// When multiple backends are provided, requests are namespaced under
// /_b/<name>/ prefixes. The UI metadata merges workflows from all backends.
package webui

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/haira-lang/haira/internal/runtime"
)

// Backend represents a single connected Haira server.
type Backend struct {
	Name      string
	URL       *url.URL
	Proxy     *httputil.ReverseProxy
	mu        sync.RWMutex
	Workflows []map[string]any
}

// Run starts the webui server. If a single backend is provided, it uses the
// original single-backend proxy logic. Multiple backends use namespaced routing.
func Run(backends []string, port int) error {
	cliFiles := runtime.CLIFiles()
	uiJS, ok := cliFiles["haira-ui.js"]
	if !ok {
		return fmt.Errorf("UI bundle not found in compiler binary (rebuild with `make build`)")
	}
	loaderHTML := string(cliFiles["loader.html"])
	if loaderHTML == "" {
		return fmt.Errorf("loader.html not found in compiler binary (rebuild with `make build`)")
	}

	// Single backend — use original simple proxy (100% backward compatible)
	if len(backends) == 1 {
		return runSingle(backends[0], port, uiJS, loaderHTML)
	}

	return runMulti(backends, port, uiJS, loaderHTML)
}

// ---------------------------------------------------------------------------
// Single-backend mode (unchanged from original)
// ---------------------------------------------------------------------------

func runSingle(backend string, port int, uiJS []byte, loaderHTML string) error {
	if !strings.HasPrefix(backend, "http") {
		backend = "http://" + backend
	}
	backendURL, err := url.Parse(backend)
	if err != nil {
		return fmt.Errorf("invalid backend URL %q: %w", backend, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.FlushInterval = -1
	proxy.ErrorHandler = proxyErrorHandler

	mux := http.NewServeMux()

	mux.HandleFunc("/_ui/assets/haira-ui.js", serveUIJS(uiJS))
	mux.HandleFunc("/_ui/config", serveUIConfig)

	for _, prefix := range []string{"/_api/", "/_arp/", "/_observe/"} {
		p := prefix
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/html") {
			meta := buildPageMetaSingle(backend, r.URL.Path)
			page := strings.Replace(loaderHTML, "{{META}}", meta, 1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(page))
			return
		}
		proxy.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Fprintf(os.Stderr, "\n  haira webui\n")
	fmt.Fprintf(os.Stderr, "  Local:   http://localhost:%d/\n", port)
	fmt.Fprintf(os.Stderr, "  Backend: %s\n", backend)
	fmt.Fprintf(os.Stderr, "  ARP:     ws://%s/_arp/v1\n\n", backendURL.Host)

	return http.ListenAndServe(addr, mux)
}

// ---------------------------------------------------------------------------
// Multi-backend mode
// ---------------------------------------------------------------------------

func runMulti(backendAddrs []string, port int, uiJS []byte, loaderHTML string) error {
	var allBackends []*Backend

	for _, addr := range backendAddrs {
		if !strings.HasPrefix(addr, "http") {
			addr = "http://" + addr
		}
		u, err := url.Parse(addr)
		if err != nil {
			return fmt.Errorf("invalid backend URL %q: %w", addr, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.FlushInterval = -1
		proxy.ErrorHandler = proxyErrorHandler

		b := &Backend{
			URL:   u,
			Proxy: proxy,
		}
		allBackends = append(allBackends, b)
	}

	// Fetch workflows from each backend and derive names
	usedNames := map[string]bool{}
	for _, b := range allBackends {
		wfs := fetchWorkflows(b.URL.String())
		b.Workflows = wfs
		name := deriveBackendName(wfs, b.URL)
		name = uniqueName(name, usedNames)
		usedNames[name] = true
		b.Name = name
	}

	// Build lookup map
	backendMap := map[string]*Backend{}
	for _, b := range allBackends {
		backendMap[b.Name] = b
	}

	// Start background refresh
	go refreshLoop(allBackends, 30*time.Second)

	mux := http.NewServeMux()

	// Static UI assets
	mux.HandleFunc("/_ui/assets/haira-ui.js", serveUIJS(uiJS))
	mux.HandleFunc("/_ui/config", serveUIConfig)

	// Namespaced backend proxy: /_b/<name>/...
	mux.HandleFunc("/_b/", func(w http.ResponseWriter, r *http.Request) {
		// Parse /_b/<name>/rest...
		trimmed := strings.TrimPrefix(r.URL.Path, "/_b/")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, `{"error":"missing backend name"}`, http.StatusBadRequest)
			return
		}
		name := parts[0]
		rest := "/"
		if len(parts) > 1 {
			rest = "/" + parts[1]
		}

		b, ok := backendMap[name]
		if !ok {
			http.Error(w, fmt.Sprintf(`{"error":"backend %q not found"}`, name), http.StatusNotFound)
			return
		}

		// WebSocket upgrade — use TCP-level proxy
		if isWebSocketUpgrade(r) {
			handleWebSocket(w, r, b.URL.Host, rest)
			return
		}

		// Rewrite path and proxy
		r.URL.Path = rest
		r.URL.RawPath = ""
		b.Proxy.ServeHTTP(w, r)
	})

	// Aggregated /_api/workflows — merge from all backends
	mux.HandleFunc("/_api/workflows", func(w http.ResponseWriter, r *http.Request) {
		var merged []map[string]any
		for _, b := range allBackends {
			b.mu.RLock()
			wfs := b.Workflows
			b.mu.RUnlock()
			for _, wf := range wfs {
				item := namespacedWorkflowItem(wf, b.Name)
				merged = append(merged, item)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(merged)
	})

	// Aggregated /_api/chats — route to correct backend via ?workflow= param
	mux.HandleFunc("/_api/chats", func(w http.ResponseWriter, r *http.Request) {
		wfParam := r.URL.Query().Get("workflow")
		if name, rest, ok := parseBackendPrefix(wfParam); ok {
			if b, exists := backendMap[name]; exists {
				q := r.URL.Query()
				q.Set("workflow", rest)
				r.URL.RawQuery = q.Encode()
				b.Proxy.ServeHTTP(w, r)
				return
			}
		}
		// No workflow param or unknown backend — try first backend
		if len(allBackends) > 0 {
			allBackends[0].Proxy.ServeHTTP(w, r)
		}
	})

	// /_api/chats/<id> and /_api/runs/<id> — try each backend
	mux.HandleFunc("/_api/chats/", func(w http.ResponseWriter, r *http.Request) {
		proxyToFirstResponder(w, r, allBackends)
	})
	mux.HandleFunc("/_api/runs/", func(w http.ResponseWriter, r *http.Request) {
		proxyToFirstResponder(w, r, allBackends)
	})
	mux.HandleFunc("/_api/runs", func(w http.ResponseWriter, r *http.Request) {
		wfParam := r.URL.Query().Get("workflow")
		if name, rest, ok := parseBackendPrefix(wfParam); ok {
			if b, exists := backendMap[name]; exists {
				q := r.URL.Query()
				q.Set("workflow", rest)
				r.URL.RawQuery = q.Encode()
				b.Proxy.ServeHTTP(w, r)
				return
			}
		}
		if len(allBackends) > 0 {
			allBackends[0].Proxy.ServeHTTP(w, r)
		}
	})

	// Catch-all: SPA pages or proxy
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/html") {
			meta := buildPageMetaMulti(allBackends, r.URL.Path)
			page := strings.Replace(loaderHTML, "{{META}}", meta, 1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(page))
			return
		}
		// Non-HTML requests to root — 404
		http.NotFound(w, r)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Fprintf(os.Stderr, "\n  haira webui (multi-backend)\n")
	fmt.Fprintf(os.Stderr, "  Local: http://localhost:%d/\n", port)
	for _, b := range allBackends {
		fmt.Fprintf(os.Stderr, "  Backend: %s → /_b/%s/\n", b.URL.String(), b.Name)
	}
	fmt.Fprintf(os.Stderr, "\n")

	return http.ListenAndServe(addr, mux)
}

// ---------------------------------------------------------------------------
// Shared handlers
// ---------------------------------------------------------------------------

func proxyErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	json.NewEncoder(w).Encode(map[string]string{
		"error": fmt.Sprintf("Backend unreachable: %v", err),
	})
}

func serveUIJS(uiJS []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(uiJS)
	}
}

func serveUIConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"src": "/_ui/assets/haira-ui.js",
	})
}

// ---------------------------------------------------------------------------
// WebSocket proxy (TCP-level, same pattern as orchestrator/proxy.go)
// ---------------------------------------------------------------------------

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, backendHost, path string) {
	backendConn, err := net.DialTimeout("tcp", backendHost, 5*time.Second)
	if err != nil {
		http.Error(w, `{"error":"backend unavailable"}`, http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, `{"error":"websocket not supported"}`, http.StatusInternalServerError)
		return
	}

	clientConn, clientBuf, err := hj.Hijack()
	if err != nil {
		http.Error(w, `{"error":"hijack failed"}`, http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Forward the original request with rewritten path
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

// ---------------------------------------------------------------------------
// Backend name derivation
// ---------------------------------------------------------------------------

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func deriveBackendName(workflows []map[string]any, u *url.URL) string {
	if len(workflows) > 0 {
		if name, ok := workflows[0]["name"].(string); ok && name != "" {
			s := slugify(name)
			if s != "" {
				return s
			}
		}
	}
	// Fallback: host-port
	host := u.Hostname()
	port := u.Port()
	if port != "" {
		return slugify(host + "-" + port)
	}
	return slugify(host)
}

func uniqueName(name string, used map[string]bool) string {
	if !used[name] {
		return name
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", name, i)
		if !used[candidate] {
			return candidate
		}
	}
}

// ---------------------------------------------------------------------------
// Workflow fetching and namespacing
// ---------------------------------------------------------------------------

func fetchWorkflows(backendURL string) []map[string]any {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(backendURL + "/_api/workflows")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var workflows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&workflows); err != nil {
		return nil
	}
	return workflows
}

func namespacedWorkflowItem(wf map[string]any, backendName string) map[string]any {
	origPath, _ := wf["path"].(string)
	item := map[string]any{
		"name":    wf["name"],
		"path":    "/_b/" + backendName + origPath,
		"method":  wf["method"],
		"backend": backendName,
	}

	isStream, _ := wf["is_stream"].(bool)
	if isStream {
		item["is_stream"] = true
	}

	for _, key := range []string{"title", "description", "params", "suggestions", "chat_param"} {
		if v, ok := wf[key]; ok && v != nil {
			item[key] = v
		}
	}

	return item
}

// parseBackendPrefix extracts the backend name from a /_b/<name>/... path.
// Returns (name, rest, true) or ("", "", false).
func parseBackendPrefix(path string) (string, string, bool) {
	if !strings.HasPrefix(path, "/_b/") {
		return "", "", false
	}
	trimmed := strings.TrimPrefix(path, "/_b/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", "", false
	}
	rest := "/"
	if len(parts) > 1 {
		rest = "/" + parts[1]
	}
	return parts[0], rest, true
}

// ---------------------------------------------------------------------------
// Session/run routing: try each backend until one responds with 200
// ---------------------------------------------------------------------------

func proxyToFirstResponder(w http.ResponseWriter, r *http.Request, backends []*Backend) {
	client := &http.Client{Timeout: 5 * time.Second}
	for _, b := range backends {
		checkURL := b.URL.String() + r.URL.Path
		if r.URL.RawQuery != "" {
			checkURL += "?" + r.URL.RawQuery
		}
		resp, err := client.Get(checkURL)
		if err != nil {
			continue
		}
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			// This backend has it — proxy the real request
			b.Proxy.ServeHTTP(w, r)
			return
		}
		resp.Body.Close()
	}
	// None found
	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}

// ---------------------------------------------------------------------------
// Background refresh
// ---------------------------------------------------------------------------

func refreshLoop(backends []*Backend, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		for _, b := range backends {
			wfs := fetchWorkflows(b.URL.String())
			if wfs != nil {
				b.mu.Lock()
				b.Workflows = wfs
				b.mu.Unlock()
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Page metadata — single backend (original logic)
// ---------------------------------------------------------------------------

func buildPageMetaSingle(backend, path string) string {
	workflows := fetchWorkflows(backend)
	if workflows == nil {
		workflows = []map[string]any{}
	}
	return buildMetaFromWorkflows(workflows, path, "")
}

// ---------------------------------------------------------------------------
// Page metadata — multi backend
// ---------------------------------------------------------------------------

func buildPageMetaMulti(backends []*Backend, path string) string {
	// / → index page with merged workflows
	if path == "/" {
		var items []map[string]any
		for _, b := range backends {
			b.mu.RLock()
			wfs := b.Workflows
			b.mu.RUnlock()
			for _, wf := range wfs {
				item := buildWorkflowListItem(wf, b.Name)
				items = append(items, item)
			}
		}
		meta := map[string]any{
			"mode":      "index",
			"workflows": items,
		}
		out, _ := json.Marshal(meta)
		return string(out)
	}

	// /workbench/_b/<name>/<path> → find matching backend and workflow
	wfPath := strings.TrimPrefix(path, "/workbench")
	if name, rest, ok := parseBackendPrefix(wfPath); ok {
		for _, b := range backends {
			if b.Name != name {
				continue
			}
			b.mu.RLock()
			wfs := b.Workflows
			b.mu.RUnlock()
			for _, wf := range wfs {
				if wf["path"] == rest {
					return buildWorkflowMeta(wf, b.Name)
				}
			}
		}
	}

	return `{"mode":"index","workflows":[]}`
}

// buildWorkflowListItem creates an item for the index page workflow list.
func buildWorkflowListItem(wf map[string]any, backendName string) map[string]any {
	origPath, _ := wf["path"].(string)
	namespacedPath := "/_b/" + backendName + origPath

	item := map[string]any{
		"name":    wf["name"],
		"path":    namespacedPath,
		"method":  wf["method"],
		"backend": backendName,
		"arpUrl":  "/_b/" + backendName + "/_arp/v1",
	}

	isStream, _ := wf["is_stream"].(bool)
	if isStream {
		item["uiType"] = "Chat"
	} else {
		item["uiType"] = "Form"
	}

	for _, key := range []string{"title", "description", "params", "suggestions", "chat_param"} {
		if v, ok := wf[key]; ok && v != nil {
			if key == "chat_param" {
				item["chatParam"] = v
			} else {
				item[key] = v
			}
		}
	}

	if params, ok := wf["params"].([]any); ok {
		for _, p := range params {
			if pm, ok := p.(map[string]any); ok {
				if pm["type"] == "file" {
					item["hasFile"] = true
					item["fileParam"] = pm["name"]
					break
				}
			}
		}
	}

	return item
}

// buildWorkflowMeta creates per-workflow page metadata.
func buildWorkflowMeta(wf map[string]any, backendName string) string {
	origPath, _ := wf["path"].(string)
	namespacedPath := "/_b/" + backendName + origPath

	isStream, _ := wf["is_stream"].(bool)
	mode := "form"
	if isStream {
		mode = "chat"
	}

	meta := map[string]any{
		"mode":    mode,
		"name":    wf["name"],
		"method":  wf["method"],
		"path":    namespacedPath,
		"params":  []any{},
		"backend": backendName,
		"arpUrl":  "/_b/" + backendName + "/_arp/v1",
	}

	for _, key := range []string{"title", "description", "params", "suggestions"} {
		if v, ok := wf[key]; ok && v != nil {
			meta[key] = v
		}
	}
	if cp, ok := wf["chat_param"].(string); ok && cp != "" {
		meta["chatParam"] = cp
	}
	if params, ok := wf["params"].([]any); ok {
		for _, p := range params {
			if pm, ok := p.(map[string]any); ok {
				if pm["type"] == "file" {
					meta["hasFile"] = true
					meta["fileParam"] = pm["name"]
					break
				}
			}
		}
	}

	out, _ := json.Marshal(meta)
	return string(out)
}

// buildMetaFromWorkflows is the original single-backend metadata builder.
func buildMetaFromWorkflows(workflows []map[string]any, path, backendName string) string {
	if path == "/" {
		var items []map[string]any
		for _, wf := range workflows {
			item := map[string]any{
				"name":   wf["name"],
				"path":   wf["path"],
				"method": wf["method"],
			}
			isStream, _ := wf["is_stream"].(bool)
			if isStream {
				item["uiType"] = "Chat"
			} else {
				item["uiType"] = "Form"
			}
			for _, key := range []string{"title", "description", "params", "suggestions", "chat_param"} {
				if v, ok := wf[key]; ok && v != nil {
					if key == "chat_param" {
						item["chatParam"] = v
					} else {
						item[key] = v
					}
				}
			}
			if params, ok := wf["params"].([]any); ok {
				for _, p := range params {
					if pm, ok := p.(map[string]any); ok {
						if pm["type"] == "file" {
							item["hasFile"] = true
							item["fileParam"] = pm["name"]
							break
						}
					}
				}
			}
			items = append(items, item)
		}
		meta := map[string]any{
			"mode":      "index",
			"workflows": items,
		}
		b, _ := json.Marshal(meta)
		return string(b)
	}

	wfPath := strings.TrimPrefix(path, "/workbench")
	for _, wf := range workflows {
		if wf["path"] == wfPath {
			isStream, _ := wf["is_stream"].(bool)
			mode := "form"
			if isStream {
				mode = "chat"
			}
			meta := map[string]any{
				"mode":   mode,
				"name":   wf["name"],
				"method": wf["method"],
				"path":   wf["path"],
				"params": []any{},
			}
			for _, key := range []string{"title", "description", "params", "suggestions"} {
				if v, ok := wf[key]; ok && v != nil {
					meta[key] = v
				}
			}
			if cp, ok := wf["chat_param"].(string); ok && cp != "" {
				meta["chatParam"] = cp
			}
			if params, ok := wf["params"].([]any); ok {
				for _, p := range params {
					if pm, ok := p.(map[string]any); ok {
						if pm["type"] == "file" {
							meta["hasFile"] = true
							meta["fileParam"] = pm["name"]
							break
						}
					}
				}
			}
			b, _ := json.Marshal(meta)
			return string(b)
		}
	}

	return `{"mode":"index","workflows":[]}`
}

// StartAssetServer starts a lightweight HTTP server in the background that serves
// only the UI JS bundle. It picks a random available port and sets HAIRA_UI_URL
// so that compiled programs can load the UI from it.
//
// Returns a cleanup function to stop the server, or nil if the UI bundle is not available.
func StartAssetServer() func() {
	cliFiles := runtime.CLIFiles()
	uiJS, ok := cliFiles["haira-ui.js"]
	if !ok {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/haira-ui.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(uiJS)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil
	}

	port := ln.Addr().(*net.TCPAddr).Port
	uiURL := fmt.Sprintf("http://127.0.0.1:%d/haira-ui.js", port)
	os.Setenv("HAIRA_UI_URL", uiURL)

	go http.Serve(ln, mux)

	return func() {
		ln.Close()
		os.Unsetenv("HAIRA_UI_URL")
	}
}

// Package webui implements the `haira webui` command.
// It serves the Haira UI SDK (embedded in the compiler binary) and proxies
// all other requests to a running Haira backend (which is a pure ARP server
// with no UI of its own).
//
// This allows users to get the full UI experience without installing anything
// extra — the haira CLI is all they need.
package webui

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/haira-lang/haira/internal/runtime"
)

// Run starts the webui server.
func Run(backend string, port int) error {
	// Load UI bundle + loader from embedded CLI files
	cliFiles := runtime.CLIFiles()
	uiJS, ok := cliFiles["haira-ui.js"]
	if !ok {
		return fmt.Errorf("UI bundle not found in compiler binary (rebuild with `make build`)")
	}
	loaderHTML := string(cliFiles["loader.html"])
	if loaderHTML == "" {
		return fmt.Errorf("loader.html not found in compiler binary (rebuild with `make build`)")
	}

	// Normalize backend URL
	if !strings.HasPrefix(backend, "http") {
		backend = "http://" + backend
	}
	backendURL, err := url.Parse(backend)
	if err != nil {
		return fmt.Errorf("invalid backend URL %q: %w", backend, err)
	}

	// Reverse proxy to backend — default for non-UI routes
	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Backend unreachable: %v", err),
		})
	}

	mux := http.NewServeMux()

	// Serve the UI JS bundle locally
	mux.HandleFunc("/_ui/assets/haira-ui.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(uiJS)
	})

	// UI config — point to our local bundle
	mux.HandleFunc("/_ui/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"src": "/_ui/assets/haira-ui.js",
		})
	})

	// Proxy API/ARP/observe routes to backend
	for _, prefix := range []string{"/_api/", "/_arp/", "/_observe/"} {
		p := prefix
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	}

	// Catch-all: serve loader.html for browser navigation, proxy everything else
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only serve the SPA page for browser navigation (GET requests that accept text/html).
		// All other requests (non-GET, fetch/XHR, SSE) are proxied to the backend.
		if r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/html") {
			meta := buildPageMeta(backend, r.URL.Path)
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

// buildPageMeta constructs the page metadata JSON for a given path.
// It fetches the workflow list from the backend's /_api/workflows endpoint
// and builds the appropriate metadata (index page or per-workflow page).
func buildPageMeta(backend, path string) string {
	// Fetch workflow list from backend API
	resp, err := http.Get(backend + "/_api/workflows")
	if err != nil {
		return `{"mode":"index","workflows":[]}`
	}
	defer resp.Body.Close()

	var workflows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&workflows); err != nil {
		return `{"mode":"index","workflows":[]}`
	}

	// / → index page with all workflows
	if path == "/" {
		// Build workflow items for the index page
		var items []map[string]any
		for _, wf := range workflows {
			item := map[string]any{
				"name":   wf["name"],
				"path":   wf["path"],
				"method": wf["method"],
			}
			// Determine UI type
			isStream, _ := wf["is_stream"].(bool)
			if isStream {
				item["uiType"] = "Chat"
			} else {
				item["uiType"] = "Form"
			}
			// Copy optional fields
			for _, key := range []string{"title", "description", "params", "suggestions", "chat_param"} {
				if v, ok := wf[key]; ok && v != nil {
					if key == "chat_param" {
						item["chatParam"] = v
					} else {
						item[key] = v
					}
				}
			}
			// Check for file params
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

	// /workbench/<path> → find matching workflow
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
			// Copy optional fields
			for _, key := range []string{"title", "description", "params", "suggestions"} {
				if v, ok := wf[key]; ok && v != nil {
					meta[key] = v
				}
			}
			if cp, ok := wf["chat_param"].(string); ok && cp != "" {
				meta["chatParam"] = cp
			}
			// Check for file params
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

	// Not found — return index
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

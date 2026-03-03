package haira

import (
	_ "embed"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

//go:embed ui/loader.html
var loaderHTML string

//go:embed ui/haira-ui.js
var uiJS []byte

// registerUIRoutes sets up routes to serve the embedded web UI.
// The UI is enabled by default. Set HAIRA_NO_UI=1 to disable.
// Set HAIRA_UI_URL to load the JS bundle from an external source (CDN, dev server).
func (s *Server) registerUIRoutes() {
	if os.Getenv("HAIRA_NO_UI") != "" {
		return
	}

	// Compute a short content hash for cache-busting the JS bundle URL.
	sum := sha256.Sum256(uiJS)
	jsVersion := hex.EncodeToString(sum[:4]) // 8 hex chars, e.g. "a1b2c3d4"

	// Serve the JS bundle
	s.mux.HandleFunc("/_ui/assets/haira-ui.js", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		rw.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		rw.Write(uiJS)
	})

	// UI config — tells the loader where to find the JS bundle.
	// HAIRA_UI_URL overrides to an external source (CDN, dev server).
	s.mux.HandleFunc("/_ui/config", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Cache-Control", "no-cache, must-revalidate")
		src := "/_ui/assets/haira-ui.js?v=" + jsVersion
		if extURL := os.Getenv("HAIRA_UI_URL"); extURL != "" {
			src = extURL
		}
		json.NewEncoder(rw).Encode(map[string]string{"src": src})
	})

	// SPA shell — serve loader.html for browser navigation requests.
	// More specific routes (workflows, /_api/, /_arp/, /_observe/) take precedence.
	s.mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/html") {
			meta := s.buildPageMeta(r.URL.Path)
			page := strings.Replace(loaderHTML, "{{META}}", meta, 1)
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.Write([]byte(page))
			return
		}
		http.NotFound(rw, r)
	})
}

// buildPageMeta generates the JSON metadata that the SPA shell uses for routing.
func (s *Server) buildPageMeta(path string) string {
	// Root or index pages → list all workflows
	if path == "/" {
		return s.buildIndexMeta()
	}

	// Workbench page → find matching workflow
	wfPath := strings.TrimPrefix(path, "/workbench")
	for _, wf := range s.workflows {
		if wf.Path == wfPath {
			return s.buildWorkflowMeta(wf)
		}
	}

	// Fallback to index
	return s.buildIndexMeta()
}

// buildIndexMeta creates metadata for the workflow index page.
func (s *Server) buildIndexMeta() string {
	var items []map[string]any
	for _, wf := range s.workflows {
		items = append(items, s.workflowListItem(wf))
	}
	meta := map[string]any{
		"mode":      "index",
		"workflows": items,
	}
	out, _ := json.Marshal(meta)
	return string(out)
}

// buildWorkflowMeta creates metadata for a single workflow page.
func (s *Server) buildWorkflowMeta(wf *WorkflowDef) string {
	isStream := wf.IsStream || wf.StreamHandler != nil
	mode := "form"
	if wf.UIMode != "" {
		mode = wf.UIMode
	} else if isStream {
		mode = "chat"
	}

	meta := map[string]any{
		"mode":   mode,
		"name":   wf.Name,
		"method": wf.Method,
		"path":   wf.Path,
	}

	if wf.UITitle != "" {
		meta["title"] = wf.UITitle
	}
	if wf.UIDescription != "" {
		meta["description"] = wf.UIDescription
	}
	if wf.UIAccent != "" {
		meta["accent"] = wf.UIAccent
	}
	if wf.UITheme != "" {
		meta["theme"] = wf.UITheme
	}
	if wf.UIAvatar != "" {
		meta["avatar"] = wf.UIAvatar
	}
	if wf.UILogo != "" {
		meta["logo"] = wf.UILogo
	}
	if len(wf.Suggestions) > 0 {
		meta["suggestions"] = wf.Suggestions
	}
	if len(wf.Params) > 0 {
		meta["params"] = wf.Params
	}
	if isStream {
		meta["chatParam"] = findChatParam(wf.Params)
	}
	if len(wf.Steps) > 0 {
		meta["steps"] = wf.Steps
	}
	for _, p := range wf.Params {
		if p.Type == "file" {
			meta["hasFile"] = true
			meta["fileParam"] = p.Name
			break
		}
	}

	// Include full workflow list so settings/nav work on single-workflow pages
	var items []map[string]any
	for _, w := range s.workflows {
		items = append(items, s.workflowListItem(w))
	}
	meta["workflows"] = items

	out, _ := json.Marshal(meta)
	return string(out)
}

// workflowListItem creates a workflow entry for the index page.
func (s *Server) workflowListItem(wf *WorkflowDef) map[string]any {
	isStream := wf.IsStream || wf.StreamHandler != nil

	item := map[string]any{
		"name":   wf.Name,
		"path":   wf.Path,
		"method": wf.Method,
	}

	if isStream {
		item["uiType"] = "Chat"
		item["chatParam"] = findChatParam(wf.Params)
	} else {
		item["uiType"] = "Form"
	}

	if wf.UITitle != "" {
		item["title"] = wf.UITitle
	}
	if wf.UIDescription != "" {
		item["description"] = wf.UIDescription
	}
	if len(wf.Params) > 0 {
		item["params"] = wf.Params
	}
	if len(wf.Steps) > 0 {
		item["steps"] = wf.Steps
	}
	if len(wf.Suggestions) > 0 {
		item["suggestions"] = wf.Suggestions
	}
	for _, p := range wf.Params {
		if p.Type == "file" {
			item["hasFile"] = true
			item["fileParam"] = p.Name
			break
		}
	}

	return item
}

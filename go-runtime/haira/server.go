package haira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Server is an HTTP server that routes requests to workflows.
type Server struct {
	mux       *http.ServeMux
	workflows []*WorkflowDef
}

// NewServer creates a new server with the given workflow definitions.
func NewServer(workflows []*WorkflowDef) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		workflows: workflows,
	}

	// Register auto-UI routes (unless disabled)
	if os.Getenv("HAIRA_DISABLE_UI") == "" {
		s.registerUIRoutes()
	}

	for _, w := range workflows {
		wf := w // capture for closure
		s.mux.HandleFunc(wf.Path, func(rw http.ResponseWriter, r *http.Request) {
			if r.Method != wf.Method {
				http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			params := make(map[string]any)

			if r.Method == "GET" || r.Method == "DELETE" {
				// GET/DELETE: params from query string
				for k, v := range r.URL.Query() {
					params[k] = v[0]
				}
			} else {
				// POST/PUT: check for file params → multipart, else JSON body
				hasFileParams := false
				for _, p := range wf.Params {
					if p.Type == "file" {
						hasFileParams = true
						break
					}
				}

				if hasFileParams && strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
					r.ParseMultipartForm(32 << 20) // 32MB max
					var tempFiles []string
					for _, p := range wf.Params {
						if p.Type == "file" {
							file, header, err := r.FormFile(p.Name)
							if err == nil {
								tmpPath, saveErr := saveTempFile(file, header.Filename)
								file.Close()
								if saveErr == nil {
									params[p.Name] = tmpPath
									tempFiles = append(tempFiles, tmpPath)
								}
							}
						} else {
							if v := r.FormValue(p.Name); v != "" {
								params[p.Name] = v
							}
						}
					}
					// Cleanup temp files after handler returns
					defer func() {
						for _, f := range tempFiles {
							os.Remove(f)
						}
					}()
				} else {
					if r.Body != nil {
						defer r.Body.Close()
						json.NewDecoder(r.Body).Decode(&params)
					}
					if params == nil {
						params = make(map[string]any)
					}
				}
				// Include query parameters as fallback
				for k, v := range r.URL.Query() {
					if _, exists := params[k]; !exists {
						params[k] = v[0]
					}
				}
			}

			// Check if client wants streaming (Accept: text/event-stream)
			if wf.StreamHandler != nil && r.Header.Get("Accept") == "text/event-stream" {
				s.handleSSE(rw, r, wf, params)
				return
			}

			result, err := wf.Handler(params)
			if err != nil {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(map[string]any{"error": err.Error()})
				return
			}

			rw.Header().Set("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(result)
		})
	}

	return s
}

func (s *Server) handleSSE(rw http.ResponseWriter, r *http.Request, wf *WorkflowDef, params map[string]any) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	ch, err := wf.StreamHandler(params)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(rw).Encode(map[string]any{"error": err.Error()})
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	for chunk := range ch {
		if chunk.Done {
			fmt.Fprintf(rw, "data: [DONE]\n\n")
			flusher.Flush()
			break
		}
		data, _ := json.Marshal(map[string]string{"delta": chunk.Delta})
		fmt.Fprintf(rw, "data: %s\n\n", data)
		flusher.Flush()
	}
}

// registerUIRoutes sets up /_ui/ index and per-workflow UI pages.
func (s *Server) registerUIRoutes() {
	s.mux.HandleFunc("/_ui/", s.handleUIIndex)

	for _, w := range s.workflows {
		wf := w
		uiPath := "/_ui" + wf.Path
		s.mux.HandleFunc(uiPath, func(rw http.ResponseWriter, r *http.Request) {
			if wf.IsStream && (wf.ChatEnabled == nil || *wf.ChatEnabled) {
				s.serveChatUI(rw, wf)
			} else {
				s.serveFormUI(rw, wf)
			}
		})
	}
}

func (s *Server) handleUIIndex(rw http.ResponseWriter, r *http.Request) {
	// Only serve exact /_ui/ path as index
	if r.URL.Path != "/_ui/" {
		http.NotFound(rw, r)
		return
	}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(rw, `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Haira Workflows</title>`)
	fmt.Fprint(rw, `<style>*{box-sizing:border-box;margin:0;padding:0}body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f5f5f5;padding:2rem 1rem;display:flex;justify-content:center}`)
	fmt.Fprint(rw, `.container{max-width:600px;width:100%}h1{margin-bottom:1.5rem;font-size:1.5rem}`)
	fmt.Fprint(rw, `.wf{background:#fff;border-radius:8px;box-shadow:0 1px 3px rgba(0,0,0,.1);padding:1rem 1.25rem;margin-bottom:.75rem;display:flex;align-items:center;justify-content:space-between;text-decoration:none;color:#1a1a1a;transition:box-shadow .15s}`)
	fmt.Fprint(rw, `.wf:hover{box-shadow:0 2px 8px rgba(0,0,0,.15)}.wf-name{font-weight:600}.wf-path{font-family:monospace;font-size:.85rem;color:#555}`)
	fmt.Fprint(rw, `.badge{font-size:.7rem;font-weight:700;padding:.15rem .5rem;border-radius:3px;color:#fff;margin-right:.5rem}`)
	fmt.Fprint(rw, `.badge-POST{background:#49cc90}.badge-GET{background:#61affe}.badge-PUT{background:#fca130}.badge-DELETE{background:#f93e3e}`)
	fmt.Fprint(rw, `</style></head><body><div class="container"><h1>Haira Workflows</h1>`)
	for _, wf := range s.workflows {
		uiPath := "/_ui" + wf.Path
		uiType := "Form"
		if wf.IsStream {
			uiType = "Chat"
		}
		fmt.Fprintf(rw, `<a class="wf" href="%s"><div><span class="badge badge-%s">%s</span><span class="wf-name">%s</span><br><span class="wf-path">%s</span></div><div style="color:#999;font-size:.8rem">%s</div></a>`,
			uiPath, wf.Method, wf.Method, wf.Name, wf.Path, uiType)
	}
	fmt.Fprint(rw, `</div></body></html>`)
}

// Listen starts the HTTP server on the given port.
func (s *Server) Listen(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, s.mux)
}

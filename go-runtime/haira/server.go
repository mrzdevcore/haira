package haira

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

//go:embed ui/index.html
var indexHTML string

//go:embed ui/dist/haira-ui.js
var uiJS string

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
			if r.Header.Get("Accept") == "text/event-stream" {
				if wf.StreamHandler != nil {
					s.handleSSE(rw, r, wf, params)
					return
				}
				if len(wf.Steps) > 0 {
					s.handleStepSSE(rw, r, wf, params)
					return
				}
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
			// Error with delta — send as error event before [DONE]
			if chunk.Delta != "" && strings.HasPrefix(chunk.Delta, "error:") {
				errData, _ := json.Marshal(map[string]string{"error": strings.TrimPrefix(chunk.Delta, "error: ")})
				fmt.Fprintf(rw, "event: error\ndata: %s\n\n", errData)
				flusher.Flush()
			}
			fmt.Fprintf(rw, "data: [DONE]\n\n")
			flusher.Flush()
			break
		}

		switch chunk.Type {
		case "tool_start":
			data, _ := json.Marshal(map[string]any{
				"tool": chunk.ToolName,
				"args": chunk.ToolArgs,
			})
			fmt.Fprintf(rw, "event: tool_start\ndata: %s\n\n", data)
		case "tool_render":
			data, _ := json.Marshal(map[string]any{
				"tool":      chunk.ToolName,
				"component": chunk.RenderComponent,
				"props":     json.RawMessage(chunk.RenderProps),
			})
			fmt.Fprintf(rw, "event: tool_render\ndata: %s\n\n", data)
		case "tool_end":
			data, _ := json.Marshal(map[string]any{
				"tool": chunk.ToolName,
				"ok":   chunk.ToolOK,
			})
			fmt.Fprintf(rw, "event: tool_end\ndata: %s\n\n", data)
		default:
			// Normal text delta (Type == "")
			if chunk.Delta != "" {
				data, _ := json.Marshal(map[string]string{"delta": chunk.Delta})
				fmt.Fprintf(rw, "data: %s\n\n", data)
			}
		}
		flusher.Flush()
	}
}

// handleStepSSE runs a sync workflow handler in a goroutine, streaming step events
// via SSE, then sends the final result.
func (s *Server) handleStepSSE(rw http.ResponseWriter, r *http.Request, wf *WorkflowDef, params map[string]any) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	stepCh := make(chan StepEvent, 16)
	type handlerResult struct {
		data any
		err  error
	}
	resultCh := make(chan handlerResult, 1)

	go func() {
		SetStepNotifier(stepCh)
		defer ClearStepNotifier()
		defer close(stepCh)

		result, err := wf.Handler(params)
		resultCh <- handlerResult{data: result, err: err}
	}()

	// Stream step events until the handler completes
	for event := range stepCh {
		data, _ := json.Marshal(event)
		fmt.Fprintf(rw, "event: step\ndata: %s\n\n", data)
		flusher.Flush()
	}

	// Send final result
	res := <-resultCh
	if res.err != nil {
		errData, _ := json.Marshal(map[string]string{"error": res.err.Error()})
		fmt.Fprintf(rw, "event: error\ndata: %s\n\n", errData)
	} else {
		resultData, _ := json.Marshal(res.data)
		fmt.Fprintf(rw, "event: result\ndata: %s\n\n", resultData)
	}
	fmt.Fprintf(rw, "data: [DONE]\n\n")
	flusher.Flush()
}

// registerUIRoutes sets up /_ui/ index, per-workflow UI pages, and JS assets.
func (s *Server) registerUIRoutes() {
	s.mux.HandleFunc("/_ui/assets/haira-ui.js", s.serveUIJS)
	s.mux.HandleFunc("/_ui/", s.handleUIIndex)

	for _, w := range s.workflows {
		wf := w
		uiPath := "/_ui" + wf.Path
		s.mux.HandleFunc(uiPath, func(rw http.ResponseWriter, r *http.Request) {
			switch wf.UIMode {
			case "chat":
				s.serveChatUI(rw, wf)
			case "form":
				s.serveFormUI(rw, wf)
			default:
				// Auto-detect: stream workflows default to chat, sync to form
				if wf.IsStream {
					s.serveChatUI(rw, wf)
				} else {
					s.serveFormUI(rw, wf)
				}
			}
		})
	}
}

func (s *Server) serveUIJS(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	rw.Header().Set("Cache-Control", "public, max-age=3600")
	rw.Write([]byte(uiJS))
}

func (s *Server) handleUIIndex(rw http.ResponseWriter, r *http.Request) {
	// Only serve exact /_ui/ path as index
	if r.URL.Path != "/_ui/" {
		http.NotFound(rw, r)
		return
	}

	type wfItem struct {
		Name   string `json:"name"`
		Path   string `json:"path"`
		Method string `json:"method"`
		UIType string `json:"uiType"`
		Title  string `json:"title"`
	}

	var items []wfItem
	for _, wf := range s.workflows {
		var uiType string
		switch wf.UIMode {
		case "chat":
			uiType = "Chat"
		case "form":
			uiType = "Form"
		default:
			if wf.IsStream {
				uiType = "Chat"
			} else {
				uiType = "Form"
			}
		}
		items = append(items, wfItem{
			Name:   wf.Name,
			Path:   wf.Path,
			Method: wf.Method,
			UIType: uiType,
			Title:  wf.UITitle,
		})
	}

	meta := map[string]any{
		"mode":      "index",
		"workflows": items,
	}
	metaJSON, _ := json.Marshal(meta)
	html := strings.Replace(indexHTML, "{{META}}", string(metaJSON), 1)
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Write([]byte(html))
}

// Listen starts the HTTP server on the given port.
func (s *Server) Listen(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, s.mux)
}

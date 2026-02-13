package haira

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	for _, w := range workflows {
		wf := w // capture for closure
		s.mux.HandleFunc(wf.Path, func(rw http.ResponseWriter, r *http.Request) {
			if r.Method != wf.Method {
				http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Parse request body as JSON
			var params map[string]any
			if r.Body != nil {
				defer r.Body.Close()
				json.NewDecoder(r.Body).Decode(&params)
			}
			if params == nil {
				params = make(map[string]any)
			}

			// Include query parameters
			for k, v := range r.URL.Query() {
				if _, exists := params[k]; !exists {
					params[k] = v[0]
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

// Listen starts the HTTP server on the given port.
func (s *Server) Listen(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, s.mux)
}

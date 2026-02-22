package haira

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed ui/form.html
var formHTML string

func (s *Server) serveFormUI(rw http.ResponseWriter, r *http.Request, wf *WorkflowDef) {
	hasFile := false
	for _, p := range wf.Params {
		if p.Type == "file" {
			hasFile = true
			break
		}
	}
	meta := map[string]any{
		"mode":        "form",
		"name":        wf.Name,
		"method":      wf.Method,
		"path":        wf.Path,
		"params":      wf.Params,
		"steps":       wf.Steps,
		"title":       wf.UITitle,
		"description": wf.UIDescription,
		"hasFile":     hasFile,
		"accent":      wf.UIAccent,
		"logo":        wf.UILogo,
		"theme":       wf.UITheme,
	}
	metaJSON, _ := json.Marshal(meta)
	html := strings.Replace(formHTML, "{{META}}", string(metaJSON), 1)
	serveHTML(rw, r, html)
}

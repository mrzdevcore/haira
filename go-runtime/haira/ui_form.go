package haira

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed ui/form.html
var formHTML string

var formTmpl = template.Must(template.New("form").Parse(formHTML))

func (s *Server) serveFormUI(rw http.ResponseWriter, wf *WorkflowDef) {
	hasFile := false
	for _, p := range wf.Params {
		if p.Type == "file" {
			hasFile = true
			break
		}
	}
	data := map[string]any{
		"Name":        wf.Name,
		"Method":      wf.Method,
		"Path":        wf.Path,
		"Params":      wf.Params,
		"Title":       wf.UITitle,
		"Description": wf.UIDescription,
		"HasFile":     hasFile,
	}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	formTmpl.Execute(rw, data)
}

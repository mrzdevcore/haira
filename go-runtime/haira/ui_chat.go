package haira

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed ui/chat.html
var chatHTML string

var chatTmpl = template.Must(template.New("chat").Parse(chatHTML))

func (s *Server) serveChatUI(rw http.ResponseWriter, wf *WorkflowDef) {
	chatParam := findChatParam(wf.Params)
	settingsParams := filterSettingsParams(wf.Params, chatParam)
	data := map[string]any{
		"Name":           wf.Name,
		"Path":           wf.Path,
		"ChatParam":      chatParam,
		"SettingsParams": settingsParams,
		"Title":          wf.UITitle,
	}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	chatTmpl.Execute(rw, data)
}

// findChatParam finds the primary chat input parameter name.
func findChatParam(params []WorkflowParam) string {
	for _, p := range params {
		if p.Type == "string" && (p.Name == "message" || p.Name == "prompt" || p.Name == "input") {
			return p.Name
		}
	}
	// Fallback: first string param
	for _, p := range params {
		if p.Type == "string" {
			return p.Name
		}
	}
	return "message"
}

// filterSettingsParams returns all params except the chat input param.
func filterSettingsParams(params []WorkflowParam, chatParam string) []WorkflowParam {
	var result []WorkflowParam
	for _, p := range params {
		if p.Name != chatParam {
			result = append(result, p)
		}
	}
	return result
}

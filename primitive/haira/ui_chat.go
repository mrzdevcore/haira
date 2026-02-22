package haira

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed ui/chat.html
var chatHTML string

func (s *Server) serveChatUI(rw http.ResponseWriter, r *http.Request, wf *WorkflowDef) {
	chatParam := findChatParam(wf.Params)
	settingsParams := filterSettingsParams(wf.Params, chatParam)
	hasFile := false
	fileParam := ""
	for _, p := range wf.Params {
		if p.Type == "file" {
			hasFile = true
			fileParam = p.Name
			break
		}
	}
	meta := map[string]any{
		"mode":           "chat",
		"name":           wf.Name,
		"method":         wf.Method,
		"path":           wf.Path,
		"params":         wf.Params,
		"steps":          wf.Steps,
		"title":          wf.UITitle,
		"description":    wf.UIDescription,
		"chatParam":      chatParam,
		"settingsParams": settingsParams,
		"hasFile":        hasFile,
		"fileParam":      fileParam,
		"suggestions":    wf.Suggestions,
		"accent":         wf.UIAccent,
		"logo":           wf.UILogo,
		"theme":          wf.UITheme,
		"avatar":         wf.UIAvatar,
	}
	metaJSON, _ := json.Marshal(meta)
	html := strings.Replace(chatHTML, "{{META}}", string(metaJSON), 1)
	serveHTML(rw, r, html)
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

// filterSettingsParams returns all params except the chat input param and session_id
// (session_id is managed internally by the chat component).
func filterSettingsParams(params []WorkflowParam, chatParam string) []WorkflowParam {
	var result []WorkflowParam
	for _, p := range params {
		if p.Name != chatParam && p.Name != "session_id" {
			result = append(result, p)
		}
	}
	return result
}

package haira

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

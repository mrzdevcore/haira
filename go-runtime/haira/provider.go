package haira

// Provider holds LLM provider configuration.
// Works with OpenAI-compatible APIs including Azure OpenAI.
type Provider struct {
	Name        string
	ApiKey      string
	Endpoint    string
	Model       string
	ApiVersion  string
	Backend     string  // Informational: "ollama", "llama.cpp", etc.
	Host        string  // Local backend host — resolved to Endpoint at client creation
	Temperature float64 // Default temperature for agents using this provider
	MaxTokens   int     // Default max tokens for agents using this provider
}

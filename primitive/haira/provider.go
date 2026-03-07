package haira

// Provider holds LLM provider configuration.
// Works with OpenAI-compatible APIs including Azure OpenAI, Cloudflare Workers AI,
// and many others via backend-aware endpoint resolution.
type Provider struct {
	Name            string
	ApiKey          string
	Endpoint        string  // Explicit API endpoint (overrides backend resolution)
	Model           string
	ApiVersion      string  // API version (Azure OpenAI)
	Backend         string  // Backend identifier — drives automatic endpoint resolution
	Host            string  // Backend host address (local providers like Ollama)
	AccountID       string  // Account identifier (Cloudflare Workers AI, etc.)
	Temperature     float64 // Default temperature for agents using this provider
	MaxTokens       int     // Default max tokens for agents using this provider
	InputTokenCost  float64 // USD per 1,000,000 input tokens (for cost tracking)
	OutputTokenCost float64 // USD per 1,000,000 output tokens (for cost tracking)
	Auth            string  // Auth mode: "" (default api_key), "oauth" (use stored OAuth tokens)
}

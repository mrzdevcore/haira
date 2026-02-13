package haira

// Provider holds LLM provider configuration.
// Works with OpenAI-compatible APIs including Azure OpenAI.
type Provider struct {
	Name       string
	ApiKey     string
	Endpoint   string
	Model      string
	ApiVersion string
}

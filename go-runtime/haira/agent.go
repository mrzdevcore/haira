package haira

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// AgentConfig holds the configuration for creating an agent.
type AgentConfig struct {
	Name        string
	Provider    *Provider
	System      string
	Tools       *ToolRegistry
	Temperature float64
	Memory      MemoryConfig
}

// Agent is an LLM-powered agent that can use tools.
type Agent struct {
	config AgentConfig
	client *openai.Client
	store  *SessionStore
}

// NewAgent creates a new agent from configuration.
func NewAgent(config AgentConfig) *Agent {
	var client *openai.Client

	if config.Provider.Endpoint != "" {
		// Azure OpenAI
		fmt.Fprintf(os.Stderr, "[haira] Using Azure OpenAI: endpoint=%s model=%s api_version=%s\n",
			config.Provider.Endpoint, config.Provider.Model, config.Provider.ApiVersion)
		azCfg := openai.DefaultAzureConfig(
			config.Provider.ApiKey,
			config.Provider.Endpoint,
		)
		if config.Provider.ApiVersion != "" {
			azCfg.APIVersion = config.Provider.ApiVersion
		}
		client = openai.NewClientWithConfig(azCfg)
	} else {
		// Standard OpenAI
		client = openai.NewClient(config.Provider.ApiKey)
	}

	maxTurns := 10
	if config.Memory.MaxTurns > 0 {
		maxTurns = config.Memory.MaxTurns
	}

	return &Agent{
		config: config,
		client: client,
		store:  NewSessionStore(maxTurns),
	}
}

// Ask sends a message to the agent and returns the response.
// Implements the full tool-calling loop: send → tool calls → execute → repeat.
func (a *Agent) Ask(message string, sessionID string) (string, error) {
	ctx := context.Background()

	// Build messages: system + history + new user message
	var messages []openai.ChatCompletionMessage

	if a.config.System != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.config.System,
		})
	}

	// Append session history
	for _, msg := range a.store.GetHistory(sessionID) {
		messages = append(messages, historyToOpenAI(msg))
	}

	// Add new user message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})
	a.store.AddMessage(sessionID, Message{Role: "user", Content: message})

	// Build OpenAI tool definitions
	tools := a.buildTools()

	// Tool-calling loop (max 10 iterations)
	for i := 0; i < 10; i++ {
		req := openai.ChatCompletionRequest{
			Model:       a.config.Provider.Model,
			Messages:    messages,
			Temperature: float32(a.config.Temperature),
		}
		if len(tools) > 0 {
			req.Tools = tools
		}

		resp, err := a.client.CreateChatCompletion(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[haira] LLM API error: %v\n", err)
			return "", fmt.Errorf("LLM API error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("LLM returned no choices")
		}

		choice := resp.Choices[0]

		// No tool calls → final answer
		if len(choice.Message.ToolCalls) == 0 {
			reply := choice.Message.Content
			a.store.AddMessage(sessionID, Message{Role: "assistant", Content: reply})
			return reply, nil
		}

		// Append assistant message with tool calls
		messages = append(messages, choice.Message)

		// Execute each tool call
		for _, tc := range choice.Message.ToolCalls {
			toolResult := a.executeTool(tc)
			messages = append(messages, toolResult)
		}
	}

	return "", fmt.Errorf("tool-calling loop exceeded maximum iterations")
}

func (a *Agent) executeTool(tc openai.ToolCall) openai.ChatCompletionMessage {
	toolDef, ok := a.config.Tools.Get(tc.Function.Name)
	if !ok {
		return openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    fmt.Sprintf("error: unknown tool %q", tc.Function.Name),
			ToolCallID: tc.ID,
		}
	}

	result, err := toolDef.Handler(json.RawMessage(tc.Function.Arguments))
	var content string
	if err != nil {
		content = fmt.Sprintf("error: %s", err.Error())
	} else {
		resultJSON, _ := json.Marshal(result)
		content = string(resultJSON)
	}

	return openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    content,
		ToolCallID: tc.ID,
	}
}

func (a *Agent) buildTools() []openai.Tool {
	if a.config.Tools == nil {
		return nil
	}
	var tools []openai.Tool
	for _, td := range a.config.Tools.All() {
		var params map[string]any
		json.Unmarshal(td.Parameters, &params)

		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        td.Name,
				Description: td.Description,
				Parameters:  params,
			},
		})
	}
	return tools
}

func historyToOpenAI(msg Message) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    msg.Role,
		Content: msg.Content,
	}
}

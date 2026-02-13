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
	Handoffs    []*Agent
	Temperature float64
	Memory      MemoryConfig
}

// AgentResult holds the full result of an agent call, including handoff info.
type AgentResult struct {
	Reply       string
	HandedOffTo string // name of agent that handled the request, empty if no handoff
}

// Agent is an LLM-powered agent that can use tools.
type Agent struct {
	config AgentConfig
	client *openai.Client
	store  *SessionStore
}

const handoffToolPrefix = "transfer_to_"

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
// If the agent has handoffs, handoff tool calls are followed automatically.
func (a *Agent) Ask(message string, sessionID string) (string, error) {
	result, err := a.run(message, sessionID, true)
	if err != nil {
		return "", err
	}
	return result.Reply, nil
}

// Run sends a message to the agent and returns the full AgentResult,
// including handoff information. Handoffs are followed automatically.
// Use this when you need to inspect which agent handled the request.
func (a *Agent) Run(message string, sessionID string) (*AgentResult, error) {
	return a.run(message, sessionID, true)
}

// StreamChunk represents a piece of a streaming response.
type StreamChunk struct {
	Delta string // incremental text
	Done  bool   // true when stream is complete
}

// Stream sends a message and returns a channel that yields StreamChunks.
// The channel is closed when the response is complete.
func (a *Agent) Stream(message string, sessionID string) <-chan StreamChunk {
	ch := make(chan StreamChunk, 64)
	go func() {
		defer close(ch)
		ctx := context.Background()

		var messages []openai.ChatCompletionMessage
		if a.config.System != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: a.config.System,
			})
		}
		for _, msg := range a.store.GetHistory(sessionID) {
			messages = append(messages, historyToOpenAI(msg))
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message,
		})
		a.store.AddMessage(sessionID, Message{Role: "user", Content: message})

		req := openai.ChatCompletionRequest{
			Model:       a.config.Provider.Model,
			Messages:    messages,
			Temperature: float32(a.config.Temperature),
			Stream:      true,
		}

		stream, err := a.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			ch <- StreamChunk{Delta: fmt.Sprintf("error: %v", err), Done: true}
			return
		}
		defer stream.Close()

		var fullReply string
		for {
			resp, err := stream.Recv()
			if err != nil {
				break
			}
			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					fullReply += delta
					ch <- StreamChunk{Delta: delta, Done: false}
				}
			}
		}
		a.store.AddMessage(sessionID, Message{Role: "assistant", Content: fullReply})
		ch <- StreamChunk{Delta: "", Done: true}
	}()
	return ch
}

// run is the internal implementation shared by Ask and Run.
func (a *Agent) run(message string, sessionID string, followHandoffs bool) (*AgentResult, error) {
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

	// Build OpenAI tool definitions (includes handoff tools)
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
			return nil, fmt.Errorf("LLM API error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("LLM returned no choices")
		}

		choice := resp.Choices[0]

		// No tool calls → final answer
		if len(choice.Message.ToolCalls) == 0 {
			reply := choice.Message.Content
			a.store.AddMessage(sessionID, Message{Role: "assistant", Content: reply})
			return &AgentResult{Reply: reply}, nil
		}

		// Check for handoff tool calls
		if followHandoffs {
			for _, tc := range choice.Message.ToolCalls {
				if target := a.findHandoffTarget(tc.Function.Name); target != nil {
					fmt.Fprintf(os.Stderr, "[haira] Handoff: %s → %s\n", a.config.Name, target.config.Name)
					// Delegate to target agent with the same session
					result, err := target.run(message, sessionID, true)
					if err != nil {
						return nil, fmt.Errorf("handoff to %s failed: %w", target.config.Name, err)
					}
					result.HandedOffTo = target.config.Name
					return result, nil
				}
			}
		}

		// Append assistant message with tool calls
		messages = append(messages, choice.Message)

		// Execute each tool call (non-handoff tools)
		for _, tc := range choice.Message.ToolCalls {
			toolResult := a.executeTool(tc)
			messages = append(messages, toolResult)
		}
	}

	return nil, fmt.Errorf("tool-calling loop exceeded maximum iterations")
}

// findHandoffTarget checks if a tool call name matches a handoff target.
func (a *Agent) findHandoffTarget(toolName string) *Agent {
	for _, target := range a.config.Handoffs {
		if toolName == handoffToolPrefix+target.config.Name {
			return target
		}
	}
	return nil
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
	var tools []openai.Tool

	// Add registered tools
	if a.config.Tools != nil {
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
	}

	// Add synthetic handoff tools
	for _, target := range a.config.Handoffs {
		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        handoffToolPrefix + target.config.Name,
				Description: fmt.Sprintf("Transfer the conversation to %s.", target.config.Name),
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
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

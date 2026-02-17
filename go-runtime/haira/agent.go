package haira

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	MCPClients  []*MCPClient // MCP server connections for external tools
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

	if config.Provider.Endpoint != "" && config.Provider.ApiVersion != "" {
		// Azure OpenAI
		fmt.Fprintf(os.Stderr, "[haira] Using Azure OpenAI: endpoint=%s model=%s api_version=%s\n",
			config.Provider.Endpoint, config.Provider.Model, config.Provider.ApiVersion)
		azCfg := openai.DefaultAzureConfig(
			config.Provider.ApiKey,
			config.Provider.Endpoint,
		)
		azCfg.APIVersion = config.Provider.ApiVersion
		client = openai.NewClientWithConfig(azCfg)
	} else if config.Provider.Endpoint != "" {
		// OpenAI-compatible endpoint (Ollama, Groq, Mistral, etc.)
		fmt.Fprintf(os.Stderr, "[haira] Using OpenAI-compatible: endpoint=%s model=%s\n",
			config.Provider.Endpoint, config.Provider.Model)
		cfg := openai.DefaultConfig(config.Provider.ApiKey)
		cfg.BaseURL = config.Provider.Endpoint
		client = openai.NewClientWithConfig(cfg)
	} else {
		// Standard OpenAI
		client = openai.NewClient(config.Provider.ApiKey)
	}

	maxTurns := 10
	if config.Memory.MaxTurns > 0 {
		maxTurns = config.Memory.MaxTurns
	}

	// Connect MCP servers and register their tools
	for _, mcp := range config.MCPClients {
		if err := mcp.Connect(); err != nil {
			fmt.Fprintf(os.Stderr, "[haira] MCP connection failed for %q: %v\n", mcp.config.Name, err)
			continue
		}
		RegisterMCPClient(mcp)
		tools, err := mcp.ListTools()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[haira] MCP tool discovery failed for %q: %v\n", mcp.config.Name, err)
			continue
		}
		if config.Tools == nil {
			config.Tools = NewToolRegistry()
		}
		for _, t := range tools {
			config.Tools.Register(t)
		}
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
	Delta string // incremental text (for Type="" or "delta")
	Done  bool   // true when stream is complete

	// Event type: "" or "delta" for text, "tool_start", "tool_end", "tool_render"
	Type string
	// Tool-related fields (only set when Type is "tool_start", "tool_end", or "tool_render")
	ToolName string
	ToolArgs string // JSON args (tool_start only)
	ToolOK   bool   // success status (tool_end only)
	// Render fields (tool_render only)
	RenderComponent string // component name ("status-card", "table", etc.)
	RenderProps     string // JSON-serialized tool result
}

// Stream sends a message and returns a channel that yields StreamChunks.
// Supports tool calling: emits tool_start/tool_end events during tool execution,
// then streams the final text response.
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

		tools := a.buildTools()

		// Tool-calling loop: use non-streaming calls to handle tool iterations
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
				ch <- StreamChunk{Delta: fmt.Sprintf("error: %v", err), Done: true}
				return
			}

			if len(resp.Choices) == 0 {
				ch <- StreamChunk{Delta: "error: no response choices", Done: true}
				return
			}

			choice := resp.Choices[0]

			// No tool calls → this is the final response, stream it out
			if len(choice.Message.ToolCalls) == 0 {
				reply := choice.Message.Content
				a.store.AddMessage(sessionID, Message{Role: "assistant", Content: reply})
				// Emit the full reply as a single delta (already complete)
				if reply != "" {
					ch <- StreamChunk{Delta: reply, Done: false}
				}
				ch <- StreamChunk{Delta: "", Done: true}
				return
			}

			// Has tool calls — execute them with events
			messages = append(messages, choice.Message)

			for _, tc := range choice.Message.ToolCalls {
				// Emit tool_start
				ch <- StreamChunk{
					Type:     "tool_start",
					ToolName: tc.Function.Name,
					ToolArgs: tc.Function.Arguments,
				}

				// Execute tool
				toolResult, rawResult := a.executeTool(tc)
				messages = append(messages, toolResult)

				// Emit tool_render if the result is a UiNode
				isOK := !strings.HasPrefix(toolResult.Content, "error:")
				if isOK {
					if node, ok := rawResult.(UiNode); ok {
						renderJSON, _ := json.Marshal(rawResult)
						ch <- StreamChunk{
							Type:            "tool_render",
							ToolName:        tc.Function.Name,
							RenderComponent: node.UiComponentName(),
							RenderProps:     string(renderJSON),
						}
					}
				}

				// Emit tool_end
				ch <- StreamChunk{
					Type:     "tool_end",
					ToolName: tc.Function.Name,
					ToolOK:   isOK,
				}
			}
			// Loop continues — next iteration will get the agent's response after tool results
		}

		ch <- StreamChunk{Delta: "error: max tool iterations reached", Done: true}
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
			toolResult, _ := a.executeTool(tc)
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

func (a *Agent) executeTool(tc openai.ToolCall) (openai.ChatCompletionMessage, any) {
	toolDef, ok := a.config.Tools.Get(tc.Function.Name)
	if !ok {
		return openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    fmt.Sprintf("error: unknown tool %q", tc.Function.Name),
			ToolCallID: tc.ID,
		}, nil
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
	}, result
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

package haira

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const anthropicMessagesURL = "https://api.anthropic.com/v1/messages"
const anthropicAPIVersion = "2023-06-01"

// anthropicTransport intercepts OpenAI-format HTTP requests and translates them
// to the Anthropic Messages API format. Responses are translated back to OpenAI
// format. This allows using OAuth Bearer tokens with Anthropic's native API.
type anthropicTransport struct {
	apiKey string
}

func (t *anthropicTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Only intercept chat completions requests
	if !strings.HasSuffix(req.URL.Path, "/chat/completions") {
		return http.DefaultTransport.RoundTrip(req)
	}

	// Read the OpenAI request body
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	var oaiReq openAIRequest
	if err := json.Unmarshal(body, &oaiReq); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI request: %w", err)
	}

	// Convert to Anthropic format
	antReq := convertToAnthropic(oaiReq)

	antBody, err := json.Marshal(antReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Anthropic request: %w", err)
	}

	// Create Anthropic request
	antHTTPReq, err := http.NewRequestWithContext(req.Context(), "POST", anthropicMessagesURL, bytes.NewReader(antBody))
	if err != nil {
		return nil, err
	}
	antHTTPReq.Header.Set("Content-Type", "application/json")
	antHTTPReq.Header.Set("anthropic-version", anthropicAPIVersion)
	// Use Bearer token for OAuth, x-api-key for regular keys
	if strings.HasPrefix(t.apiKey, "sk-ant-oat") {
		antHTTPReq.Header.Set("Authorization", "Bearer "+t.apiKey)
		antHTTPReq.Header.Set("anthropic-beta", "oauth-2025-04-20")
	} else {
		antHTTPReq.Header.Set("x-api-key", t.apiKey)
	}

	resp, err := http.DefaultTransport.RoundTrip(antHTTPReq)
	if err != nil {
		return nil, err
	}

	// If error response, convert to OpenAI error format
	if resp.StatusCode != 200 {
		return convertErrorResponse(resp)
	}

	// Convert response based on streaming mode
	if oaiReq.Stream {
		return convertStreamingResponse(resp)
	}
	return convertNonStreamingResponse(resp)
}

// --- Request types ---

type openAIRequest struct {
	Model         string           `json:"model"`
	Messages      []oaiMessage     `json:"messages"`
	Temperature   float64          `json:"temperature,omitempty"`
	MaxTokens     int              `json:"max_tokens,omitempty"`
	Tools         []oaiTool        `json:"tools,omitempty"`
	Stream        bool             `json:"stream,omitempty"`
	StreamOptions *oaiStreamOpts   `json:"stream_options,omitempty"`
	ResponseFormat *oaiRespFormat  `json:"response_format,omitempty"`
}

type oaiMessage struct {
	Role       string        `json:"role"`
	Content    interface{}   `json:"content"` // string or array
	ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
}

type oaiTool struct {
	Type     string          `json:"type"`
	Function oaiFunctionDef  `json:"function"`
}

type oaiFunctionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type oaiToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function oaiFunctionCall `json:"function"`
}

type oaiFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type oaiStreamOpts struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type oaiRespFormat struct {
	Type string `json:"type"`
}

// --- Anthropic request types ---

type anthropicRequest struct {
	Model       string          `json:"model"`
	Messages    []antMessage    `json:"messages"`
	System      string          `json:"system,omitempty"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	Tools       []antTool       `json:"tools,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type antMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []antContentBlock
}

type antContentBlock struct {
	Type      string      `json:"type"`
	Text      string      `json:"text,omitempty"`
	ID        string      `json:"id,omitempty"`
	Name      string      `json:"name,omitempty"`
	Input     interface{} `json:"input,omitempty"`
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   string      `json:"content,omitempty"`
}

type antTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

// --- Anthropic response types ---

type anthropicResponse struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Role       string            `json:"role"`
	Content    []antRespContent  `json:"content"`
	StopReason string            `json:"stop_reason"`
	Usage      antUsage          `json:"usage"`
}

type antRespContent struct {
	Type  string      `json:"type"`
	Text  string      `json:"text,omitempty"`
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`
}

type antUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// --- Conversion functions ---

func convertToAnthropic(oai openAIRequest) anthropicRequest {
	ant := anthropicRequest{
		Model:       oai.Model,
		Temperature: oai.Temperature,
		MaxTokens:   oai.MaxTokens,
		Stream:      oai.Stream,
	}
	if ant.MaxTokens == 0 {
		ant.MaxTokens = 8192 // Anthropic requires max_tokens
	}

	// Extract system message and convert others
	for _, msg := range oai.Messages {
		content := messageContent(msg)
		switch msg.Role {
		case "system":
			if ant.System != "" {
				ant.System += "\n\n"
			}
			ant.System += content
		case "user":
			ant.Messages = append(ant.Messages, antMessage{Role: "user", Content: content})
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				blocks := []antContentBlock{}
				if content != "" {
					blocks = append(blocks, antContentBlock{Type: "text", Text: content})
				}
				for _, tc := range msg.ToolCalls {
					var input interface{}
					json.Unmarshal([]byte(tc.Function.Arguments), &input)
					blocks = append(blocks, antContentBlock{
						Type:  "tool_use",
						ID:    tc.ID,
						Name:  tc.Function.Name,
						Input: input,
					})
				}
				ant.Messages = append(ant.Messages, antMessage{Role: "assistant", Content: blocks})
			} else {
				ant.Messages = append(ant.Messages, antMessage{Role: "assistant", Content: content})
			}
		case "tool":
			// Tool results need to be grouped as user messages with tool_result blocks
			block := antContentBlock{
				Type:      "tool_result",
				ToolUseID: msg.ToolCallID,
				Content:   content,
			}
			// Check if last message is already a user message with tool results
			if len(ant.Messages) > 0 {
				last := &ant.Messages[len(ant.Messages)-1]
				if last.Role == "user" {
					if blocks, ok := last.Content.([]antContentBlock); ok {
						last.Content = append(blocks, block)
						continue
					}
				}
			}
			ant.Messages = append(ant.Messages, antMessage{
				Role:    "user",
				Content: []antContentBlock{block},
			})
		}
	}

	// Convert tools
	for _, t := range oai.Tools {
		ant.Tools = append(ant.Tools, antTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		})
	}

	return ant
}

func messageContent(msg oaiMessage) string {
	switch v := msg.Content.(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		// Could be an array of content blocks — marshal back and extract text
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func convertNonStreamingResponse(resp *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var antResp anthropicResponse
	if err := json.Unmarshal(body, &antResp); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	oaiResp := anthropicToOpenAIResponse(antResp)
	oaiBody, _ := json.Marshal(oaiResp)

	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(oaiBody)),
	}, nil
}

func anthropicToOpenAIResponse(ant anthropicResponse) map[string]interface{} {
	// Build message content and tool_calls
	var textContent string
	var toolCalls []map[string]interface{}

	for _, block := range ant.Content {
		switch block.Type {
		case "text":
			textContent += block.Text
		case "tool_use":
			inputJSON, _ := json.Marshal(block.Input)
			toolCalls = append(toolCalls, map[string]interface{}{
				"id":   block.ID,
				"type": "function",
				"function": map[string]interface{}{
					"name":      block.Name,
					"arguments": string(inputJSON),
				},
			})
		}
	}

	finishReason := "stop"
	if ant.StopReason == "tool_use" {
		finishReason = "tool_calls"
	} else if ant.StopReason == "max_tokens" {
		finishReason = "length"
	}

	message := map[string]interface{}{
		"role":    "assistant",
		"content": textContent,
	}
	if len(toolCalls) > 0 {
		message["tool_calls"] = toolCalls
	}

	return map[string]interface{}{
		"id":      ant.ID,
		"object":  "chat.completion",
		"model":   "",
		"choices": []map[string]interface{}{
			{
				"index":         0,
				"message":       message,
				"finish_reason": finishReason,
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     ant.Usage.InputTokens,
			"completion_tokens": ant.Usage.OutputTokens,
			"total_tokens":      ant.Usage.InputTokens + ant.Usage.OutputTokens,
		},
	}
}

// --- Streaming conversion ---

func convertStreamingResponse(resp *http.Response) (*http.Response, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		var currentToolID string
		var currentToolName string
		var inputAccum strings.Builder

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				fmt.Fprintf(pw, "data: [DONE]\n\n")
				return
			}

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			eventType, _ := event["type"].(string)
			switch eventType {
			case "message_start":
				// Emit initial chunk
				chunk := map[string]interface{}{
					"object": "chat.completion.chunk",
					"choices": []map[string]interface{}{
						{"index": 0, "delta": map[string]interface{}{"role": "assistant"}, "finish_reason": nil},
					},
				}
				writeSSE(pw, chunk)

			case "content_block_start":
				cb, _ := event["content_block"].(map[string]interface{})
				if cb != nil && cb["type"] == "tool_use" {
					currentToolID, _ = cb["id"].(string)
					currentToolName, _ = cb["name"].(string)
					inputAccum.Reset()
					// Emit tool call start
					chunk := map[string]interface{}{
						"object": "chat.completion.chunk",
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]interface{}{
									"tool_calls": []map[string]interface{}{
										{
											"index": 0,
											"id":    currentToolID,
											"type":  "function",
											"function": map[string]interface{}{
												"name":      currentToolName,
												"arguments": "",
											},
										},
									},
								},
								"finish_reason": nil,
							},
						},
					}
					writeSSE(pw, chunk)
				}

			case "content_block_delta":
				delta, _ := event["delta"].(map[string]interface{})
				if delta == nil {
					continue
				}
				deltaType, _ := delta["type"].(string)
				switch deltaType {
				case "text_delta":
					text, _ := delta["text"].(string)
					if text != "" {
						chunk := map[string]interface{}{
							"object": "chat.completion.chunk",
							"choices": []map[string]interface{}{
								{"index": 0, "delta": map[string]interface{}{"content": text}, "finish_reason": nil},
							},
						}
						writeSSE(pw, chunk)
					}
				case "input_json_delta":
					partial, _ := delta["partial_json"].(string)
					inputAccum.WriteString(partial)
					// Emit tool call argument delta
					chunk := map[string]interface{}{
						"object": "chat.completion.chunk",
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]interface{}{
									"tool_calls": []map[string]interface{}{
										{
											"index": 0,
											"function": map[string]interface{}{
												"arguments": partial,
											},
										},
									},
								},
								"finish_reason": nil,
							},
						},
					}
					writeSSE(pw, chunk)
				}

			case "message_delta":
				delta, _ := event["delta"].(map[string]interface{})
				stopReason, _ := delta["stop_reason"].(string)
				finishReason := "stop"
				if stopReason == "tool_use" {
					finishReason = "tool_calls"
				} else if stopReason == "max_tokens" {
					finishReason = "length"
				}

				// Check for usage
				var usageChunk map[string]interface{}
				if u, ok := event["usage"].(map[string]interface{}); ok {
					inTok, _ := u["output_tokens"].(float64)
					usageChunk = map[string]interface{}{
						"prompt_tokens":     0,
						"completion_tokens": int(inTok),
						"total_tokens":      int(inTok),
					}
				}

				chunk := map[string]interface{}{
					"object": "chat.completion.chunk",
					"choices": []map[string]interface{}{
						{"index": 0, "delta": map[string]interface{}{}, "finish_reason": finishReason},
					},
				}
				if usageChunk != nil {
					chunk["usage"] = usageChunk
				}
				writeSSE(pw, chunk)

			case "message_stop":
				fmt.Fprintf(pw, "data: [DONE]\n\n")
				return
			}
		}
	}()

	return &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Content-Type": {"text/event-stream"},
		},
		Body: pr,
	}, nil
}

func writeSSE(w io.Writer, data interface{}) {
	b, _ := json.Marshal(data)
	fmt.Fprintf(w, "data: %s\n\n", b)
}

func convertErrorResponse(resp *http.Response) (*http.Response, error) {
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Try to parse Anthropic error format
	var antErr anthropicError
	if err := json.Unmarshal(body, &antErr); err == nil && antErr.Error.Message != "" {
		// Convert to OpenAI error format
		oaiErr := map[string]interface{}{
			"error": map[string]interface{}{
				"message": antErr.Error.Message,
				"type":    antErr.Error.Type,
				"code":    resp.StatusCode,
			},
		}
		oaiBody, _ := json.Marshal(oaiErr)
		return &http.Response{
			StatusCode: resp.StatusCode,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(oaiBody)),
		}, nil
	}

	// Return as-is
	return &http.Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SlackBlock represents a Slack Block Kit block.
type SlackBlock map[string]any

// SlackHeader creates a header block.
func SlackHeader(text string) SlackBlock {
	return SlackBlock{
		"type": "header",
		"text": map[string]any{
			"type": "plain_text",
			"text": text,
		},
	}
}

// SlackSection creates a section block with markdown text.
func SlackSection(text string) SlackBlock {
	return SlackBlock{
		"type": "section",
		"text": map[string]any{
			"type": "mrkdwn",
			"text": text,
		},
	}
}

// SlackDivider creates a divider block.
func SlackDivider() SlackBlock {
	return SlackBlock{"type": "divider"}
}

// SlackContext creates a context block with markdown elements.
func SlackContext(texts ...string) SlackBlock {
	elements := make([]map[string]any, len(texts))
	for i, t := range texts {
		elements[i] = map[string]any{"type": "mrkdwn", "text": t}
	}
	return SlackBlock{
		"type":     "context",
		"elements": elements,
	}
}

// SlackSendBlocks sends a message with Block Kit blocks via webhook.
func SlackSendBlocks(webhookURL string, blocks []any) error {
	payload := map[string]any{"blocks": blocks}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack marshal blocks: %w", err)
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("slack send blocks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack error (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// SlackClient is an authenticated Slack bot client using the Web API.
type SlackClient struct {
	token string
}

// SlackNewClient creates a new Slack bot client.
func SlackNewClient(token string) *SlackClient {
	return &SlackClient{token: token}
}

// PostMessage sends a message to a channel using the Slack Web API.
func (c *SlackClient) PostMessage(channel, text string) error {
	payload := map[string]any{
		"channel": channel,
		"text":    text,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack marshal: %w", err)
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack send: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		return fmt.Errorf("slack api error: %s", errMsg)
	}
	return nil
}

// PostBlocks sends a message with blocks to a channel using the Slack Web API.
func (c *SlackClient) PostBlocks(channel string, blocks []any) error {
	payload := map[string]any{
		"channel": channel,
		"blocks":  blocks,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack marshal: %w", err)
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack send: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		return fmt.Errorf("slack api error: %s", errMsg)
	}
	return nil
}

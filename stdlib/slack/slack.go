package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SlackSend sends a message to a Slack channel via an incoming webhook URL.
func SlackSend(webhookURL string, channel string, text string) error {
	payload := map[string]string{
		"channel": channel,
		"text":    text,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack marshal: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("slack send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

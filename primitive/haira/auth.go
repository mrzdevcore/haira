package haira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	anthropicClientID     = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	anthropicTokenURL     = "https://platform.claude.com/v1/oauth/token"
	anthropicCreateKeyURL = "https://api.anthropic.com/api/oauth/claude_cli/create_api_key"
)

// oauthCredentials holds the OAuth token data (shared format with compiler/internal/auth).
type oauthCredentials struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresAt    int64    `json:"expiresAt"` // Unix milliseconds
	Scopes       []string `json:"scopes"`
}

type hairaCredentialsFile struct {
	AnthropicOAuth  *oauthCredentials `json:"anthropicOauth"`
	AnthropicAPIKey string            `json:"anthropicApiKey,omitempty"`
}

// ResolveAnthropicToken resolves an Anthropic auth token using this priority:
//  1. ANTHROPIC_API_KEY env var
//  2. ~/.haira/credentials.json anthropicApiKey
//  3. Haira OAuth credentials (used directly with Bearer auth)
//  4. Claude Code OAuth credentials (used directly with Bearer auth)
func ResolveAnthropicToken() (string, error) {
	// 1. Env var always wins
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}

	// 2. Stored API key
	if key, err := loadHairaAPIKey(); err == nil && key != "" {
		return key, nil
	}

	// 3. Haira OAuth credentials → use token directly (native transport sends Bearer)
	if creds, err := loadHairaCredentials(); err == nil && creds != nil {
		token, err := ensureValidToken(creds, true)
		if err == nil {
			return token, nil
		}
	}

	// 4. Claude Code OAuth credentials → use token directly
	if creds, err := loadClaudeCredentials(); err == nil && creds != nil {
		token, err := ensureValidToken(creds, false)
		if err == nil {
			return token, nil
		}
	}

	return "", fmt.Errorf("not authenticated — run 'haira auth login --anthropic' or set ANTHROPIC_API_KEY")
}

// exchangeOAuthForAPIKey uses the OAuth token to create an API key via
// Anthropic's Claude CLI endpoint (same mechanism Claude Code uses).
func exchangeOAuthForAPIKey(oauthToken string) (string, error) {
	body := strings.NewReader(`{"name":"haira-runtime"}`)
	req, err := http.NewRequest("POST", anthropicCreateKeyURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+oauthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create_api_key failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		APIKey string `json:"api_key"`
		Key    string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode create_api_key response: %w", err)
	}

	apiKey := result.APIKey
	if apiKey == "" {
		apiKey = result.Key
	}
	if apiKey == "" {
		return "", fmt.Errorf("create_api_key returned empty key")
	}

	fmt.Fprintf(os.Stderr, "[haira] Created API key from OAuth credentials\n")
	return apiKey, nil
}

// saveHairaAPIKey stores a generated API key in ~/.haira/credentials.json.
func saveHairaAPIKey(apiKey string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".haira")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "credentials.json")

	// Read existing file to preserve other fields
	var raw map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &raw)
	}
	if raw == nil {
		raw = make(map[string]interface{})
	}
	raw["anthropicApiKey"] = apiKey

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// loadHairaAPIKey reads the stored API key from ~/.haira/credentials.json.
func loadHairaAPIKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(home, ".haira", "credentials.json"))
	if err != nil {
		return "", err
	}
	var f hairaCredentialsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return "", err
	}
	return f.AnthropicAPIKey, nil
}

// ensureValidToken returns the access token, refreshing if expired.
// If save is true, updated tokens are written back to ~/.haira/credentials.json.
func ensureValidToken(creds *oauthCredentials, save bool) (string, error) {
	// Check if token is still valid (with 5-minute buffer)
	if time.Now().UnixMilli() < creds.ExpiresAt-5*60*1000 {
		return creds.AccessToken, nil
	}

	// Token expired — try to refresh
	if creds.RefreshToken == "" {
		return "", fmt.Errorf("token expired and no refresh token available")
	}

	refreshed, err := refreshAnthropicToken(creds.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("token refresh failed: %w", err)
	}

	if save {
		saveHairaCredentials(refreshed)
	}

	return refreshed.AccessToken, nil
}

// refreshAnthropicToken exchanges a refresh token for new tokens.
func refreshAnthropicToken(refreshToken string) (*oauthCredentials, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {anthropicClientID},
	}

	resp, err := http.PostForm(anthropicTokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with HTTP %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	scopes := strings.Fields(tokenResp.Scope)

	return &oauthCredentials{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().UnixMilli() + tokenResp.ExpiresIn*1000,
		Scopes:       scopes,
	}, nil
}

// forceRefreshAnthropicToken bypasses the expiry check and always refreshes.
// Used when a 401 is received — the token may have been revoked server-side
// even though it hasn't expired locally.
func forceRefreshAnthropicToken() (string, error) {
	// Try haira credentials first
	if creds, err := loadHairaCredentials(); err == nil && creds != nil && creds.RefreshToken != "" {
		refreshed, err := refreshAnthropicToken(creds.RefreshToken)
		if err == nil {
			saveHairaCredentials(refreshed)
			return refreshed.AccessToken, nil
		}
	}
	// Try Claude Code credentials
	if creds, err := loadClaudeCredentials(); err == nil && creds != nil && creds.RefreshToken != "" {
		refreshed, err := refreshAnthropicToken(creds.RefreshToken)
		if err == nil {
			// Save to haira creds so future refreshes use the new refresh token
			saveHairaCredentials(refreshed)
			return refreshed.AccessToken, nil
		}
	}
	return "", fmt.Errorf("no refresh token available — run 'haira auth login' or set ANTHROPIC_API_KEY")
}

// loadHairaCredentials reads ~/.haira/credentials.json.
func loadHairaCredentials() (*oauthCredentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".haira", "credentials.json"))
	if err != nil {
		return nil, err
	}
	var f hairaCredentialsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.AnthropicOAuth, nil
}

// saveHairaCredentials writes credentials to ~/.haira/credentials.json.
func saveHairaCredentials(creds *oauthCredentials) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".haira")
	os.MkdirAll(dir, 0700)

	f := hairaCredentialsFile{AnthropicOAuth: creds}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "credentials.json"), data, 0600)
}

// loadClaudeCredentials reads OAuth tokens from Claude Code's storage.
// On macOS, tries the Keychain first ("Claude Code-credentials" service).
// Falls back to ~/.claude.json claudeAiOauth key.
func loadClaudeCredentials() (*oauthCredentials, error) {
	// macOS: try Keychain first
	if runtime.GOOS == "darwin" {
		if creds, err := loadFromKeychain(); err == nil && creds != nil {
			return creds, nil
		}
	}

	// Fallback: try ~/.claude.json
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	oauthData, ok := raw["claudeAiOauth"]
	if !ok {
		return nil, nil
	}
	var creds oauthCredentials
	if err := json.Unmarshal(oauthData, &creds); err != nil {
		return nil, err
	}
	if creds.AccessToken == "" {
		return nil, nil
	}
	return &creds, nil
}

// loadFromKeychain reads Claude Code credentials from the macOS Keychain.
func loadFromKeychain() (*oauthCredentials, error) {
	out, err := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w").Output()
	if err != nil {
		return nil, err
	}

	var keychainData struct {
		ClaudeAiOAuth *oauthCredentials `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(out, &keychainData); err != nil {
		return nil, err
	}
	if keychainData.ClaudeAiOAuth == nil || keychainData.ClaudeAiOAuth.AccessToken == "" {
		return nil, nil
	}
	return keychainData.ClaudeAiOAuth, nil
}

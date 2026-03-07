package haira

import (
	"encoding/json"
	"fmt"
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
	anthropicClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	anthropicTokenURL = "https://platform.claude.com/v1/oauth/token"
)

// oauthCredentials holds the OAuth token data (shared format with compiler/internal/auth).
type oauthCredentials struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresAt    int64    `json:"expiresAt"` // Unix milliseconds
	Scopes       []string `json:"scopes"`
}

type hairaCredentialsFile struct {
	AnthropicOAuth *oauthCredentials `json:"anthropicOauth"`
}

// ResolveAnthropicToken resolves an Anthropic API token using this priority:
//  1. ANTHROPIC_API_KEY env var
//  2. ~/.haira/credentials.json (from `haira auth login`)
//  3. Claude Code credentials (macOS Keychain or ~/.claude.json)
func ResolveAnthropicToken() (string, error) {
	// 1. Env var always wins
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}

	// 2. Haira credentials
	if creds, err := loadHairaCredentials(); err == nil && creds != nil {
		token, err := ensureValidToken(creds, true)
		if err == nil {
			return token, nil
		}
	}

	// 3. Claude Code fallback
	if creds, err := loadClaudeCredentials(); err == nil && creds != nil {
		token, err := ensureValidToken(creds, false)
		if err == nil {
			return token, nil
		}
	}

	return "", fmt.Errorf("not authenticated — run 'haira auth login' or set ANTHROPIC_API_KEY")
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

// Package auth provides OAuth credential management for Haira programs.
// Supports importing credentials from Claude Code and storing them in ~/.haira/credentials.json.
//
// Usage from Haira:
//
//	import "auth"
//	token, err = auth.login("anthropic")
//	auth.logout()
//	status = auth.status()
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	credentialsDirName = ".haira"
	credentialsName    = "credentials.json"
)

// OAuthCredentials holds the OAuth token data.
type OAuthCredentials struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresAt    int64    `json:"expiresAt"` // Unix milliseconds
	Scopes       []string `json:"scopes"`
}

// credentialsFile is the on-disk format for ~/.haira/credentials.json.
type credentialsFile struct {
	AnthropicOAuth  *OAuthCredentials `json:"anthropicOauth"`
	AnthropicAPIKey string            `json:"anthropicApiKey,omitempty"`
}

// credentialsPath returns the path to ~/.haira/credentials.json.
func credentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, credentialsDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create %s: %w", dir, err)
	}
	return filepath.Join(dir, credentialsName), nil
}

// AuthLogin imports OAuth credentials for the given provider.
// Currently supports "anthropic" (imports from Claude Code credentials).
// Returns the access token string on success.
func AuthLogin(provider string) (string, error) {
	switch provider {
	case "anthropic", "claude":
		return loginAnthropic()
	default:
		return "", fmt.Errorf("unsupported auth provider %q (supported: anthropic)", provider)
	}
}

// AuthLogout removes stored credentials.
func AuthLogout() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing credentials: %w", err)
	}
	return nil
}

// AuthStatus returns a human-readable string describing the current auth state.
func AuthStatus() string {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		masked := key[:8] + "..." + key[len(key)-4:]
		return fmt.Sprintf("Using ANTHROPIC_API_KEY from environment: %s", masked)
	}

	// Check stored API key
	if f, err := loadCredentialsFile(); err == nil && f.AnthropicAPIKey != "" {
		key := f.AnthropicAPIKey
		masked := key[:10] + "..." + key[len(key)-4:]
		return fmt.Sprintf("Authenticated (stored API key): %s", masked)
	}

	creds, source, err := loadBestCredentials()
	if err != nil || creds == nil {
		return "Not authenticated. Run 'haira auth login --anthropic' or set ANTHROPIC_API_KEY."
	}

	expiry := time.UnixMilli(creds.ExpiresAt)
	scopeStr := strings.Join(creds.Scopes, ", ")

	if time.Now().After(expiry) {
		return fmt.Sprintf("Credentials found (%s) but expired at %s. Scopes: %s",
			source, expiry.Format(time.RFC3339), scopeStr)
	}

	return fmt.Sprintf("Authenticated (%s). Token expires: %s. Scopes: %s",
		source, expiry.Format(time.RFC3339), scopeStr)
}

// ResolveToken resolves an API token for the given provider.
// Priority: env var → ~/.haira/credentials.json → Claude Code credentials.
func ResolveToken(provider string) (string, error) {
	switch provider {
	case "anthropic", "claude":
		return resolveAnthropicToken()
	default:
		return "", fmt.Errorf("unsupported auth provider %q", provider)
	}
}

// ── Anthropic-specific ──

func loginAnthropic() (string, error) {
	// Check if ANTHROPIC_API_KEY is set
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}

	// Try to import from Claude Code
	creds, err := loadClaudeCodeCredentials()
	if err != nil {
		return "", fmt.Errorf("could not read Claude Code credentials: %w\n\nMake sure Claude Code is installed and authenticated (run 'claude' first)", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no OAuth credentials found\n\nEnsure Claude Code is installed and authenticated:\n  1. Install Claude Code: npm install -g @anthropic-ai/claude-code\n  2. Run 'claude' and complete the login\n  3. Then call auth.login(\"anthropic\") again\n\nAlternatively, set ANTHROPIC_API_KEY environment variable")
	}

	// Save a copy to ~/.haira/credentials.json
	if err := saveCredentials(creds); err != nil {
		return "", fmt.Errorf("saving credentials: %w", err)
	}

	return creds.AccessToken, nil
}

func resolveAnthropicToken() (string, error) {
	// 1. Env var always wins
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key, nil
	}

	// 2. Stored API key (from haira auth login)
	if f, err := loadCredentialsFile(); err == nil && f.AnthropicAPIKey != "" {
		return f.AnthropicAPIKey, nil
	}

	// 3. Haira OAuth credentials (legacy)
	if creds, err := loadCredentials(); err == nil && creds != nil {
		if time.Now().UnixMilli() < creds.ExpiresAt-5*60*1000 {
			return creds.AccessToken, nil
		}
	}

	// 4. Claude Code fallback
	if creds, err := loadClaudeCodeCredentials(); err == nil && creds != nil {
		if time.Now().UnixMilli() < creds.ExpiresAt-5*60*1000 {
			return creds.AccessToken, nil
		}
	}

	return "", fmt.Errorf("not authenticated — run 'haira auth login --anthropic' or set ANTHROPIC_API_KEY")
}

func loadCredentialsFile() (*credentialsFile, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f credentialsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// ── Credential storage ──

func loadBestCredentials() (*OAuthCredentials, string, error) {
	if creds, err := loadCredentials(); err == nil && creds != nil {
		return creds, "haira", nil
	}
	if creds, err := loadClaudeCodeCredentials(); err == nil && creds != nil {
		return creds, "claude-code", nil
	}
	return nil, "", nil
}

func loadCredentials() (*OAuthCredentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var f credentialsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}
	return f.AnthropicOAuth, nil
}

func saveCredentials(creds *OAuthCredentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	f := credentialsFile{AnthropicOAuth: creds}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadClaudeCodeCredentials() (*OAuthCredentials, error) {
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
	var creds OAuthCredentials
	if err := json.Unmarshal(oauthData, &creds); err != nil {
		return nil, err
	}
	if creds.AccessToken == "" {
		return nil, nil
	}
	return &creds, nil
}

func loadFromKeychain() (*OAuthCredentials, error) {
	out, err := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w").Output()
	if err != nil {
		return nil, err
	}

	var keychainData struct {
		ClaudeAiOAuth *OAuthCredentials `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(out, &keychainData); err != nil {
		return nil, err
	}
	if keychainData.ClaudeAiOAuth == nil || keychainData.ClaudeAiOAuth.AccessToken == "" {
		return nil, nil
	}
	return keychainData.ClaudeAiOAuth, nil
}

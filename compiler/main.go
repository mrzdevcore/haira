// Haira — An agentic orchestration programming language compiler.
package main

import (
	"bufio"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/haira-lang/haira/internal/console"
	"github.com/haira-lang/haira/internal/driver"
	"github.com/haira-lang/haira/internal/lsp"
	"github.com/haira-lang/haira/internal/manifest"
	"github.com/haira-lang/haira/internal/orchestrator"
	"github.com/haira-lang/haira/internal/webui"
)

var version = "dev"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	rest := args[1:]

	var err error
	switch cmd {
	case "build":
		err = cmdBuild(rest)
	case "run":
		err = cmdRun(rest)
	case "parse":
		err = cmdParse(rest)
	case "check":
		err = cmdCheck(rest)
	case "lex":
		err = cmdLex(rest)
	case "emit":
		err = cmdEmit(rest)
	case "test":
		err = cmdTest(rest)
	case "eval":
		err = cmdEval(rest)
	case "fmt":
		err = cmdFmt(rest)
	case "init":
		err = cmdInit(rest)
	case "console":
		err = cmdConsole(rest)
	case "serve":
		err = cmdServe(rest)
	case "deploy":
		err = cmdDeploy(rest)
	case "ps":
		err = cmdPs(rest)
	case "stop":
		err = cmdOrchestratorAction(rest, "stop")
	case "restart":
		err = cmdOrchestratorAction(rest, "restart")
	case "logs":
		err = cmdLogs(rest)
	case "undeploy":
		err = cmdUndeploy(rest)
	case "webui":
		err = cmdWebUI(rest)
	case "auth":
		err = cmdAuth(rest)
	case "lsp":
		err = lsp.RunStdio()
	case "version", "--version", "-v":
		fmt.Printf("haira %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// resolveEntryFile resolves the source file from args or package.haira.
// Returns the file path and remaining args.
func resolveEntryFile(args []string) (string, []string) {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	// Look for package.haira in current directory
	if pkg := manifest.Find("."); pkg != "" {
		p, err := manifest.Load(pkg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
			return "", args
		}
		return p.Entry, args
	}
	return "", args
}

func cmdBuild(args []string) error {
	file, rest := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira build [file] [-o output] [--target native|workers]\n  No file specified and no package.haira found")
	}
	output := ""
	target := "native"
	for i := 0; i < len(rest); i++ {
		if (rest[i] == "-o" || rest[i] == "--output") && i+1 < len(rest) {
			output = rest[i+1]
			i++
		} else if rest[i] == "--target" && i+1 < len(rest) {
			target = rest[i+1]
			i++
		} else if strings.HasPrefix(rest[i], "--target=") {
			target = strings.TrimPrefix(rest[i], "--target=")
		}
	}
	if target != "native" && target != "workers" && target != "claude-code" {
		return fmt.Errorf("unknown target %q (valid: native, workers, claude-code)", target)
	}
	return driver.Compile(file, output, target)
}

func cmdRun(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira run [file]\n  No file specified and no package.haira found")
	}

	// If HAIRA_UI_URL is not already set, start a local UI asset server
	// so compiled programs can serve the UI without CDN dependency.
	// The asset server runs in the background and sets HAIRA_UI_URL for the child process.
	if os.Getenv("HAIRA_UI_URL") == "" {
		if cleanup := webui.StartAssetServer(); cleanup != nil {
			defer cleanup()
		}
	}

	return driver.Run(file)
}

func cmdParse(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira parse [file]\n  No file specified and no package.haira found")
	}
	return driver.ParseFile(file)
}

func cmdCheck(args []string) error {
	if len(args) == 0 {
		file, _ := resolveEntryFile(args)
		if file == "" {
			return fmt.Errorf("usage: haira check [file] [files...]\n  No file specified and no package.haira found")
		}
		return driver.CheckFile(file)
	}
	for _, file := range args {
		if err := driver.CheckFile(file); err != nil {
			return err
		}
	}
	return nil
}

func cmdLex(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira lex [file]\n  No file specified and no package.haira found")
	}
	return driver.LexFile(file)
}

func cmdEmit(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira emit [file]\n  No file specified and no package.haira found")
	}
	return driver.EmitFile(file)
}

func cmdTest(args []string) error {
	file, rest := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira test [file] [flags...]\n  No file specified and no package.haira found")
	}
	return driver.Test(file, rest)
}

func cmdFmt(args []string) error {
	if len(args) == 0 {
		file, _ := resolveEntryFile(args)
		if file == "" {
			return fmt.Errorf("usage: haira fmt [file] [files...]\n  No file specified and no package.haira found")
		}
		return driver.FormatFile(file)
	}
	for _, file := range args {
		if err := driver.FormatFile(file); err != nil {
			return err
		}
	}
	return nil
}

func cmdEval(args []string) error {
	file, _ := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira eval [file]\n  No file specified and no package.haira found")
	}
	return driver.Eval(file)
}

func cmdConsole(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira console <host:port>\n  Connect to a running Haira server")
	}
	return console.Run(args[0], args[1:])
}

func cmdInit(args []string) error {
	// Check if package.haira already exists
	if manifest.Find(".") != "" {
		return fmt.Errorf("package.haira already exists in current directory")
	}

	// Use directory name as project name
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	name := filepath.Base(dir)

	content := manifest.DefaultManifest(name)
	if err := os.WriteFile(manifest.Filename, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", manifest.Filename, err)
	}

	fmt.Printf("Created %s\n", manifest.Filename)
	return nil
}

// --- Orchestration commands ---

const defaultOrchestratorHost = "localhost:8900"

func parseFlag(args []string, flag string, defaultVal string) (string, []string) {
	for i := 0; i < len(args); i++ {
		if (args[i] == flag || args[i] == "--"+strings.TrimPrefix(flag, "-")) && i+1 < len(args) {
			val := args[i+1]
			rest := append(args[:i], args[i+2:]...)
			return val, rest
		}
	}
	return defaultVal, args
}

func cmdServe(args []string) error {
	port := "8900"
	dataDir := ""

	port, args = parseFlag(args, "--port", port)
	dataDir, _ = parseFlag(args, "--dir", dataDir)

	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		dataDir = filepath.Join(home, ".haira", "orchestrator")
	}

	p, err := orchestrator.ParsePort(port)
	if err != nil {
		return fmt.Errorf("invalid port: %s", port)
	}

	o, err := orchestrator.New(dataDir)
	if err != nil {
		return err
	}
	return o.Serve(p)
}

func cmdDeploy(args []string) error {
	file, rest := resolveEntryFile(args)
	if file == "" {
		return fmt.Errorf("usage: haira deploy <file> [--name NAME] [--host HOST:PORT]")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Derive name from filename if not specified
	name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	host := defaultOrchestratorHost

	name, rest = parseFlag(rest, "--name", name)
	host, _ = parseFlag(rest, "--host", host)

	// POST to orchestrator
	body := fmt.Sprintf(`{"name": %q, "source_path": %q}`, name, absPath)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/_api/deploy", host),
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("connect to orchestrator at %s: %w\n  Is 'haira serve' running?", host, err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if msg, ok := result["error"].(string); ok {
			return fmt.Errorf("deploy failed: %s", msg)
		}
		return fmt.Errorf("deploy failed (status %d)", resp.StatusCode)
	}

	fmt.Printf("Deployed: %s\n", name)
	if u, ok := result["url"].(string); ok {
		fmt.Printf("URL: http://%s%s\n", host, u)
	}
	return nil
}

func cmdPs(args []string) error {
	host := defaultOrchestratorHost
	host, _ = parseFlag(args, "--host", host)

	resp, err := http.Get(fmt.Sprintf("http://%s/_api/deployments", host))
	if err != nil {
		return fmt.Errorf("connect to orchestrator at %s: %w\n  Is 'haira serve' running?", host, err)
	}
	defer resp.Body.Close()

	var deployments []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(deployments) == 0 {
		fmt.Println("No deployments.")
		return nil
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "NAME\tSTATUS\tPORT\tPID\tRESTARTS\tURL\n")
	for _, d := range deployments {
		name, _ := d["name"].(string)
		status, _ := d["status"].(string)
		port := int(d["port"].(float64))
		pid := int(d["pid"].(float64))
		restarts := int(d["restarts"].(float64))
		url, _ := d["url"].(string)
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%s\n", name, status, port, pid, restarts, url)
	}
	tw.Flush()
	return nil
}

func cmdOrchestratorAction(args []string, action string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira %s <name> [--host HOST:PORT]", action)
	}
	name := args[0]
	host := defaultOrchestratorHost
	host, _ = parseFlag(args[1:], "--host", host)

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("http://%s/_api/deployments/%s/%s", host, name, action), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to orchestrator at %s: %w\n  Is 'haira serve' running?", host, err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if msg, ok := result["error"].(string); ok {
			return fmt.Errorf("%s failed: %s", action, msg)
		}
		return fmt.Errorf("%s failed (status %d)", action, resp.StatusCode)
	}

	if s, ok := result["status"].(string); ok {
		fmt.Printf("%s: %s\n", name, s)
	}
	return nil
}

func cmdLogs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira logs <name> [--follow] [--host HOST:PORT]")
	}
	name := args[0]
	host := defaultOrchestratorHost
	follow := false

	rest := args[1:]
	host, rest = parseFlag(rest, "--host", host)
	for _, a := range rest {
		if a == "--follow" || a == "-f" {
			follow = true
		}
	}

	url := fmt.Sprintf("http://%s/_api/deployments/%s/logs", host, name)
	if follow {
		url += "?follow=true"
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("connect to orchestrator at %s: %w\n  Is 'haira serve' running?", host, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logs failed: %s", string(body))
	}

	if follow {
		// SSE stream — read data: lines
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				fmt.Println(strings.TrimPrefix(line, "data: "))
			}
		}
	} else {
		io.Copy(os.Stdout, resp.Body)
	}
	return nil
}

func cmdWebUI(args []string) error {
	port := "3000"

	// Collect all --connect / -c flags into a slice
	var backends []string
	for {
		var val string
		val, args = parseFlag(args, "--connect", "")
		if val == "" {
			val, args = parseFlag(args, "-c", "")
		}
		if val == "" {
			break
		}
		backends = append(backends, val)
	}
	if len(backends) == 0 {
		backends = []string{"localhost:8080"}
	}

	port, args = parseFlag(args, "--port", port)
	port, _ = parseFlag(args, "-p", port)

	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port: %s", port)
	}

	return webui.Run(backends, p)
}

func cmdUndeploy(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira undeploy <name> [--host HOST:PORT]")
	}
	name := args[0]
	host := defaultOrchestratorHost
	host, _ = parseFlag(args[1:], "--host", host)

	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("http://%s/_api/deployments/%s", host, name), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to orchestrator at %s: %w\n  Is 'haira serve' running?", host, err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if msg, ok := result["error"].(string); ok {
			return fmt.Errorf("undeploy failed: %s", msg)
		}
		return fmt.Errorf("undeploy failed (status %d)", resp.StatusCode)
	}

	fmt.Printf("Undeployed: %s\n", name)
	return nil
}

func cmdAuth(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: haira auth <login|logout|status> [--anthropic]")
	}

	sub := args[0]
	switch sub {
	case "login":
		provider := "anthropic"
		for _, a := range args[1:] {
			if a == "--anthropic" || a == "anthropic" {
				provider = "anthropic"
			}
		}
		if provider != "anthropic" {
			return fmt.Errorf("only --anthropic is supported")
		}

		fmt.Println("Anthropic Authentication")

		// Try to import from Claude Code OAuth credentials first
		token, oauthErr := authLoginFromOAuth()
		if oauthErr == nil && token != "" {
			masked := token[:10] + "..." + token[len(token)-4:]
			fmt.Printf("Authenticated: %s\n", masked)
			fmt.Println("Credentials stored in ~/.haira/credentials.json")
			return nil
		}

		// Fall back to manual API key entry
		fmt.Println("Could not import from Claude Code:", oauthErr)
		fmt.Println()
		fmt.Println("Enter your API key (from https://console.anthropic.com/settings/keys):")
		fmt.Print("> ")

		reader := bufio.NewReader(os.Stdin)
		apiKey, readErr := reader.ReadString('\n')
		if readErr != nil {
			return fmt.Errorf("reading input: %w", readErr)
		}
		apiKey = strings.TrimSpace(apiKey)
		if apiKey == "" {
			return fmt.Errorf("no API key provided")
		}
		if !strings.HasPrefix(apiKey, "sk-ant-") {
			return fmt.Errorf("invalid API key format (expected sk-ant-...)")
		}

		// Save to ~/.haira/credentials.json
		if err := saveAnthropicAPIKey(apiKey); err != nil {
			return err
		}
		masked := apiKey[:10] + "..." + apiKey[len(apiKey)-4:]
		fmt.Printf("\nSaved: %s\n", masked)
		fmt.Println("Credentials stored in ~/.haira/credentials.json")
		return nil

	case "logout":
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path := filepath.Join(home, ".haira", "credentials.json")
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Println("Logged out. Credentials removed.")
		return nil

	case "status":
		if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
			masked := key[:8] + "..." + key[len(key)-4:]
			fmt.Printf("Using ANTHROPIC_API_KEY from environment: %s\n", masked)
			return nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(filepath.Join(home, ".haira", "credentials.json"))
		if err != nil {
			fmt.Println("Not authenticated. Run: haira auth login --anthropic")
			return nil
		}
		var f map[string]interface{}
		json.Unmarshal(data, &f)
		if apiKey, ok := f["anthropicApiKey"].(string); ok && apiKey != "" {
			masked := apiKey[:10] + "..." + apiKey[len(apiKey)-4:]
			fmt.Printf("Authenticated (stored API key): %s\n", masked)
		} else if _, ok := f["anthropicOauth"]; ok {
			fmt.Println("Found OAuth credentials (not supported by Anthropic API).")
			fmt.Println("Run: haira auth login --anthropic  to use an API key instead.")
		} else {
			fmt.Println("Not authenticated. Run: haira auth login --anthropic")
		}
		return nil

	default:
		return fmt.Errorf("unknown auth subcommand %q (use: login, logout, status)", sub)
	}
}

const (
	oauthClientID    = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	oauthAuthorizeURL = "https://platform.claude.com/oauth/authorize"
	oauthTokenURL = "https://platform.claude.com/v1/oauth/token"
	oauthScopes   = "user:inference user:profile"
)

// authLoginFromOAuth tries two strategies:
// 1. Import existing Claude Code OAuth token and exchange for API key (no browser)
// 2. Browser-based PKCE OAuth flow with platform callback
func authLoginFromOAuth() (string, error) {
	// Strategy 1: Use existing Claude Code credentials (fastest, no browser needed)
	fmt.Println("Checking for existing Claude Code credentials...")
	fullCreds, err := loadClaudeCodeFullCredentials()
	if err == nil && fullCreds != nil {
		fmt.Println("Found Claude Code OAuth credentials. Importing...")
		saveOAuthCredentials(fullCreds)
		token, _ := fullCreds["accessToken"].(string)
		return token, nil
	}

	// Strategy 2: Browser-based PKCE flow using platform callback
	fmt.Println()
	fmt.Println("Starting browser-based authentication...")
	return authLoginBrowser()
}

// loadClaudeCodeFullCredentials reads the full OAuth credentials from Claude Code's storage.
func loadClaudeCodeFullCredentials() (map[string]interface{}, error) {
	// Try macOS Keychain first
	if out, err := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w").Output(); err == nil {
		var keychainData map[string]interface{}
		if err := json.Unmarshal(out, &keychainData); err == nil {
			if oauth, ok := keychainData["claudeAiOauth"].(map[string]interface{}); ok {
				if token, _ := oauth["accessToken"].(string); token != "" {
					return oauth, nil
				}
			}
		}
	}

	// Fallback: ~/.claude.json
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if oauth, ok := raw["claudeAiOauth"].(map[string]interface{}); ok {
		if token, _ := oauth["accessToken"].(string); token != "" {
			return oauth, nil
		}
	}
	return nil, fmt.Errorf("no OAuth credentials found in Claude Code storage")
}

// authLoginBrowser runs PKCE OAuth flow using the platform's registered callback URL.
// The user sees the auth code on the platform's success page and pastes it back.
func authLoginBrowser() (string, error) {
	verifier := generateCodeVerifier()
	challenge := generateCodeChallenge(verifier)
	state := generateCodeVerifier()

	redirectURI := "https://platform.claude.com/oauth/code/callback"

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&code_challenge=%s&code_challenge_method=S256&state=%s",
		oauthAuthorizeURL,
		oauthClientID,
		url.QueryEscape(redirectURI),
		url.QueryEscape(oauthScopes),
		challenge,
		url.QueryEscape(state),
	)

	fmt.Println("Opening browser for Anthropic authentication...")
	fmt.Println()
	if err := openBrowser(authURL); err != nil {
		fmt.Println("Could not open browser. Please visit this URL manually:")
		fmt.Println(authURL)
	}

	fmt.Println()
	fmt.Println("After approving, you'll see a success page.")
	fmt.Println("Copy the full URL from your browser's address bar and paste it here:")
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	line = strings.TrimSpace(line)

	// Parse the authorization code from the callback URL
	authCode, err := extractAuthCode(line, state)
	if err != nil {
		return "", err
	}

	// Exchange auth code for tokens
	tokenData := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {oauthClientID},
		"code":          {authCode},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
	tokenResp, err := http.PostForm(oauthTokenURL, tokenData)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != 200 {
		body, _ := io.ReadAll(tokenResp.Body)
		return "", fmt.Errorf("token exchange returned HTTP %d: %s", tokenResp.StatusCode, string(body))
	}

	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokens); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Save OAuth credentials for future refresh
	oauthCreds := map[string]interface{}{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresAt":    time.Now().UnixMilli() + tokens.ExpiresIn*1000,
		"scopes":       strings.Fields(tokens.Scope),
	}
	saveOAuthCredentials(oauthCreds)

	return tokens.AccessToken, nil
}

// extractAuthCode parses an authorization code from a callback URL or raw code input.
func extractAuthCode(input string, expectedState string) (string, error) {
	// Try parsing as URL first
	if strings.HasPrefix(input, "http") {
		u, err := url.Parse(input)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		if st := u.Query().Get("state"); st != "" && st != expectedState {
			return "", fmt.Errorf("state mismatch — possible CSRF attack")
		}
		code := u.Query().Get("code")
		if code != "" {
			return code, nil
		}
		if errMsg := u.Query().Get("error"); errMsg != "" {
			return "", fmt.Errorf("authorization error: %s", errMsg)
		}
		return "", fmt.Errorf("no authorization code found in URL")
	}
	// Treat as raw code
	if len(input) > 0 {
		return input, nil
	}
	return "", fmt.Errorf("no authorization code provided")
}

func saveOAuthCredentials(creds map[string]interface{}) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".haira")
	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, "credentials.json")

	var f map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &f)
	}
	if f == nil {
		f = make(map[string]interface{})
	}
	f["anthropicOauth"] = creds

	data, _ := json.MarshalIndent(f, "", "  ")
	os.WriteFile(path, data, 0600)
}

func generateCodeVerifier() string {
	b := make([]byte, 32)
	crand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func openBrowser(url string) error {
	switch {
	case isCommandAvailable("open"):
		return exec.Command("open", url).Start()
	case isCommandAvailable("xdg-open"):
		return exec.Command("xdg-open", url).Start()
	default:
		return fmt.Errorf("no browser opener found")
	}
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func saveAnthropicAPIKey(key string) error {
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
	var f map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &f)
	}
	if f == nil {
		f = make(map[string]interface{})
	}
	f["anthropicApiKey"] = key

	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func printUsage() {
	fmt.Print(`Haira — An agentic orchestration programming language

Usage: haira <command> [arguments]

Commands:
  build [file] [-o output] [--target native|workers|claude-code]   Compile to binary
  run [file]                  Compile and execute
  parse [file]                Show the AST
  check [file] [files...]     Parse and report errors
  lex [file]                  Show tokens
  emit [file]                 Show generated Go code
  test [file] [flags...]      Run test blocks
  eval [file]                 Run eval blocks (agent evaluation)
  fmt [file] [files...]       Format source files in-place
  init                        Create a package.haira manifest
  console <host:port>         Connect to a Haira server (interactive terminal)
  webui [--connect host:port] Serve the Haira UI (connects to a running backend)
  auth <login|logout|status>  Manage API credentials
  lsp                         Start the language server (LSP)
  version                     Show version
  help                        Show this help

Orchestration:
  serve [--port 8900]         Start the orchestration daemon
  deploy <file> [--name X]    Deploy a project to the orchestrator
  ps                          List all deployments
  stop <name>                 Stop a deployment
  restart <name>              Restart a deployment
  logs <name> [--follow]      View deployment logs
  undeploy <name>             Remove a deployment

If no file is specified, haira looks for package.haira in the current directory.

`)
}

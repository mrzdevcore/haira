// Haira — An agentic orchestration programming language compiler.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

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
		return fmt.Errorf("usage: haira build [file] [-o output]\n  No file specified and no package.haira found")
	}
	output := ""
	for i := 0; i < len(rest); i++ {
		if (rest[i] == "-o" || rest[i] == "--output") && i+1 < len(rest) {
			output = rest[i+1]
			i++
		}
	}
	return driver.Compile(file, output)
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
	json.NewDecoder(resp.Body).Decode(&result)

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
	json.NewDecoder(resp.Body).Decode(&deployments)

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
	json.NewDecoder(resp.Body).Decode(&result)

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
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != http.StatusOK {
		if msg, ok := result["error"].(string); ok {
			return fmt.Errorf("undeploy failed: %s", msg)
		}
		return fmt.Errorf("undeploy failed (status %d)", resp.StatusCode)
	}

	fmt.Printf("Undeployed: %s\n", name)
	return nil
}

func printUsage() {
	fmt.Print(`Haira — An agentic orchestration programming language

Usage: haira <command> [arguments]

Commands:
  build [file] [-o output]   Compile to a native binary
  run [file]                  Compile and execute
  parse [file]                Show the AST
  check [file] [files...]     Parse and report errors
  lex [file]                  Show tokens
  emit [file]                 Show generated Go code
  test [file] [flags...]      Run test blocks
  fmt [file] [files...]       Format source files in-place
  init                        Create a package.haira manifest
  console <host:port>         Connect to a Haira server (interactive terminal)
  webui [--connect host:port] Serve the Haira UI (connects to a running backend)
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

package console

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Workflow metadata (from /_api/workflows)
// ---------------------------------------------------------------------------

type workflowInfo struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	IsStream    bool     `json:"is_stream"`
	ChatParam   string   `json:"chat_param,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// All slash commands for tab completion.
var slashCommands = []string{"/quit", "/exit", "/help", "/workflows", "/session", "/new", "/clear"}

// ---------------------------------------------------------------------------
// Run — entry point
// ---------------------------------------------------------------------------

// Run connects to a Haira server and starts an interactive console session.
func Run(addr string, args []string) error {
	baseURL := normalizeAddr(addr)

	// Initialize terminal
	term := NewTerminal()
	if err := term.EnableRaw(); err != nil && term.IsTTY() {
		fmt.Fprintf(os.Stderr, "%sWarning: raw mode unavailable: %s%s\n", ansiDim, err, ansiReset)
	}
	defer term.Restore()
	term.WatchResize()
	term.SetupCleanExit()
	term.HandleSuspend()

	// Discover workflows
	workflows, err := discoverWorkflows(baseURL)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", baseURL, err)
	}
	if len(workflows) == 0 {
		return fmt.Errorf("no workflows found on %s", baseURL)
	}

	printWelcome(term, baseURL)

	// Select workflow
	wf, err := selectWorkflow(term, workflows)
	if err != nil {
		return err
	}

	// Generate session
	sessionID := generateSessionID()

	printBanner(term, wf, sessionID)

	// REPL
	return repl(term, baseURL, workflows, wf, sessionID)
}

// ---------------------------------------------------------------------------
// Address normalization
// ---------------------------------------------------------------------------

func normalizeAddr(addr string) string {
	if strings.HasPrefix(addr, ":") {
		addr = "localhost" + addr
	}
	addr = strings.TrimRight(addr, "/")
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	return addr
}

// ---------------------------------------------------------------------------
// Workflow discovery
// ---------------------------------------------------------------------------

func discoverWorkflows(baseURL string) ([]workflowInfo, error) {
	resp, err := http.Get(baseURL + "/_api/workflows")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var workflows []workflowInfo
	if err := json.NewDecoder(resp.Body).Decode(&workflows); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return workflows, nil
}

// ---------------------------------------------------------------------------
// Workflow selection
// ---------------------------------------------------------------------------

func selectWorkflow(term *Terminal, workflows []workflowInfo) (*workflowInfo, error) {
	streaming := filterStreaming(workflows)
	if len(streaming) == 1 {
		return streaming[0], nil
	}
	if len(workflows) == 1 {
		return &workflows[0], nil
	}

	printWorkflowList(term, workflows, nil)
	fmt.Fprintf(term.out, "Select [1-%d]: ", len(workflows))

	line := readSimpleLine(term)
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > len(workflows) {
		return nil, fmt.Errorf("invalid selection")
	}
	return &workflows[n-1], nil
}

// readSimpleLine reads a single line — uses readline in raw mode, bufio otherwise.
func readSimpleLine(term *Terminal) string {
	if term.IsRaw() {
		editor := NewLineEditor(term, "", NewHistory(0))
		line, _ := editor.ReadLine()
		return line
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

func filterStreaming(workflows []workflowInfo) []*workflowInfo {
	var result []*workflowInfo
	for i := range workflows {
		if workflows[i].IsStream {
			result = append(result, &workflows[i])
		}
	}
	return result
}

func printWorkflowList(term *Terminal, workflows []workflowInfo, active *workflowInfo) {
	fmt.Fprintf(term.out, "%sWorkflows:%s\n", ansiBold, ansiReset)
	for i, wf := range workflows {
		kind := "form"
		if wf.IsStream {
			kind = "streaming"
		}
		marker := ""
		if active != nil && wf.Path == active.Path {
			marker = ansiGreen + " [active]" + ansiReset
		}
		name := wf.Name
		if wf.Title != "" {
			name = wf.Title
		}
		fmt.Fprintf(term.out, "  %s%d.%s %s (%s) — %s%s\n",
			ansiCyan, i+1, ansiReset, name, wf.Path, kind, marker)
	}
	fmt.Fprintln(term.out)
}

// ---------------------------------------------------------------------------
// Session ID
// ---------------------------------------------------------------------------

func generateSessionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "cli-" + hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// Welcome screen
// ---------------------------------------------------------------------------

func printWelcome(term *Terminal, baseURL string) {
	fmt.Fprintln(term.out)
	w := term.Width()
	if w >= 50 {
		fmt.Fprintf(term.out, "  %s%shaira console%s\n", ansiBold, ansiYellow, ansiReset)
		fmt.Fprintf(term.out, "  %sBuild agents and workflows, not boilerplate.%s\n", ansiDim, ansiReset)
	} else {
		fmt.Fprintf(term.out, "  %shaira%s\n", ansiBold, ansiReset)
	}
	fmt.Fprintf(term.out, "  %sConnected to %s%s\n", ansiDim, baseURL, ansiReset)
	fmt.Fprintf(term.out, "  %sType /help for commands, Ctrl+D to exit%s\n", ansiDim, ansiReset)
	fmt.Fprintln(term.out)
}

// ---------------------------------------------------------------------------
// Banner
// ---------------------------------------------------------------------------

func printBanner(term *Terminal, wf *workflowInfo, sessionID string) {
	name := wf.Name
	if wf.Title != "" {
		name = wf.Title
	}
	fmt.Fprintf(term.out, "Using: %s%s%s\n", ansiBold, name, ansiReset)
	fmt.Fprintf(term.out, "Session: %s%s%s\n", ansiDim, sessionID, ansiReset)

	if len(wf.Suggestions) > 0 {
		fmt.Fprintln(term.out)
		for _, s := range wf.Suggestions {
			fmt.Fprintf(term.out, "  %s→%s %s\n", ansiDim, ansiReset, s)
		}
	}
	fmt.Fprintln(term.out)
}

// ---------------------------------------------------------------------------
// REPL
// ---------------------------------------------------------------------------

func repl(term *Terminal, baseURL string, workflows []workflowInfo, wf *workflowInfo, sessionID string) error {
	history := NewHistory(500)
	prompt := wf.Name + "> "
	editor := NewLineEditor(term, prompt, history)
	editor.SetCompletions(slashCommands)
	rend := newRenderer(term)

	for {
		editor.SetPrompt(prompt)
		line, eof := editor.ReadLine()

		if eof {
			fmt.Fprintln(term.out, "Goodbye.")
			return nil
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		history.Add(line)

		// Slash commands
		if strings.HasPrefix(line, "/") {
			cmd := strings.Fields(line)
			switch cmd[0] {
			case "/quit", "/exit":
				fmt.Fprintln(term.out, "Goodbye.")
				return nil

			case "/help":
				printHelp(term)

			case "/clear":
				term.ClearScreen()

			case "/workflows":
				printWorkflowList(term, workflows, wf)
				fmt.Fprintf(term.out, "Select [1-%d] or Enter to keep current: ", len(workflows))
				sel := readSimpleLine(term)
				sel = strings.TrimSpace(sel)
				if sel != "" {
					n, err := strconv.Atoi(sel)
					if err == nil && n >= 1 && n <= len(workflows) {
						wf = &workflows[n-1]
						sessionID = generateSessionID()
						prompt = wf.Name + "> "
						printBanner(term, wf, sessionID)
					}
				}

			case "/session":
				if len(cmd) > 1 {
					sessionID = cmd[1]
				}
				fmt.Fprintf(term.out, "Session: %s%s%s\n\n", ansiDim, sessionID, ansiReset)

			case "/new":
				sessionID = generateSessionID()
				fmt.Fprintf(term.out, "New session: %s%s%s\n\n", ansiDim, sessionID, ansiReset)

			default:
				fmt.Fprintf(term.out, "%sUnknown command: %s. Type /help for commands.%s\n\n", ansiDim, cmd[0], ansiReset)
			}
			continue
		}

		// Send message and stream response with cancellation support
		ctx, cancel := context.WithCancel(context.Background())

		// Ctrl+C during streaming cancels the request
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			select {
			case <-sigCh:
				cancel()
			case <-ctx.Done():
			}
			signal.Stop(sigCh)
		}()

		if err := sendAndStream(ctx, baseURL, wf, sessionID, line, rend); err != nil {
			if ctx.Err() == context.Canceled {
				fmt.Fprintf(term.out, "\n%s(cancelled)%s\n", ansiDim, ansiReset)
			} else {
				fmt.Fprintf(term.out, "%s%sError: %s%s\n", ansiRed, ansiBold, err, ansiReset)
			}
		}
		cancel()
		fmt.Fprintln(term.out)
	}
}

func printHelp(term *Terminal) {
	fmt.Fprint(term.out, `
Commands:
  /workflows     List and switch workflows
  /session [id]  Show or switch to a session
  /new           Start a new session
  /clear         Clear screen
  /help          Show this help
  /quit          Exit

Shortcuts:
  Ctrl+C         Cancel current input or request
  Ctrl+D         Exit (on empty line)
  Ctrl+L         Clear screen
  Ctrl+A / Home  Move to start of line
  Ctrl+E / End   Move to end of line
  Ctrl+K         Delete to end of line
  Ctrl+U         Delete to start of line
  Ctrl+W         Delete previous word
  Up/Down        Navigate command history
  Tab            Complete slash commands
  \              Continue input on next line
`)
}

// ---------------------------------------------------------------------------
// HTTP SSE client
// ---------------------------------------------------------------------------

func sendAndStream(ctx context.Context, baseURL string, wf *workflowInfo, sessionID, message string, rend *renderer) error {
	chatParam := wf.ChatParam
	if chatParam == "" {
		chatParam = "message"
	}
	body := map[string]any{
		chatParam:    message,
		"session_id": sessionID,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := baseURL + wf.Path
	req, err := http.NewRequestWithContext(ctx, wf.Method, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Start thinking spinner
	rend.startThinking()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		rend.stopThinking()
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		rend.stopThinking()
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return parseSSE(ctx, resp.Body, rend)
}

// parseSSE reads an SSE stream and dispatches events to the renderer.
func parseSSE(ctx context.Context, body io.Reader, rend *renderer) error {
	scanner := bufio.NewScanner(body)
	var currentEvent string

	for scanner.Scan() {
		// Check for cancellation
		if ctx.Err() != nil {
			rend.stopThinking()
			return ctx.Err()
		}

		line := scanner.Text()

		if line == "" {
			currentEvent = ""
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				rend.stopThinking()
				rend.renderEvent("", "[DONE]")
				return nil
			}

			// Stop thinking spinner on first data/tool event
			rend.stopThinking()
			rend.renderEvent(currentEvent, data)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		rend.stopThinking()
		return err
	}

	rend.stopThinking()
	fmt.Fprintln(rend.out)
	return nil
}

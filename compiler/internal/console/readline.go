package console

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------------
// History — command history ring buffer
// ---------------------------------------------------------------------------

// History stores command history for up/down arrow navigation.
type History struct {
	entries []string
	pos     int    // current browsing position; -1 = not browsing
	max     int    // maximum entries
	saved   string // the line being edited when user starts browsing
}

// NewHistory creates a history buffer with the given max capacity.
func NewHistory(max int) *History {
	return &History{
		entries: make([]string, 0, 64),
		pos:     -1,
		max:     max,
	}
}

// Add appends a line to history, deduplicating consecutive entries.
func (h *History) Add(line string) {
	if line == "" {
		return
	}
	// Dedup consecutive
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == line {
		return
	}
	h.entries = append(h.entries, line)
	if len(h.entries) > h.max {
		h.entries = h.entries[1:]
	}
}

// Previous navigates up in history. On first call, saves the current line.
func (h *History) Previous(current string) (string, bool) {
	if len(h.entries) == 0 {
		return current, false
	}
	if h.pos == -1 {
		// Start browsing from the end
		h.saved = current
		h.pos = len(h.entries) - 1
	} else if h.pos > 0 {
		h.pos--
	} else {
		return h.entries[0], false // Already at oldest
	}
	return h.entries[h.pos], true
}

// Next navigates down in history. Returns the saved line when reaching the end.
func (h *History) Next() (string, bool) {
	if h.pos == -1 {
		return "", false
	}
	if h.pos < len(h.entries)-1 {
		h.pos++
		return h.entries[h.pos], true
	}
	// Past the newest entry — restore saved line
	h.pos = -1
	return h.saved, true
}

// Reset stops history browsing.
func (h *History) Reset() {
	h.pos = -1
	h.saved = ""
}

// ---------------------------------------------------------------------------
// LineEditor — interactive line editing
// ---------------------------------------------------------------------------

// LineEditor provides readline-like line editing with history and tab completion.
type LineEditor struct {
	term        *Terminal
	history     *History
	prompt      string
	buf         []rune
	pos         int      // cursor position in buf
	killRing    string   // last killed text
	multiLine   bool     // in multi-line continuation mode
	lines       []string // accumulated multi-line lines
	completions []string // available completions (slash commands)
	scanner     *bufio.Scanner // for non-TTY fallback
}

// NewLineEditor creates a line editor bound to a terminal.
func NewLineEditor(term *Terminal, prompt string, history *History) *LineEditor {
	return &LineEditor{
		term:    term,
		history: history,
		prompt:  prompt,
	}
}

// SetPrompt changes the prompt string.
func (le *LineEditor) SetPrompt(prompt string) {
	le.prompt = prompt
}

// SetCompletions sets the list of tokens available for tab completion.
func (le *LineEditor) SetCompletions(completions []string) {
	le.completions = completions
}

// ReadLine reads a complete line from the user. Returns the line and whether
// the input was terminated by EOF (Ctrl+D on empty line). Returns ("", false)
// on Ctrl+C (line cleared). Falls back to bufio.Scanner if not a TTY.
func (le *LineEditor) ReadLine() (string, bool) {
	if !le.term.IsRaw() {
		return le.readLineFallback()
	}

	le.buf = le.buf[:0]
	le.pos = 0
	le.history.Reset()

	if !le.multiLine {
		le.redraw()
	} else {
		le.redrawContinuation()
	}

	for {
		key := le.term.ReadKey()

		switch key.Type {
		case KeyRune:
			le.insertRune(key.Rune)

		case KeyEnter:
			line := string(le.buf)
			fmt.Fprintln(le.term.out) // Move past input line

			// Multi-line continuation: line ending with backslash
			if strings.HasSuffix(line, "\\") {
				le.lines = append(le.lines, strings.TrimSuffix(line, "\\"))
				le.buf = le.buf[:0]
				le.pos = 0
				le.multiLine = true
				le.redrawContinuation()
				continue
			}

			if le.multiLine {
				le.lines = append(le.lines, line)
				result := strings.Join(le.lines, "\n")
				le.lines = le.lines[:0]
				le.multiLine = false
				return result, false
			}
			return line, false

		case KeyBackspace:
			le.deleteBackward()

		case KeyDelete:
			le.deleteForward()

		case KeyLeft, KeyCtrlB:
			le.moveLeft()

		case KeyRight, KeyCtrlF:
			le.moveRight()

		case KeyHome, KeyCtrlA:
			le.moveHome()

		case KeyEnd, KeyCtrlE:
			le.moveEnd()

		case KeyUp, KeyCtrlP:
			le.historyUp()

		case KeyDown, KeyCtrlN:
			le.historyDown()

		case KeyCtrlC:
			// Clear line and re-prompt
			le.buf = le.buf[:0]
			le.pos = 0
			le.multiLine = false
			le.lines = le.lines[:0]
			fmt.Fprintln(le.term.out)
			return "", false

		case KeyCtrlD:
			if len(le.buf) == 0 && !le.multiLine {
				return "", true // EOF
			}
			le.deleteForward()

		case KeyCtrlK:
			le.killToEnd()

		case KeyCtrlU:
			le.killToStart()

		case KeyCtrlW:
			le.killWord()

		case KeyCtrlL:
			le.clearScreen()

		case KeyTab:
			le.tryComplete()

		case KeyEscape, KeyUnknown:
			// Ignore
		}
	}
}

// ---------------------------------------------------------------------------
// Editing operations
// ---------------------------------------------------------------------------

func (le *LineEditor) insertRune(r rune) {
	if le.pos == len(le.buf) {
		le.buf = append(le.buf, r)
	} else {
		le.buf = append(le.buf, 0)
		copy(le.buf[le.pos+1:], le.buf[le.pos:])
		le.buf[le.pos] = r
	}
	le.pos++
	le.redraw()
}

func (le *LineEditor) deleteBackward() {
	if le.pos > 0 {
		le.buf = append(le.buf[:le.pos-1], le.buf[le.pos:]...)
		le.pos--
		le.redraw()
	}
}

func (le *LineEditor) deleteForward() {
	if le.pos < len(le.buf) {
		le.buf = append(le.buf[:le.pos], le.buf[le.pos+1:]...)
		le.redraw()
	}
}

func (le *LineEditor) moveLeft() {
	if le.pos > 0 {
		le.pos--
		le.redraw()
	}
}

func (le *LineEditor) moveRight() {
	if le.pos < len(le.buf) {
		le.pos++
		le.redraw()
	}
}

func (le *LineEditor) moveHome() {
	le.pos = 0
	le.redraw()
}

func (le *LineEditor) moveEnd() {
	le.pos = len(le.buf)
	le.redraw()
}

func (le *LineEditor) killToEnd() {
	if le.pos < len(le.buf) {
		le.killRing = string(le.buf[le.pos:])
		le.buf = le.buf[:le.pos]
		le.redraw()
	}
}

func (le *LineEditor) killToStart() {
	if le.pos > 0 {
		le.killRing = string(le.buf[:le.pos])
		le.buf = le.buf[le.pos:]
		le.pos = 0
		le.redraw()
	}
}

func (le *LineEditor) killWord() {
	if le.pos == 0 {
		return
	}
	// Find start of word: skip trailing spaces, then skip non-spaces
	end := le.pos
	start := le.pos
	for start > 0 && unicode.IsSpace(le.buf[start-1]) {
		start--
	}
	for start > 0 && !unicode.IsSpace(le.buf[start-1]) {
		start--
	}
	le.killRing = string(le.buf[start:end])
	le.buf = append(le.buf[:start], le.buf[end:]...)
	le.pos = start
	le.redraw()
}

// ---------------------------------------------------------------------------
// History navigation
// ---------------------------------------------------------------------------

func (le *LineEditor) historyUp() {
	if entry, ok := le.history.Previous(string(le.buf)); ok {
		le.buf = []rune(entry)
		le.pos = len(le.buf)
		le.redraw()
	}
}

func (le *LineEditor) historyDown() {
	if entry, ok := le.history.Next(); ok {
		le.buf = []rune(entry)
		le.pos = len(le.buf)
		le.redraw()
	}
}

// ---------------------------------------------------------------------------
// Tab completion
// ---------------------------------------------------------------------------

func (le *LineEditor) tryComplete() {
	line := string(le.buf)
	if !strings.HasPrefix(line, "/") {
		return
	}

	prefix := line
	var matches []string
	for _, cmd := range le.completions {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 0 {
		return
	}

	if len(matches) == 1 {
		// Single match — auto-complete with trailing space
		le.buf = []rune(matches[0] + " ")
		le.pos = len(le.buf)
		le.redraw()
		return
	}

	// Multiple matches — find longest common prefix
	common := matches[0]
	for _, m := range matches[1:] {
		common = commonPrefix(common, m)
	}
	if len(common) > len(prefix) {
		le.buf = []rune(common)
		le.pos = len(le.buf)
		le.redraw()
		return
	}

	// Show all matches
	fmt.Fprintln(le.term.out)
	for _, m := range matches {
		fmt.Fprintf(le.term.out, "  %s\n", m)
	}
	le.redraw()
}

func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// ---------------------------------------------------------------------------
// Screen operations
// ---------------------------------------------------------------------------

func (le *LineEditor) clearScreen() {
	le.term.ClearScreen()
	le.redraw()
}

func (le *LineEditor) redraw() {
	prompt := le.prompt
	if le.multiLine {
		prompt = "... "
	}
	le.term.ClearLine()
	fmt.Fprint(le.term.out, prompt)
	fmt.Fprint(le.term.out, string(le.buf))
	// Reposition cursor if not at the end
	if le.pos < len(le.buf) {
		back := len(le.buf) - le.pos
		fmt.Fprintf(le.term.out, "\033[%dD", back)
	}
}

func (le *LineEditor) redrawContinuation() {
	le.term.ClearLine()
	fmt.Fprint(le.term.out, "... ")
	fmt.Fprint(le.term.out, string(le.buf))
	if le.pos < len(le.buf) {
		back := len(le.buf) - le.pos
		fmt.Fprintf(le.term.out, "\033[%dD", back)
	}
}

// ---------------------------------------------------------------------------
// Non-TTY fallback
// ---------------------------------------------------------------------------

func (le *LineEditor) readLineFallback() (string, bool) {
	if le.scanner == nil {
		le.scanner = bufio.NewScanner(le.term.in)
	}
	if !le.scanner.Scan() {
		return "", true // EOF
	}
	return le.scanner.Text(), false
}

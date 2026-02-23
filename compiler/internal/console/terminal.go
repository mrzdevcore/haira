package console

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// Key — parsed keystroke
// ---------------------------------------------------------------------------

// KeyType represents the type of a parsed keystroke.
type KeyType int

const (
	KeyRune      KeyType = iota // Printable character (stored in Key.Rune)
	KeyEnter                    // Enter / Return
	KeyBackspace                // Backspace
	KeyDelete                   // Delete (ESC[3~)
	KeyTab                      // Tab
	KeyLeft                     // Left arrow
	KeyRight                    // Right arrow
	KeyUp                       // Up arrow
	KeyDown                     // Down arrow
	KeyHome                     // Home (ESC[H or Ctrl+A)
	KeyEnd                      // End (ESC[F or Ctrl+E)
	KeyCtrlA                    // Ctrl+A — home
	KeyCtrlB                    // Ctrl+B — back one char
	KeyCtrlC                    // Ctrl+C — cancel
	KeyCtrlD                    // Ctrl+D — EOF / delete forward
	KeyCtrlE                    // Ctrl+E — end
	KeyCtrlF                    // Ctrl+F — forward one char
	KeyCtrlK                    // Ctrl+K — kill to end
	KeyCtrlL                    // Ctrl+L — clear screen
	KeyCtrlN                    // Ctrl+N — next history
	KeyCtrlP                    // Ctrl+P — previous history
	KeyCtrlU                    // Ctrl+U — kill to start
	KeyCtrlW                    // Ctrl+W — kill word
	KeyEscape                   // Bare ESC
	KeyUnknown                  // Unrecognized
)

// Key represents a parsed keystroke.
type Key struct {
	Type KeyType
	Rune rune
}

// ---------------------------------------------------------------------------
// Terminal — owns raw mode, input reading, output helpers
// ---------------------------------------------------------------------------

// Terminal manages raw terminal state and provides input/output primitives.
type Terminal struct {
	in      *os.File   // stdin
	out     *os.File   // stderr (prompts, spinners, status)
	content *os.File   // stdout (response content)
	fd      uintptr    // stdin file descriptor
	isTTY   bool       // whether stdin is a terminal
	raw     bool       // whether raw mode is active
	restore func()     // restores original terminal state
	cols    atomic.Int32
	rows    atomic.Int32
	mu      sync.Mutex // protects raw mode transitions
}

// NewTerminal creates a terminal. Prompts and status go to stderr,
// content goes to stdout. If stdin is not a TTY, raw mode is unavailable.
func NewTerminal() *Terminal {
	t := &Terminal{
		in:      os.Stdin,
		out:     os.Stderr,
		content: os.Stdout,
		fd:      os.Stdin.Fd(),
		isTTY:   isTTY(),
	}
	// Initialize terminal size
	if cols, rows, err := getWinsize(t.fd); err == nil {
		t.cols.Store(int32(cols))
		t.rows.Store(int32(rows))
	} else {
		t.cols.Store(80)
		t.rows.Store(24)
	}
	return t
}

// IsTTY returns true if stdin is connected to a terminal.
func (t *Terminal) IsTTY() bool {
	return t.isTTY
}

// IsRaw returns true if the terminal is in raw mode.
func (t *Terminal) IsRaw() bool {
	return t.raw
}

// EnableRaw switches to raw mode. No-op if not a TTY or already raw.
func (t *Terminal) EnableRaw() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.isTTY || t.raw {
		return nil
	}
	restore, err := enableRawMode(t.fd)
	if err != nil {
		return err
	}
	t.restore = restore
	t.raw = true
	return nil
}

// Restore returns the terminal to its original state. Safe to call multiple times.
func (t *Terminal) Restore() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.raw && t.restore != nil {
		t.restore()
		t.raw = false
	}
}

// Width returns the terminal column count.
func (t *Terminal) Width() int {
	w := int(t.cols.Load())
	if w <= 0 {
		return 80
	}
	return w
}

// Height returns the terminal row count.
func (t *Terminal) Height() int {
	h := int(t.rows.Load())
	if h <= 0 {
		return 24
	}
	return h
}

// WatchResize listens for SIGWINCH and updates terminal dimensions.
// No-op on Windows (no SIGWINCH).
func (t *Terminal) WatchResize() {
	sig := winchSignal()
	if sig == nil {
		return
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sig)
	go func() {
		for range ch {
			if cols, rows, err := getWinsize(t.fd); err == nil {
				t.cols.Store(int32(cols))
				t.rows.Store(int32(rows))
			}
		}
	}()
}

// SetupCleanExit ensures the terminal is restored on SIGTERM/SIGHUP.
// On Windows, listens for os.Interrupt only.
func (t *Terminal) SetupCleanExit() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, cleanExitSignals()...)
	go func() {
		<-ch
		t.Restore()
		os.Exit(1)
	}()
}

// HandleSuspend handles Ctrl+Z (SIGTSTP) by restoring terminal state,
// then re-enabling raw mode on SIGCONT.
// No-op on Windows (no job control signals).
func (t *Terminal) HandleSuspend() {
	if runtime.GOOS == "windows" {
		return
	}
	susp := suspendSignal()
	cont := continueSignal()
	if susp == nil || cont == nil {
		return
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, susp, cont)
	go func() {
		for sig := range ch {
			if sig == susp {
				t.Restore()
				signal.Stop(ch)
				sendSuspendSignal()
			} else {
				t.EnableRaw()
				signal.Notify(ch, susp, cont)
			}
		}
	}()
}

// ---------------------------------------------------------------------------
// ReadKey — keystroke parser
// ---------------------------------------------------------------------------

// ReadKey reads a single keystroke from stdin. Blocks until input is available.
// Handles multi-byte escape sequences (arrow keys, home, end, delete) and
// UTF-8 multi-byte characters.
func (t *Terminal) ReadKey() Key {
	buf := make([]byte, 1)
	if _, err := t.in.Read(buf); err != nil {
		return Key{Type: KeyCtrlD} // Treat read error as EOF
	}
	b := buf[0]

	// Control characters (0x00-0x1f)
	switch b {
	case 0x01: // Ctrl+A
		return Key{Type: KeyCtrlA}
	case 0x02: // Ctrl+B
		return Key{Type: KeyCtrlB}
	case 0x03: // Ctrl+C
		return Key{Type: KeyCtrlC}
	case 0x04: // Ctrl+D
		return Key{Type: KeyCtrlD}
	case 0x05: // Ctrl+E
		return Key{Type: KeyCtrlE}
	case 0x06: // Ctrl+F
		return Key{Type: KeyCtrlF}
	case 0x09: // Tab
		return Key{Type: KeyTab}
	case 0x0a, 0x0d: // Enter (LF or CR)
		return Key{Type: KeyEnter}
	case 0x0b: // Ctrl+K
		return Key{Type: KeyCtrlK}
	case 0x0c: // Ctrl+L
		return Key{Type: KeyCtrlL}
	case 0x0e: // Ctrl+N
		return Key{Type: KeyCtrlN}
	case 0x10: // Ctrl+P
		return Key{Type: KeyCtrlP}
	case 0x15: // Ctrl+U
		return Key{Type: KeyCtrlU}
	case 0x17: // Ctrl+W
		return Key{Type: KeyCtrlW}
	case 0x1b: // ESC — start of escape sequence
		return t.readEscapeSequence()
	case 0x7f: // Backspace (macOS sends 0x7f)
		return Key{Type: KeyBackspace}
	case 0x08: // Backspace (some terminals send 0x08)
		return Key{Type: KeyBackspace}
	}

	// UTF-8 multi-byte character
	if b >= 0x80 {
		return t.readUTF8(b)
	}

	// Printable ASCII (0x20-0x7e)
	if b >= 0x20 && b < 0x7f {
		return Key{Type: KeyRune, Rune: rune(b)}
	}

	return Key{Type: KeyUnknown}
}

// readEscapeSequence parses an escape sequence after ESC (0x1b).
func (t *Terminal) readEscapeSequence() Key {
	buf := make([]byte, 1)

	// Read second byte with a short window — if nothing follows ESC quickly,
	// treat it as a bare Escape.
	if _, err := t.in.Read(buf); err != nil {
		return Key{Type: KeyEscape}
	}

	switch buf[0] {
	case '[': // CSI sequence
		return t.readCSI()
	case 'O': // SS3 sequence (some terminals send ESC O A for arrows)
		if _, err := t.in.Read(buf); err != nil {
			return Key{Type: KeyEscape}
		}
		switch buf[0] {
		case 'A':
			return Key{Type: KeyUp}
		case 'B':
			return Key{Type: KeyDown}
		case 'C':
			return Key{Type: KeyRight}
		case 'D':
			return Key{Type: KeyLeft}
		case 'H':
			return Key{Type: KeyHome}
		case 'F':
			return Key{Type: KeyEnd}
		}
		return Key{Type: KeyUnknown}
	}

	return Key{Type: KeyEscape}
}

// readCSI parses a CSI sequence (ESC [ ...).
func (t *Terminal) readCSI() Key {
	buf := make([]byte, 1)

	// Read parameter bytes and final byte
	var params []byte
	for {
		if _, err := t.in.Read(buf); err != nil {
			return Key{Type: KeyUnknown}
		}
		b := buf[0]
		// Parameter bytes: 0x30-0x3f (digits, semicolons, etc.)
		if b >= 0x30 && b <= 0x3f {
			params = append(params, b)
			continue
		}
		// Final byte: 0x40-0x7e
		if b >= 0x40 && b <= 0x7e {
			return t.dispatchCSI(params, b)
		}
		// Intermediate bytes: 0x20-0x2f — skip and continue
		if b >= 0x20 && b <= 0x2f {
			continue
		}
		return Key{Type: KeyUnknown}
	}
}

// dispatchCSI maps a CSI final byte (with optional params) to a Key.
func (t *Terminal) dispatchCSI(params []byte, final byte) Key {
	switch final {
	case 'A': // Up
		return Key{Type: KeyUp}
	case 'B': // Down
		return Key{Type: KeyDown}
	case 'C': // Right
		return Key{Type: KeyRight}
	case 'D': // Left
		return Key{Type: KeyLeft}
	case 'H': // Home
		return Key{Type: KeyHome}
	case 'F': // End
		return Key{Type: KeyEnd}
	case '~': // Tilde sequences: ESC[1~=Home, ESC[3~=Delete, ESC[4~=End
		if len(params) == 1 {
			switch params[0] {
			case '1':
				return Key{Type: KeyHome}
			case '3':
				return Key{Type: KeyDelete}
			case '4':
				return Key{Type: KeyEnd}
			}
		}
	}
	return Key{Type: KeyUnknown}
}

// readUTF8 reads a multi-byte UTF-8 character given the first byte.
func (t *Terminal) readUTF8(first byte) Key {
	// Determine number of continuation bytes from leading bits
	var size int
	switch {
	case first&0xe0 == 0xc0:
		size = 2
	case first&0xf0 == 0xe0:
		size = 3
	case first&0xf8 == 0xf0:
		size = 4
	default:
		return Key{Type: KeyUnknown}
	}

	bytes := make([]byte, size)
	bytes[0] = first
	for i := 1; i < size; i++ {
		b := make([]byte, 1)
		if _, err := t.in.Read(b); err != nil {
			return Key{Type: KeyUnknown}
		}
		bytes[i] = b[0]
	}

	runes := []rune(string(bytes))
	if len(runes) == 1 {
		return Key{Type: KeyRune, Rune: runes[0]}
	}
	return Key{Type: KeyUnknown}
}

// ---------------------------------------------------------------------------
// ANSI output helpers — write to stderr (t.out)
// ---------------------------------------------------------------------------

// ClearLine clears the current line and moves cursor to column 0.
func (t *Terminal) ClearLine() {
	fmt.Fprint(t.out, "\033[2K\r")
}

// ClearScreen clears the screen and moves cursor to top-left.
func (t *Terminal) ClearScreen() {
	fmt.Fprint(t.out, "\033[2J\033[H")
}

// ShowCursor makes the cursor visible.
func (t *Terminal) ShowCursor() {
	fmt.Fprint(t.out, "\033[?25h")
}

// HideCursor makes the cursor invisible.
func (t *Terminal) HideCursor() {
	fmt.Fprint(t.out, "\033[?25l")
}

// MoveUp moves the cursor up n lines.
func (t *Terminal) MoveUp(n int) {
	if n > 0 {
		fmt.Fprintf(t.out, "\033[%dA", n)
	}
}

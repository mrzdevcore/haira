package console

import (
	"fmt"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Spinner — animated status indicator
// ---------------------------------------------------------------------------

// Spinner animates a status message on a single terminal line.
type Spinner struct {
	term    *Terminal
	message string
	frames  []string
	index   int
	active  bool
	mu      sync.Mutex
	done    chan struct{}
}

// Braille spinner frames for a smooth animation.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewSpinner creates a spinner that writes to the terminal's status output.
func NewSpinner(term *Terminal, message string) *Spinner {
	return &Spinner{
		term:    term,
		message: message,
		frames:  spinnerFrames,
		done:    make(chan struct{}),
	}
}

// Start begins the spinner animation in a background goroutine.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				if !s.active {
					s.mu.Unlock()
					return
				}
				frame := s.frames[s.index%len(s.frames)]
				s.index++
				msg := s.message
				s.mu.Unlock()
				fmt.Fprintf(s.term.out, "\033[2K\r%s  %s %s%s", ansiDim, frame, msg, ansiReset)
			}
		}
	}()
}

// Update changes the spinner message while it's running.
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Stop halts the spinner and clears its line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()
	close(s.done)
	// Small sleep to let the last tick finish writing
	time.Sleep(10 * time.Millisecond)
	fmt.Fprint(s.term.out, "\033[2K\r")
}

// StopWithMessage halts the spinner and replaces it with a final message.
func (s *Spinner) StopWithMessage(msg string) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()
	close(s.done)
	time.Sleep(10 * time.Millisecond)
	fmt.Fprintf(s.term.out, "\033[2K\r%s\n", msg)
}

// IsActive returns whether the spinner is currently running.
func (s *Spinner) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// ---------------------------------------------------------------------------
// ToolSpinner — convenience wrapper for tool execution
// ---------------------------------------------------------------------------

// ToolSpinner tracks a tool execution with an animated spinner.
type ToolSpinner struct {
	term    *Terminal
	spinner *Spinner
	name    string
}

// NewToolSpinner creates a spinner for a named tool.
func NewToolSpinner(term *Terminal, name string) *ToolSpinner {
	return &ToolSpinner{
		term:    term,
		spinner: NewSpinner(term, name+" ..."),
		name:    name,
	}
}

// Start begins the tool spinner animation.
func (ts *ToolSpinner) Start() {
	ts.spinner.Start()
}

// Done stops the spinner with a success/failure indicator.
func (ts *ToolSpinner) Done(ok bool) {
	if ok {
		ts.spinner.StopWithMessage(fmt.Sprintf("%s  ✓ %s%s", ansiGreen, ts.name, ansiReset))
	} else {
		ts.spinner.StopWithMessage(fmt.Sprintf("%s  ✗ %s%s", ansiRed, ts.name, ansiReset))
	}
}

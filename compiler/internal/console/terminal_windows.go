//go:build windows

package console

import (
	"errors"
	"os"
)

// enableRawMode is a no-op on Windows. The interactive console is not
// supported on Windows; the CLI still works for build/run/deploy commands.
func enableRawMode(fd uintptr) (restore func(), err error) {
	return nil, errors.New("raw mode not supported on Windows")
}

// getWinsize returns default dimensions on Windows.
func getWinsize(fd uintptr) (cols, rows int, err error) {
	return 80, 24, nil
}

// suspendSignal returns nil on Windows (no SIGTSTP).
func suspendSignal() os.Signal {
	return nil
}

// continueSignal returns nil on Windows (no SIGCONT).
func continueSignal() os.Signal {
	return nil
}

// winchSignal returns nil on Windows (no SIGWINCH).
func winchSignal() os.Signal {
	return nil
}

// cleanExitSignals returns signals that should trigger a clean exit on Windows.
func cleanExitSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}

// sendSuspendSignal is a no-op on Windows (no job control).
func sendSuspendSignal() {}

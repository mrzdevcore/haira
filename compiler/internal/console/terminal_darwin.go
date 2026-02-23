//go:build darwin

package console

import (
	"os"
	"syscall"
	"unsafe"
)

// enableRawMode puts the terminal into raw mode and returns a restore function.
// On darwin, uses TIOCGETA/TIOCSETA ioctls with syscall.Termios.
func enableRawMode(fd uintptr) (restore func(), err error) {
	var orig syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCGETA), uintptr(unsafe.Pointer(&orig))); errno != 0 {
		return nil, errno
	}

	raw := orig
	// Input: disable break, CR→NL, parity, strip, XON/XOFF
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	// Output: keep OPOST enabled so \n → \r\n translation works
	// Control: set 8-bit chars
	raw.Cflag |= syscall.CS8
	// Local: disable echo, canonical mode, signals, extended input
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	// Read returns after 1 byte, no timeout
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(&raw))); errno != 0 {
		return nil, errno
	}

	return func() {
		syscall.Syscall(syscall.SYS_IOCTL, fd,
			uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(&orig)))
	}, nil
}

// getWinsize returns the terminal width (columns) and height (rows).
func getWinsize(fd uintptr) (cols, rows int, err error) {
	var ws struct {
		Row, Col, Xpixel, Ypixel uint16
	}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws))); errno != 0 {
		return 0, 0, errno
	}
	return int(ws.Col), int(ws.Row), nil
}

// suspendSignal returns the SIGTSTP signal value for darwin.
func suspendSignal() os.Signal {
	return syscall.SIGTSTP
}

// continueSignal returns the SIGCONT signal value for darwin.
func continueSignal() os.Signal {
	return syscall.SIGCONT
}

// winchSignal returns the SIGWINCH signal value for darwin.
func winchSignal() os.Signal {
	return syscall.SIGWINCH
}

// cleanExitSignals returns signals that should trigger a clean exit on darwin.
func cleanExitSignals() []os.Signal {
	return []os.Signal{syscall.SIGTERM, syscall.SIGHUP}
}

// sendSuspendSignal sends SIGTSTP to the current process.
func sendSuspendSignal() {
	syscall.Kill(syscall.Getpid(), syscall.SIGTSTP)
}

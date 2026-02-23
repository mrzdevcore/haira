//go:build linux

package console

import (
	"os"
	"syscall"
	"unsafe"
)

// Linux ioctl constants for terminal control.
const (
	ioctlTCGETS    = 0x5401
	ioctlTCSETS    = 0x5402
	ioctlTIOCGWINSZ = 0x5413
)

// Linux termios flag constants (not all exposed by syscall package).
const (
	linuxICANON = 0x2
	linuxECHO   = 0x8
	linuxISIG   = 0x1
	linuxIEXTEN = 0x8000
	linuxBRKINT = 0x2
	linuxICRNL  = 0x100
	linuxINPCK  = 0x10
	linuxISTRIP = 0x20
	linuxIXON   = 0x400
	linuxCS8    = 0x30
	linuxVMIN   = 6
	linuxVTIME  = 5
)

// linuxTermios mirrors the Linux kernel termios structure.
type linuxTermios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Line   uint8
	Cc     [32]uint8
	_      [3]byte // padding
	Ispeed uint32
	Ospeed uint32
}

// enableRawMode puts the terminal into raw mode and returns a restore function.
// On Linux, uses TCGETS/TCSETS ioctls.
func enableRawMode(fd uintptr) (restore func(), err error) {
	var orig linuxTermios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(ioctlTCGETS), uintptr(unsafe.Pointer(&orig))); errno != 0 {
		return nil, errno
	}

	raw := orig
	raw.Iflag &^= linuxBRKINT | linuxICRNL | linuxINPCK | linuxISTRIP | linuxIXON
	// Keep OPOST enabled so \n → \r\n translation works
	raw.Cflag |= linuxCS8
	raw.Lflag &^= linuxECHO | linuxICANON | linuxISIG | linuxIEXTEN
	raw.Cc[linuxVMIN] = 1
	raw.Cc[linuxVTIME] = 0

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(ioctlTCSETS), uintptr(unsafe.Pointer(&raw))); errno != 0 {
		return nil, errno
	}

	return func() {
		syscall.Syscall(syscall.SYS_IOCTL, fd,
			uintptr(ioctlTCSETS), uintptr(unsafe.Pointer(&orig)))
	}, nil
}

// getWinsize returns the terminal width (columns) and height (rows).
func getWinsize(fd uintptr) (cols, rows int, err error) {
	var ws struct {
		Row, Col, Xpixel, Ypixel uint16
	}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd,
		uintptr(ioctlTIOCGWINSZ), uintptr(unsafe.Pointer(&ws))); errno != 0 {
		return 0, 0, errno
	}
	return int(ws.Col), int(ws.Row), nil
}

// suspendSignal returns the SIGTSTP signal value for linux.
func suspendSignal() os.Signal {
	return syscall.SIGTSTP
}

// continueSignal returns the SIGCONT signal value for linux.
func continueSignal() os.Signal {
	return syscall.SIGCONT
}

// winchSignal returns the SIGWINCH signal value for linux.
func winchSignal() os.Signal {
	return syscall.SIGWINCH
}

// cleanExitSignals returns signals that should trigger a clean exit on linux.
func cleanExitSignals() []os.Signal {
	return []os.Signal{syscall.SIGTERM, syscall.SIGHUP}
}

// sendSuspendSignal sends SIGTSTP to the current process.
func sendSuspendSignal() {
	syscall.Kill(syscall.Getpid(), syscall.SIGTSTP)
}

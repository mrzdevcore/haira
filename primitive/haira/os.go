package haira

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// OsCwd returns the current working directory.
func OsCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

// OsExec executes a shell command and returns its combined output.
func OsExec(command string) (string, error) {
	shell := "/bin/sh"
	flag := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	}
	cmd := exec.Command(shell, flag, command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

// OsExecTimeout executes a shell command with a timeout in seconds.
func OsExecTimeout(command string, timeoutSecs int) (string, error) {
	shell := "/bin/sh"
	flag := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	}
	cmd := exec.Command(shell, flag, command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return "", err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return out.String(), err
	case <-time.After(time.Duration(timeoutSecs) * time.Second):
		cmd.Process.Kill()
		return out.String(), &exec.Error{Name: command, Err: os.ErrDeadlineExceeded}
	}
}

// dangerousCommands is the blocklist for OsSafeExec.
var dangerousCommands = []string{
	"rm -rf /",
	"sudo rm",
	"chmod 777",
	"mkfs",
	"> /dev/",
	"dd if=",
	"format c:",
	":(){:|:&};:",
}

// OsSafeExec executes a shell command after checking it against a safety blocklist.
// Blocks commands that could cause irreversible damage.
func OsSafeExec(command string, timeoutSecs int) (string, error) {
	normalized := strings.ToLower(command)
	for _, blocked := range dangerousCommands {
		if strings.Contains(normalized, blocked) {
			return "", fmt.Errorf("blocked: command contains dangerous pattern %q", blocked)
		}
	}
	if timeoutSecs <= 0 {
		timeoutSecs = 30
	}
	return OsExecTimeout(command, timeoutSecs)
}

// OsArch returns the OS architecture (e.g. "amd64", "arm64").
func OsArch() string {
	return runtime.GOARCH
}

// OsPlatform returns the OS platform (e.g. "darwin", "linux", "windows").
func OsPlatform() string {
	return runtime.GOOS
}

// OsHostname returns the system hostname.
func OsHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

// OsExit terminates the process with the given exit code.
func OsExit(code int) {
	os.Exit(code)
}

// OsArgs returns the command-line arguments (excluding the program name).
func OsArgs() []any {
	args := os.Args[1:]
	result := make([]any, len(args))
	for i, a := range args {
		result[i] = a
	}
	return result
}

// OsGetenv returns the value of an environment variable.
func OsGetenv(key string) string {
	return os.Getenv(key)
}

// OsSetenv sets an environment variable.
func OsSetenv(key string, value string) error {
	return os.Setenv(key, value)
}

// OsEnviron returns all environment variables as a map.
func OsEnviron() map[string]any {
	result := make(map[string]any)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// OsChdir changes the current working directory.
func OsChdir(dir string) error {
	return os.Chdir(dir)
}

// OsTempDir returns the default temp directory.
func OsTempDir() string {
	return os.TempDir()
}

package orchestrator

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// Process represents a running deployed binary.
type Process struct {
	Name       string
	Cmd        *exec.Cmd
	Port       int
	BinaryPath string
	LogFile    *os.File
	StartedAt  time.Time
	cancel     context.CancelFunc
}

// ProcessManager manages child processes for deployed projects.
type ProcessManager struct {
	mu        sync.RWMutex
	processes map[string]*Process
	nextPort  int
	dataDir   string
}

// NewProcessManager creates a new process manager.
func NewProcessManager(dataDir string, startPort int) *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
		nextPort:  startPort,
		dataDir:   dataDir,
	}
}

// Start spawns a deployed binary as a child process.
func (pm *ProcessManager) Start(name, binaryPath string, port int) (*Process, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if p, exists := pm.processes[name]; exists && p.Cmd.Process != nil {
		// Already running — stop it first
		pm.stopLocked(name)
	}

	// Create log directory
	logDir := filepath.Join(pm.dataDir, "logs")
	os.MkdirAll(logDir, 0o755)

	logPath := filepath.Join(logDir, name+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binaryPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(), fmt.Sprintf("HAIRA_PORT=%d", port))

	if err := cmd.Start(); err != nil {
		cancel()
		logFile.Close()
		return nil, fmt.Errorf("start process: %w", err)
	}

	proc := &Process{
		Name:       name,
		Cmd:        cmd,
		Port:       port,
		BinaryPath: binaryPath,
		LogFile:    logFile,
		StartedAt:  time.Now(),
		cancel:     cancel,
	}
	pm.processes[name] = proc

	// Wait in background to reap zombie processes
	go func() {
		cmd.Wait()
		logFile.Close()
	}()

	return proc, nil
}

// Stop terminates a running process gracefully.
func (pm *ProcessManager) Stop(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.stopLocked(name)
}

func (pm *ProcessManager) stopLocked(name string) error {
	proc, exists := pm.processes[name]
	if !exists {
		return nil
	}

	if proc.Cmd.Process == nil {
		delete(pm.processes, name)
		return nil
	}

	// Send SIGTERM
	proc.Cmd.Process.Signal(syscall.SIGTERM)

	// Wait up to 5 seconds for graceful shutdown
	done := make(chan struct{})
	go func() {
		proc.Cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Exited gracefully
	case <-time.After(5 * time.Second):
		// Force kill
		proc.Cmd.Process.Kill()
		<-done
	}

	proc.cancel()
	if proc.LogFile != nil {
		proc.LogFile.Close()
	}
	delete(pm.processes, name)
	return nil
}

// IsRunning checks if a process is alive.
func (pm *ProcessManager) IsRunning(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	proc, exists := pm.processes[name]
	if !exists || proc.Cmd.Process == nil {
		return false
	}
	// Signal 0 checks if process is alive without actually signaling
	return proc.Cmd.Process.Signal(syscall.Signal(0)) == nil
}

// HealthCheck attempts an HTTP request to verify the process is serving.
func (pm *ProcessManager) HealthCheck(name string) bool {
	pm.mu.RLock()
	proc, exists := pm.processes[name]
	pm.mu.RUnlock()

	if !exists {
		return false
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/_api/workflows", proc.Port))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// WaitForHealthy polls until the process passes a health check or times out.
func (pm *ProcessManager) WaitForHealthy(name string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pm.HealthCheck(name) {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

// AllocatePort finds the next available port.
func (pm *ProcessManager) AllocatePort() int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for {
		port := pm.nextPort
		pm.nextPort++

		// Check if port is in use by another process we manage
		inUse := false
		for _, p := range pm.processes {
			if p.Port == port {
				inUse = true
				break
			}
		}
		if inUse {
			continue
		}

		// Check if port is available on the system
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		ln.Close()
		return port
	}
}

// SetNextPort updates the port counter (used when restoring from DB).
func (pm *ProcessManager) SetNextPort(port int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if port >= pm.nextPort {
		pm.nextPort = port + 1
	}
}

// GetProcess returns info about a running process.
func (pm *ProcessManager) GetProcess(name string) *Process {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.processes[name]
}

// StopAll terminates all running processes concurrently.
func (pm *ProcessManager) StopAll() {
	pm.mu.RLock()
	names := make([]string, 0, len(pm.processes))
	for name := range pm.processes {
		names = append(names, name)
	}
	pm.mu.RUnlock()

	var wg sync.WaitGroup
	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			pm.Stop(n)
		}(name)
	}
	wg.Wait()
}

// StartHealthLoop runs periodic health checks on all processes.
func (pm *ProcessManager) StartHealthLoop(store *Store, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			pm.mu.RLock()
			names := make([]string, 0, len(pm.processes))
			for name := range pm.processes {
				names = append(names, name)
			}
			pm.mu.RUnlock()

			for _, name := range names {
				if !pm.IsRunning(name) {
					pm.handleCrash(name, store)
				}
			}
		}
	}()
}

func (pm *ProcessManager) handleCrash(name string, store *Store) {
	d, err := store.GetByName(name)
	if err != nil || d == nil {
		return
	}

	maxRestarts := 5
	if d.Restarts >= maxRestarts {
		d.Status = "crashed"
		store.Update(d)
		return
	}

	// Exponential backoff: 1s, 2s, 4s, 8s, 16s
	backoff := time.Duration(1<<uint(d.Restarts)) * time.Second
	time.Sleep(backoff)

	d.Restarts++
	d.Status = "running"

	proc, err := pm.Start(name, d.BinaryPath, d.Port)
	if err != nil {
		d.Status = "crashed"
		store.Update(d)
		return
	}

	d.PID = proc.Cmd.Process.Pid
	store.Update(d)

	// Verify it actually came up
	if !pm.WaitForHealthy(name, 10*time.Second) {
		fmt.Fprintf(os.Stderr, "[orchestrator] %s failed health check after restart\n", name)
	}
}

// LogPath returns the path to a deployment's log file.
func (pm *ProcessManager) LogPath(name string) string {
	return filepath.Join(pm.dataDir, "logs", name+".log")
}

// PortByName returns the port used by a running process.
func (pm *ProcessManager) PortByName(name string) (int, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.processes[name]
	if !ok {
		return 0, false
	}
	return p.Port, true
}

// PIDByName returns the process ID, or 0 if not running.
func (pm *ProcessManager) PIDByName(name string) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.processes[name]
	if !ok || p.Cmd.Process == nil {
		return 0
	}
	return p.Cmd.Process.Pid
}

// ParsePort parses a port string to int.
func ParsePort(s string) (int, error) {
	return strconv.Atoi(s)
}

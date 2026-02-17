package haira

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WorkflowParam describes a workflow parameter for UI generation.
type WorkflowParam struct {
	Name string
	Type string // "string", "int", "float", "bool", "file"
}

// WorkflowDef defines a workflow that can be exposed as an HTTP endpoint or MCP tool.
type WorkflowDef struct {
	Name          string
	Method        string // HTTP method (GET, POST, etc.)
	Path          string // URL path
	Description   string // optional description (from triple-quoted string, used by MCP)
	Params        []WorkflowParam
	Steps         []string // step names in execution order (for pipeline UI)
	IsStream      bool
	UITitle       string // from @webui(title: "...")
	UIDescription string // from @webui(description: "...")
	ChatEnabled   *bool  // nil = default (on for stream), false = disabled
	Handler       func(params map[string]any) (any, error)
	StreamHandler func(params map[string]any) (<-chan StreamChunk, error)
}

// StepLogEntry represents a single log message within a step.
type StepLogEntry struct {
	Level   string `json:"level"` // "info", "warn", "error"
	Message string `json:"message"`
}

// StepEvent represents a step status change for SSE notification.
type StepEvent struct {
	Name       string        `json:"name"`
	Status     string        `json:"status"` // "start", "end", "failed", "retry", "log"
	DurationMs int64         `json:"duration_ms,omitempty"`
	Error      string        `json:"error,omitempty"`
	Attempt    int           `json:"attempt,omitempty"`
	DelayMs    int           `json:"delay_ms,omitempty"`
	Log        *StepLogEntry `json:"log,omitempty"`
}

// stepNotifiers maps goroutine IDs to step event channels.
var stepNotifiers sync.Map

// SetStepNotifier registers a step event channel for the current goroutine.
func SetStepNotifier(ch chan<- StepEvent) {
	stepNotifiers.Store(goid(), ch)
}

// ClearStepNotifier removes the step notifier for the current goroutine.
func ClearStepNotifier() {
	stepNotifiers.Delete(goid())
}

// notifyStep sends a step event to the registered notifier, if any.
func notifyStep(event StepEvent) {
	if v, ok := stepNotifiers.Load(goid()); ok {
		ch := v.(chan<- StepEvent)
		ch <- event
	}
}

// goid returns the current goroutine's ID.
func goid() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// Stack trace starts with "goroutine NNN ["
	s := string(buf[:n])
	s = strings.TrimPrefix(s, "goroutine ")
	s = s[:strings.IndexByte(s, ' ')]
	id, _ := strconv.ParseUint(s, 10, 64)
	return id
}

// StepStart logs the start of a workflow step and returns the start time.
func StepStart(workflow, step string) time.Time {
	fmt.Printf("[%s] step:start  %q\n", workflow, step)
	notifyStep(StepEvent{Name: step, Status: "start"})
	return time.Now()
}

// StepEnd logs the completion of a workflow step with duration and status.
func StepEnd(workflow, step string, startTime time.Time, err error) {
	duration := time.Since(startTime)
	if err != nil {
		fmt.Printf("[%s] step:end    %q  duration=%s  status=failed  error=%q\n", workflow, step, formatDuration(duration), err.Error())
		notifyStep(StepEvent{Name: step, Status: "failed", DurationMs: duration.Milliseconds(), Error: err.Error()})
	} else {
		fmt.Printf("[%s] step:end    %q  duration=%s  status=success\n", workflow, step, formatDuration(duration))
		notifyStep(StepEvent{Name: step, Status: "end", DurationMs: duration.Milliseconds()})
	}
}

// StepRetry logs a retry attempt for a workflow step.
func StepRetry(workflow, step string, attempt, delayMs int) {
	fmt.Printf("[%s] step:retry  %q  attempt=%d  delay=%dms\n", workflow, step, attempt, delayMs)
	notifyStep(StepEvent{Name: step, Status: "retry", Attempt: attempt, DelayMs: delayMs})
}

// StepLog sends a log message for a workflow step to the UI.
// Level should be "info", "warn", or "error".
func StepLog(workflow, step, level, message string) {
	fmt.Printf("[%s] step:log   %q  level=%s  %s\n", workflow, step, level, message)
	notifyStep(StepEvent{Name: step, Status: "log", Log: &StepLogEntry{Level: level, Message: message}})
}

// LogPrint prints a leveled log message to stdout/stderr (used outside steps).
func LogPrint(level string, message any) {
	prefix := strings.ToUpper(level)
	if level == "error" {
		fmt.Fprintf(os.Stderr, "[%s] %v\n", prefix, message)
	} else {
		fmt.Printf("[%s] %v\n", prefix, message)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

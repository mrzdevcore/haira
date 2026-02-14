package haira

import (
	"fmt"
	"time"
)

// WorkflowDef defines a workflow that can be exposed as an HTTP endpoint.
type WorkflowDef struct {
	Name          string
	Method        string // HTTP method (GET, POST, etc.)
	Path          string // URL path
	Handler       func(params map[string]any) (any, error)
	StreamHandler func(params map[string]any) (<-chan StreamChunk, error)
}

// StepStart logs the start of a workflow step and returns the start time.
func StepStart(workflow, step string) time.Time {
	fmt.Printf("[%s] step:start  %q\n", workflow, step)
	return time.Now()
}

// StepEnd logs the completion of a workflow step with duration and status.
func StepEnd(workflow, step string, startTime time.Time, err error) {
	duration := time.Since(startTime)
	if err != nil {
		fmt.Printf("[%s] step:end    %q  duration=%s  status=failed  error=%q\n", workflow, step, formatDuration(duration), err.Error())
	} else {
		fmt.Printf("[%s] step:end    %q  duration=%s  status=success\n", workflow, step, formatDuration(duration))
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

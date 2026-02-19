package haira

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// RunStatus represents the lifecycle of a workflow run.
type RunStatus string

const (
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
)

// Run captures the full state of a single workflow execution.
type Run struct {
	ID           string         `json:"id"`
	WorkflowName string         `json:"workflow_name"`
	WorkflowPath string         `json:"workflow_path"`
	Status       RunStatus      `json:"status"`
	Params       map[string]any `json:"params,omitempty"`
	Steps        []StepEvent    `json:"steps"`
	Result       any            `json:"result,omitempty"`
	Error        string         `json:"error,omitempty"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   *time.Time     `json:"finished_at,omitempty"`
}

// RunSummary is the compact form returned by the list endpoint.
type RunSummary struct {
	ID           string     `json:"id"`
	WorkflowName string     `json:"workflow_name"`
	WorkflowPath string     `json:"workflow_path"`
	Status       RunStatus  `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	StepCount    int        `json:"step_count"`
}

// ---------------------------------------------------------------------------
// In-memory run tracking + pub/sub for live SSE streaming
// ---------------------------------------------------------------------------

// runSubscribers tracks in-progress runs and their SSE subscribers.
// Persistent storage is handled by the Store interface; this is purely
// for real-time event broadcasting and step buffering during execution.
type runSubscribers struct {
	mu          sync.RWMutex
	runs        map[string]*Run            // in-progress runs only
	subscribers map[string][]chan StepEvent // runID → subscriber channels
}

var globalRunSubs = &runSubscribers{
	runs:        make(map[string]*Run),
	subscribers: make(map[string][]chan StepEvent),
}

var runCounter atomic.Int64

func nextRunID() string {
	n := runCounter.Add(1)
	return fmt.Sprintf("run_%s_%03d", time.Now().Format("20060102_150405"), n)
}

// TrackRun starts tracking a run in memory for live event streaming.
func (s *runSubscribers) TrackRun(run *Run) {
	s.mu.Lock()
	s.runs[run.ID] = run
	s.mu.Unlock()
}

// RecordStepEvent appends a step event to an in-progress run and broadcasts to subscribers.
func (s *runSubscribers) RecordStepEvent(runID string, event StepEvent) {
	s.mu.Lock()
	run := s.runs[runID]
	if run != nil {
		run.Steps = append(run.Steps, event)
	}
	subs := s.subscribers[runID]
	s.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
			// subscriber too slow, skip
		}
	}
}

// FinishRun closes all subscriber channels and removes the run from in-memory tracking.
func (s *runSubscribers) FinishRun(runID string) {
	s.mu.Lock()
	for _, ch := range s.subscribers[runID] {
		close(ch)
	}
	delete(s.subscribers, runID)
	delete(s.runs, runID)
	s.mu.Unlock()
}

// GetSteps returns buffered steps and status for an in-progress run.
// Returns nil, "" if the run is not tracked (already finished).
func (s *runSubscribers) GetSteps(runID string) ([]StepEvent, RunStatus) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	run := s.runs[runID]
	if run == nil {
		return nil, ""
	}
	steps := make([]StepEvent, len(run.Steps))
	copy(steps, run.Steps)
	return steps, run.Status
}

// Subscribe returns a channel that receives live step events for an in-progress run.
func (s *runSubscribers) Subscribe(runID string) chan StepEvent {
	ch := make(chan StepEvent, 32)
	s.mu.Lock()
	s.subscribers[runID] = append(s.subscribers[runID], ch)
	s.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel for a run.
func (s *runSubscribers) Unsubscribe(runID string, ch chan StepEvent) {
	s.mu.Lock()
	subs := s.subscribers[runID]
	for i, sub := range subs {
		if sub == ch {
			s.subscribers[runID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	s.mu.Unlock()
}

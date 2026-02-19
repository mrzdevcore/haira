package haira

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

type runStore struct {
	mu          sync.RWMutex
	runs        []*Run
	subscribers map[string][]chan StepEvent
	maxRuns     int
}

var globalRunStore = &runStore{
	subscribers: make(map[string][]chan StepEvent),
	maxRuns:     50,
}

var runCounter atomic.Int64

func init() {
	if v := os.Getenv("HAIRA_MAX_RUNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			globalRunStore.maxRuns = n
		}
	}
}

func nextRunID() string {
	n := runCounter.Add(1)
	return fmt.Sprintf("run_%s_%03d", time.Now().Format("20060102_150405"), n)
}

// CreateRun starts tracking a new run. Returns the run ID.
func (s *runStore) CreateRun(wfName, wfPath string, params map[string]any) string {
	id := nextRunID()
	run := &Run{
		ID:           id,
		WorkflowName: wfName,
		WorkflowPath: wfPath,
		Status:       RunStatusRunning,
		Params:       params,
		Steps:        []StepEvent{},
		StartedAt:    time.Now(),
	}

	s.mu.Lock()
	// Prepend (newest first)
	s.runs = append([]*Run{run}, s.runs...)
	// Evict old completed runs if over cap
	if len(s.runs) > s.maxRuns {
		s.evictLocked()
	}
	s.mu.Unlock()

	return id
}

// evictLocked removes the oldest completed/failed runs to stay within maxRuns.
// Must be called with s.mu held.
func (s *runStore) evictLocked() {
	for len(s.runs) > s.maxRuns {
		// Find the oldest completed/failed run to remove (search from end)
		removed := false
		for i := len(s.runs) - 1; i >= 0; i-- {
			if s.runs[i].Status != RunStatusRunning {
				s.runs = append(s.runs[:i], s.runs[i+1:]...)
				removed = true
				break
			}
		}
		if !removed {
			break // all runs are still running, can't evict
		}
	}
}

// RecordStepEvent appends a step event to a run and broadcasts to subscribers.
func (s *runStore) RecordStepEvent(runID string, event StepEvent) {
	s.mu.Lock()
	run := s.findLocked(runID)
	if run != nil {
		run.Steps = append(run.Steps, event)
	}
	// Broadcast to subscribers
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

// CompleteRun marks a run as completed with a result.
func (s *runStore) CompleteRun(runID string, result any) {
	now := time.Now()
	s.mu.Lock()
	run := s.findLocked(runID)
	if run != nil {
		run.Status = RunStatusCompleted
		run.Result = result
		run.FinishedAt = &now
	}
	// Close and remove all subscribers
	s.closeSubscribersLocked(runID)
	s.mu.Unlock()

	s.persist()
}

// FailRun marks a run as failed with an error message.
func (s *runStore) FailRun(runID string, errMsg string) {
	now := time.Now()
	s.mu.Lock()
	run := s.findLocked(runID)
	if run != nil {
		run.Status = RunStatusFailed
		run.Error = errMsg
		run.FinishedAt = &now
	}
	s.closeSubscribersLocked(runID)
	s.mu.Unlock()

	s.persist()
}

// ListRuns returns summaries of recent runs, newest first.
// If wfPath is non-empty, filters by workflow path.
func (s *runStore) ListRuns(wfPath string) []RunSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]RunSummary, 0, len(s.runs))
	for _, run := range s.runs {
		if wfPath != "" && run.WorkflowPath != wfPath {
			continue
		}
		result = append(result, RunSummary{
			ID:           run.ID,
			WorkflowName: run.WorkflowName,
			WorkflowPath: run.WorkflowPath,
			Status:       run.Status,
			StartedAt:    run.StartedAt,
			FinishedAt:   run.FinishedAt,
			StepCount:    len(run.Steps),
		})
	}
	return result
}

// GetRun returns the full run by ID, or nil if not found.
func (s *runStore) GetRun(id string) *Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.findLocked(id)
}

// Subscribe returns a channel that receives live step events for an in-progress run.
func (s *runStore) Subscribe(runID string) chan StepEvent {
	ch := make(chan StepEvent, 32)
	s.mu.Lock()
	s.subscribers[runID] = append(s.subscribers[runID], ch)
	s.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel for a run.
func (s *runStore) Unsubscribe(runID string, ch chan StepEvent) {
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

func (s *runStore) findLocked(id string) *Run {
	for _, run := range s.runs {
		if run.ID == id {
			return run
		}
	}
	return nil
}

func (s *runStore) closeSubscribersLocked(runID string) {
	for _, ch := range s.subscribers[runID] {
		close(ch)
	}
	delete(s.subscribers, runID)
}

// --- File persistence ---

const runsFile = ".haira-runs.json"

func (s *runStore) persist() {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.runs, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return
	}
	os.WriteFile(runsFile, data, 0644)
}

func (s *runStore) load() {
	data, err := os.ReadFile(runsFile)
	if err != nil {
		return
	}
	var runs []*Run
	if json.Unmarshal(data, &runs) != nil {
		return
	}
	// Mark interrupted runs as failed
	for _, run := range runs {
		if run.Status == RunStatusRunning {
			now := time.Now()
			run.Status = RunStatusFailed
			run.Error = "server restarted"
			run.FinishedAt = &now
		}
	}
	s.mu.Lock()
	s.runs = runs
	s.mu.Unlock()
}

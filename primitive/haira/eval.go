package haira

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// EvalCase represents a single test case in an agent evaluation.
type EvalCase struct {
	Input    string // The input prompt
	Expected string // Expected behavior or output substring
	Rubric   string // Optional grading rubric for LLM-as-judge
}

// EvalResult holds the result of a single eval case.
type EvalResult struct {
	Input    string
	Expected string
	Output   string
	Pass     bool
	Score    float64 // 0.0 to 1.0
	Duration time.Duration
	Error    string
}

// EvalSummary holds the aggregate results of an eval run.
type EvalSummary struct {
	Name      string
	Agent     string
	Results   []EvalResult
	Passed    int
	Failed    int
	Total     int
	Score     float64 // average score
	Threshold float64
	Duration  time.Duration
}

// RunEval executes an eval suite against an agent.
func RunEval(name string, agent *Agent, cases []EvalCase, threshold float64, parallel bool) *EvalSummary {
	start := time.Now()
	summary := &EvalSummary{
		Name:      name,
		Agent:     agent.config.Name,
		Total:     len(cases),
		Threshold: threshold,
	}

	results := make([]EvalResult, len(cases))

	if parallel {
		var wg sync.WaitGroup
		for i, c := range cases {
			wg.Add(1)
			go func(idx int, tc EvalCase) {
				defer wg.Done()
				results[idx] = runSingleEval(agent, tc)
			}(i, c)
		}
		wg.Wait()
	} else {
		for i, c := range cases {
			results[i] = runSingleEval(agent, c)
		}
	}

	var totalScore float64
	for _, r := range results {
		if r.Pass {
			summary.Passed++
		} else {
			summary.Failed++
		}
		totalScore += r.Score
	}

	summary.Results = results
	if len(results) > 0 {
		summary.Score = totalScore / float64(len(results))
	}
	summary.Duration = time.Since(start)

	return summary
}

// runSingleEval runs a single eval case against an agent.
func runSingleEval(agent *Agent, tc EvalCase) EvalResult {
	start := time.Now()
	result := EvalResult{
		Input:    tc.Input,
		Expected: tc.Expected,
	}

	output, err := agent.Ask(tc.Input, "")
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		result.Score = 0.0
		result.Pass = false
		return result
	}

	result.Output = output

	// Simple substring match scoring
	if tc.Expected != "" {
		lower := strings.ToLower(output)
		expected := strings.ToLower(tc.Expected)
		if strings.Contains(lower, expected) {
			result.Score = 1.0
			result.Pass = true
		} else {
			result.Score = 0.0
			result.Pass = false
		}
	} else {
		// No expected value — just check for non-empty output
		if output != "" {
			result.Score = 1.0
			result.Pass = true
		}
	}

	return result
}

// PrintEvalSummary prints a formatted eval summary to stdout.
func PrintEvalSummary(s *EvalSummary) {
	fmt.Printf("\n=== Eval: %s ===\n", s.Name)
	fmt.Printf("Agent: %s\n", s.Agent)
	fmt.Printf("Duration: %s\n\n", s.Duration.Round(time.Millisecond))

	for i, r := range s.Results {
		status := "PASS"
		if !r.Pass {
			status = "FAIL"
		}
		fmt.Printf("  [%s] Case %d: %s\n", status, i+1, truncate(r.Input, 60))
		if r.Error != "" {
			fmt.Printf("         Error: %s\n", r.Error)
		}
		if !r.Pass && r.Expected != "" {
			fmt.Printf("         Expected: %s\n", truncate(r.Expected, 80))
			fmt.Printf("         Got:      %s\n", truncate(r.Output, 80))
		}
	}

	fmt.Printf("\nScore: %.1f%% (%d/%d passed)\n", s.Score*100, s.Passed, s.Total)
	if s.Score >= s.Threshold {
		fmt.Printf("Result: PASS (threshold: %.0f%%)\n", s.Threshold*100)
	} else {
		fmt.Printf("Result: FAIL (threshold: %.0f%%, got: %.0f%%)\n", s.Threshold*100, s.Score*100)
	}
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

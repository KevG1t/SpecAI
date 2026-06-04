package tui

import (
	"testing"
	"time"

	"github.com/KevG1t/specai/internal/model"
	"github.com/KevG1t/specai/internal/pipeline"
	"github.com/KevG1t/specai/internal/planner"
	"github.com/KevG1t/specai/internal/system"
	tea "github.com/charmbracelet/bubbletea"
)

// runBatchCmd executes every sub-command returned by a tea.Batch so that
// goroutine-backed commands (like the install closure) actually run.
func runBatchCmd(cmd tea.Cmd, results chan<- tea.Msg) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, inner := range batch {
			if inner != nil {
				go func(c tea.Cmd) {
					m := c()
					if m != nil {
						select {
						case results <- m:
						default:
						}
					}
				}(inner)
			}
		}
	} else if msg != nil {
		select {
		case results <- msg:
		default:
		}
	}
}

// TestSendFnNilIsSafeNoOp verifies that a Model with SendFn == nil does not
// panic when startInstalling fires the onProgress callback (SPEC-BUG2-R6,
// Scenario BUG2-S4).
func TestSendFnNilIsSafeNoOp(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")

	// Wire an ExecuteFn that fires onProgress once then returns.
	m.ExecuteFn = func(
		sel model.Selection,
		plan planner.ResolvedPlan,
		det system.DetectionResult,
		onProgress pipeline.ProgressFunc,
	) pipeline.ExecutionResult {
		if onProgress != nil {
			onProgress(pipeline.ProgressEvent{
				StepID: "step-1",
				Status: pipeline.StepStatusRunning,
			})
		}
		return pipeline.ExecutionResult{}
	}

	// SendFn is nil by default — must not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("startInstalling panicked with nil SendFn: %v", r)
		}
	}()

	_, cmd := m.startInstalling()
	results := make(chan tea.Msg, 4)
	runBatchCmd(cmd, results)
	// Wait briefly to let the goroutine complete (must not panic).
	time.Sleep(500 * time.Millisecond)
}

// TestSendFnCapturesStepProgressMsg verifies that a mock SendFn receives the
// correct StepProgressMsg values when the pipeline fires onProgress
// (SPEC-BUG2-R2, SPEC-BUG2-R6, Scenario BUG2-S5).
func TestSendFnCapturesStepProgressMsg(t *testing.T) {
	m := NewModel(system.DetectionResult{}, "dev")

	ch := make(chan tea.Msg, 4)
	m.SendFn = func(msg tea.Msg) {
		ch <- msg
	}

	m.ExecuteFn = func(
		sel model.Selection,
		plan planner.ResolvedPlan,
		det system.DetectionResult,
		onProgress pipeline.ProgressFunc,
	) pipeline.ExecutionResult {
		if onProgress != nil {
			onProgress(pipeline.ProgressEvent{StepID: "step-1", Status: pipeline.StepStatusRunning})
			onProgress(pipeline.ProgressEvent{StepID: "step-1", Status: pipeline.StepStatusSucceeded})
		}
		return pipeline.ExecutionResult{}
	}

	_, cmd := m.startInstalling()
	if cmd == nil {
		t.Fatal("expected a Cmd from startInstalling, got nil")
	}

	results := make(chan tea.Msg, 4)
	runBatchCmd(cmd, results)

	collect := func() []StepProgressMsg {
		var msgs []StepProgressMsg
		deadline := time.After(2 * time.Second)
		for len(msgs) < 2 {
			select {
			case raw := <-ch:
				if pm, ok := raw.(StepProgressMsg); ok {
					msgs = append(msgs, pm)
				}
			case <-deadline:
				return msgs
			}
		}
		return msgs
	}

	msgs := collect()
	if len(msgs) < 2 {
		t.Fatalf("expected 2 StepProgressMsg, got %d", len(msgs))
	}
	if msgs[0].StepID != "step-1" || msgs[0].Status != pipeline.StepStatusRunning {
		t.Errorf("msg[0] = %+v, want StepID=step-1 Status=Running", msgs[0])
	}
	if msgs[1].StepID != "step-1" || msgs[1].Status != pipeline.StepStatusSucceeded {
		t.Errorf("msg[1] = %+v, want StepID=step-1 Status=Succeeded", msgs[1])
	}
}

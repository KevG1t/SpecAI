package planner

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/pipeline"
)

func TestExecutePlan_Progress(t *testing.T) {
	ch := make(chan pipeline.ProgressEvent, 10)
	
	plan := pipeline.StagePlan{
		Apply: []pipeline.Step{dummyStep{"test-step"}},
	}
	
	go ExecutePlan(plan, ch)

	ev := <-ch
	if ev.Status != pipeline.StepStatusRunning {
		t.Errorf("Expected first event to be running, got %s", ev.Status)
	}
}

type dummyStep struct{ id string }
func (s dummyStep) ID() string { return s.id }
func (s dummyStep) Run() error { return nil }

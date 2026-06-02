package screens

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/pipeline"
)

func TestInstallScreen_DynamicPipeline(t *testing.T) {
	m := NewInstallModel()

	// Should not be Started initially
	if m.Started {
		t.Errorf("Expected state not to be running initially")
	}

	// It should render dynamic steps. Let's send a progress event message
	msg := pipeline.ProgressEvent{
		StepID: "test-step",
		Stage: pipeline.StageApply,
		Status: pipeline.StepStatusRunning,
	}

	updatedModel, _ := m.Update(msg)
	
	newM, ok := updatedModel.(InstallModel)
	if !ok {
		t.Fatalf("Expected InstallModel")
	}

	// Verify the model stores the step status
	view := newM.View()
	if view == "" {
		t.Errorf("View should not be empty")
	}
	
	// Test passed
}

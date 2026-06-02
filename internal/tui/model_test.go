package tui

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/screens"
)

func TestTUIAsyncUpdates(t *testing.T) {
	// GIVEN the TUI is on the welcome screen
	m := NewMainModel()

	// Simulate selecting "Install" — should now go to ScreenAgentSelect first.
	msg := screens.OptionSelectedMsg{Option: "Install"}
	newModel, _ := m.Update(msg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenAgentSelect {
		t.Fatalf("expected screen to be ScreenAgentSelect after Install, got %v", m.currentScreen)
	}

	// Simulate agent selection → should move to ScreenInstall
	agentMsg := screens.AgentsSelectedMsg{Agents: []model.AgentID{model.AgentClaudeCode}}
	newModel, _ = m.Update(agentMsg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenInstall {
		t.Fatalf("expected screen to be ScreenInstall after AgentsSelectedMsg, got %v", m.currentScreen)
	}

	// Simulate pipeline progress to ensure it doesn't crash MainModel.
	progressMsg := screens.ProgressMsg{
		TaskName: "Validando OS",
		Status:   "En progreso",
		Progress: 50.0,
	}

	newModel, _ = m.Update(progressMsg)
	m = newModel.(MainModel)

	if m.latestProg.TaskName != "Validando OS" {
		t.Fatalf("expected latestProg.TaskName to be 'Validando OS', got %v", m.latestProg.TaskName)
	}

	// Test PipelineFinishedMsg without payload (legacy behavior — no screen transition).
	doneMsg := screens.PipelineFinishedMsg{Err: nil}
	newModel, _ = m.Update(doneMsg)
	m = newModel.(MainModel)

	// PipelineFinishedMsg without Payload must NOT transition to ScreenComplete.
	if m.currentScreen == ScreenComplete {
		t.Fatalf("PipelineFinishedMsg without Payload must not transition to ScreenComplete")
	}
}

// TestTUIModel_PipelineFinishedMsg_WithPayload_ShowsCompleteScreen verifies that a
// PipelineFinishedMsg carrying a non-nil CompletePayload transitions to ScreenComplete.
func TestTUIModel_PipelineFinishedMsg_WithPayload_ShowsCompleteScreen(t *testing.T) {
	m := NewMainModel()

	payload := &screens.CompletePayload{
		ConfiguredAgents:    2,
		InstalledComponents: 4,
	}
	msg := screens.PipelineFinishedMsg{Err: nil, Payload: payload}
	newModel, _ := m.Update(msg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenComplete {
		t.Fatalf("expected ScreenComplete after PipelineFinishedMsg with Payload, got %v", m.currentScreen)
	}

	view := m.View()
	if view == "" {
		t.Fatal("ScreenComplete view must not be empty")
	}
}


func TestTUIModel_AgentsSelectedMsg_PopulatesIDEs(t *testing.T) {
	m := NewMainModel()

	// Simulate Install option selected → goes to AgentSelect screen
	msg := screens.OptionSelectedMsg{Option: "Install"}
	newModel, _ := m.Update(msg)
	m = newModel.(MainModel)

	// Now simulate AgentsSelectedMsg with 2 agents
	selectedAgents := []model.AgentID{model.AgentClaudeCode, model.AgentCursor}
	agentMsg := screens.AgentsSelectedMsg{Agents: selectedAgents}
	newModel, _ = m.Update(agentMsg)
	m = newModel.(MainModel)

	// ctx.IDEs should have exactly 2 adapters matching the selected agents
	if m.installCtx == nil {
		t.Fatal("installCtx should not be nil after AgentsSelectedMsg")
	}
	if len(m.installCtx.IDEs) != 2 {
		t.Errorf("installCtx.IDEs has %d adapters, want 2", len(m.installCtx.IDEs))
	}
}


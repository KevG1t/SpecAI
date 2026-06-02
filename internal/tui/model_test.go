package tui

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/tui/screens"
)

func TestTUIAsyncUpdates(t *testing.T) {
	// GIVEN the TUI is on the welcome screen
	m := NewMainModel()

	// Simulate selecting "Install" — should now go to ScreenDetection first.
	msg := screens.OptionSelectedMsg{Option: "Install"}
	newModel, _ := m.Update(msg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenDetection {
		t.Fatalf("expected screen to be ScreenDetection after Install, got %v", m.currentScreen)
	}

	// Simulate detection confirmed → should go to ScreenAgentSelect.
	detectionMsg := screens.DetectionConfirmedMsg{Result: &system.DetectionResult{}}
	newModel, _ = m.Update(detectionMsg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenAgentSelect {
		t.Fatalf("expected screen to be ScreenAgentSelect after DetectionConfirmedMsg, got %v", m.currentScreen)
	}

	// Simulate agent selection → should now move to ScreenPersona (new flow).
	agentMsg := screens.AgentsSelectedMsg{Agents: []model.AgentID{model.AgentVSCodeCopilot}}
	newModel, _ = m.Update(agentMsg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenPersona {
		t.Fatalf("expected screen to be ScreenPersona after AgentsSelectedMsg, got %v", m.currentScreen)
	}

	// Simulate persona selection → should move to ScreenPreset.
	personaMsg := screens.PersonaSelectedMsg{Persona: model.PersonaArgentina}
	newModel, _ = m.Update(personaMsg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenPreset {
		t.Fatalf("expected screen to be ScreenPreset after PersonaSelectedMsg, got %v", m.currentScreen)
	}

	// Simulate preset selection → VSCode Copilot has no agent-specific screen, should go to ScreenDependencyTree.
	presetMsg := screens.PresetSelectedMsg{Preset: model.PresetFull}
	newModel, _ = m.Update(presetMsg)
	m = newModel.(MainModel)

	if m.currentScreen != ScreenDependencyTree {
		t.Fatalf("expected screen to be ScreenDependencyTree after PresetSelectedMsg (no agent-specific screen), got %v", m.currentScreen)
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

	// Simulate Install option selected → goes to Detection screen now
	msg := screens.OptionSelectedMsg{Option: "Install"}
	newModel, _ := m.Update(msg)
	m = newModel.(MainModel)

	// Confirm detection → goes to AgentSelect
	detMsg := screens.DetectionConfirmedMsg{Result: &system.DetectionResult{}}
	newModel, _ = m.Update(detMsg)
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

// TestOptionSelectedMsg_InstallGoesToDetection verifies that selecting "Install"
// from the welcome screen transitions to ScreenDetection (not AgentSelect directly).
func TestOptionSelectedMsg_InstallGoesToDetection(t *testing.T) {
	m := NewMainModel()
	newModel, _ := m.Update(screens.OptionSelectedMsg{Option: "Install"})
	m = newModel.(MainModel)

	if m.currentScreen != ScreenDetection {
		t.Errorf("Install option should go to ScreenDetection, got %v", m.currentScreen)
	}
}

// TestNextScreenAfterConfig verifies the routing table for nextScreenAfterConfig.
func TestNextScreenAfterConfig(t *testing.T) {
	tests := []struct {
		name        string
		agents      []model.AgentID
		wantScreen  Screen
	}{
		{
			name:       "ClaudeCode only → ClaudeModelPicker",
			agents:     []model.AgentID{model.AgentClaudeCode},
			wantScreen: ScreenClaudeModelPicker,
		},
		{
			name:       "KiroIDE only → KiroModelPicker",
			agents:     []model.AgentID{model.AgentKiroIDE},
			wantScreen: ScreenKiroModelPicker,
		},
		{
			name:       "OpenCode only → SDDMode",
			agents:     []model.AgentID{model.AgentOpenCode},
			wantScreen: ScreenSDDMode,
		},
		{
			name:       "No SDD agent (VSCodeCopilot) → DependencyTree",
			agents:     []model.AgentID{model.AgentVSCodeCopilot},
			wantScreen: ScreenDependencyTree,
		},
		{
			name:       "No SDD agent (Cursor) → DependencyTree",
			agents:     []model.AgentID{model.AgentCursor},
			wantScreen: ScreenDependencyTree,
		},
		{
			name:       "ClaudeCode + OpenCode → ClaudeModelPicker (Claude takes priority)",
			agents:     []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
			wantScreen: ScreenClaudeModelPicker,
		},
		{
			name:       "KiroIDE + OpenCode → KiroModelPicker (Kiro over OpenCode)",
			agents:     []model.AgentID{model.AgentKiroIDE, model.AgentOpenCode},
			wantScreen: ScreenKiroModelPicker,
		},
		{
			name:       "Empty agents → DependencyTree",
			agents:     []model.AgentID{},
			wantScreen: ScreenDependencyTree,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MainModel{selectedAgents: tt.agents}
			got := m.nextScreenAfterConfig()
			if got != tt.wantScreen {
				t.Errorf("nextScreenAfterConfig() = %v, want %v", got, tt.wantScreen)
			}
		})
	}
}

// TestSetCurrentScreen_SetsPreviousScreen verifies that setCurrentScreen stores
// the current screen as previousScreen before transitioning.
func TestSetCurrentScreen_SetsPreviousScreen(t *testing.T) {
	m := MainModel{currentScreen: ScreenPersona}
	m.setCurrentScreen(ScreenPreset)

	if m.currentScreen != ScreenPreset {
		t.Errorf("currentScreen = %v, want ScreenPreset", m.currentScreen)
	}
	if m.previousScreen != ScreenPersona {
		t.Errorf("previousScreen = %v, want ScreenPersona", m.previousScreen)
	}
}

// TestBackMsg_ReturnsToRealPreviousScreen verifies that BackMsg goes to the actual
// previous screen (not hardcoded ScreenWelcome).
func TestBackMsg_ReturnsToRealPreviousScreen(t *testing.T) {
	// Simulate user going Persona → Preset
	m := MainModel{currentScreen: ScreenWelcome}
	m.setCurrentScreen(ScreenPersona)
	m.setCurrentScreen(ScreenPreset)

	// Now press Back from ScreenPreset
	newModel, _ := m.Update(screens.BackMsg{})
	m = newModel.(MainModel)

	if m.currentScreen != ScreenPersona {
		t.Errorf("after Back from ScreenPreset, currentScreen = %v, want ScreenPersona", m.currentScreen)
	}
}

// TestBackMsg_NoOpWhenPreviousScreenIsZero verifies that Back from the first screen
// does not panic and keeps the current screen.
func TestBackMsg_NoOpWhenPreviousScreenIsZero(t *testing.T) {
	// previousScreen is zero value — no prior navigation
	m := MainModel{currentScreen: ScreenWelcome}

	newModel, _ := m.Update(screens.BackMsg{})
	m = newModel.(MainModel)

	// Should remain on ScreenWelcome, no panic
	if m.currentScreen != ScreenWelcome {
		t.Errorf("BackMsg with zero previousScreen changed screen to %v, want ScreenWelcome", m.currentScreen)
	}
}

// TestNextScreenAfterStrictTDD verifies that nextScreenAfterStrictTDD always
// returns ScreenDependencyTree (ScreenOpenCodePlugins is excluded from this change).
func TestNextScreenAfterStrictTDD(t *testing.T) {
	tests := []struct {
		name   string
		preset model.PresetID
	}{
		{"Full preset → DependencyTree", model.PresetFull},
		{"Custom preset → DependencyTree", model.PresetCustom},
		{"Minimal preset → DependencyTree", model.PresetMinimal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MainModel{selectedPreset: tt.preset}
			got := m.nextScreenAfterStrictTDD()
			if got != ScreenDependencyTree {
				t.Errorf("nextScreenAfterStrictTDD() = %v, want ScreenDependencyTree", got)
			}
		})
	}
}


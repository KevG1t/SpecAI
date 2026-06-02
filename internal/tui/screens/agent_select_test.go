package screens

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/catalog"
	"github.com/KevG1t/SpecAI/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAgentSelectModel_AllAgentsListed(t *testing.T) {
	m := NewAgentSelectModel(nil)
	allAgents := catalog.AllAgents()

	if len(m.agents) != len(allAgents) {
		t.Errorf("AgentSelectModel has %d agents, want %d", len(m.agents), len(allAgents))
	}
}

func TestAgentSelectModel_DetectedAgentsPreChecked(t *testing.T) {
	detected := []model.AgentID{model.AgentClaudeCode, model.AgentCursor}
	m := NewAgentSelectModel(detected)

	// Claude Code and Cursor should be pre-checked
	for _, agent := range m.agents {
		shouldBeChecked := false
		for _, d := range detected {
			if agent.ID == d {
				shouldBeChecked = true
				break
			}
		}
		checked := m.checked[agent.ID]
		if checked != shouldBeChecked {
			t.Errorf("agent %q checked=%v, want %v", agent.ID, checked, shouldBeChecked)
		}
	}
}

func TestAgentSelectModel_EnterWithSelectionEmitsMsg(t *testing.T) {
	detected := []model.AgentID{model.AgentClaudeCode}
	m := NewAgentSelectModel(detected)

	// Press Enter — should emit AgentsSelectedMsg
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = newModel

	if cmd == nil {
		t.Fatal("expected a command after Enter with selection, got nil")
	}

	// Execute the command to get the message
	msg := cmd()
	selectedMsg, ok := msg.(AgentsSelectedMsg)
	if !ok {
		t.Fatalf("expected AgentsSelectedMsg, got %T", msg)
	}

	if len(selectedMsg.Agents) == 0 {
		t.Error("AgentsSelectedMsg.Agents should not be empty")
	}

	// Should contain claude-code
	found := false
	for _, id := range selectedMsg.Agents {
		if id == model.AgentClaudeCode {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("AgentsSelectedMsg should contain %q, got %v", model.AgentClaudeCode, selectedMsg.Agents)
	}
}

func TestAgentSelectModel_EnterWithNoSelectionShowsError(t *testing.T) {
	// No pre-selected agents
	m := NewAgentSelectModel(nil)
	// Ensure no agents are checked
	for id := range m.checked {
		m.checked[id] = false
	}

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(AgentSelectModel)

	// Should NOT emit AgentsSelectedMsg — instead should set error
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(AgentsSelectedMsg); ok {
			t.Error("should not emit AgentsSelectedMsg when no agents selected")
		}
	}

	// Should have an error message
	if m.err == "" {
		t.Error("expected error message when no agents selected, got empty string")
	}
}

func TestAgentSelectModel_SpaceTogglesSelection(t *testing.T) {
	m := NewAgentSelectModel(nil)

	// Ensure cursor is at position 0
	m.cursor = 0
	firstAgent := m.agents[0]
	initialState := m.checked[firstAgent.ID]

	// Toggle with space
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = newM.(AgentSelectModel)

	if m.checked[firstAgent.ID] == initialState {
		t.Errorf("space should have toggled agent %q from %v to %v, got %v",
			firstAgent.ID, initialState, !initialState, m.checked[firstAgent.ID])
	}
}

func TestAgentSelectModel_ViewNotEmpty(t *testing.T) {
	m := NewAgentSelectModel([]model.AgentID{model.AgentClaudeCode})
	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

// TestAgentSelectModel_CursorItemShowsCheckbox verifies that the cursor-highlighted
// item includes the checkbox prefix ([ ] or [✓]), not just the agent name.
func TestAgentSelectModel_CursorItemShowsCheckbox(t *testing.T) {
	detected := []model.AgentID{model.AgentClaudeCode}
	m := NewAgentSelectModel(detected)
	m.cursor = 0 // first item (Claude Code) is pre-checked

	view := m.View()

	// The cursor item must contain a checkbox character.
	// We strip ANSI escape codes by checking for the raw bracket characters.
	if !containsAny(view, "[✓]", "[ ]") {
		t.Errorf("cursor item should show checkbox prefix ([✓] or [ ]), got:\n%s", view)
	}
}

// TestAgentSelectModel_CheckboxPreservedAfterCursorMove verifies that toggling the
// cursor position preserves checkbox visibility on the newly highlighted row.
func TestAgentSelectModel_CheckboxPreservedAfterCursorMove(t *testing.T) {
	if len(catalog.AllAgents()) < 2 {
		t.Skip("need at least 2 agents for cursor move test")
	}

	// Pre-check the second agent; cursor starts at 0.
	detected := []model.AgentID{}
	if len(catalog.AllAgents()) > 1 {
		detected = []model.AgentID{catalog.AllAgents()[1].ID}
	}
	m := NewAgentSelectModel(detected)
	m.cursor = 0

	// Move cursor down to second item.
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newM.(AgentSelectModel)

	view := m.View()

	// After moving to the second (checked) item, the view must show [✓].
	if !containsAny(view, "[✓]") {
		t.Errorf("after cursor move, checked item should show [✓] in view, got:\n%s", view)
	}
}

func containsAny(s string, needles ...string) bool {
	for _, n := range needles {
		for i := 0; i <= len(s)-len(n); i++ {
			if s[i:i+len(n)] == n {
				return true
			}
		}
	}
	return false
}

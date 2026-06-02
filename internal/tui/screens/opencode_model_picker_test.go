package screens

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOpenCodeModelPicker_InitialState(t *testing.T) {
	m := NewOpenCodeModelPickerModel()

	if m.navLevel != ocNavPhaseList {
		t.Errorf("initial navLevel should be ocNavPhaseList, got %d", m.navLevel)
	}

	view := m.View()
	if !strings.Contains(view, "Orchestrator") {
		t.Error("initial view should contain 'Orchestrator' row")
	}
	if !strings.Contains(view, "Set all SDD phases") {
		t.Error("initial view should contain 'Set all SDD phases' row")
	}
}

// advanceTo advances the model by pressing enter repeatedly to navigate
// through phase list → provider select.
func advanceToProviderSelect(m OpenCodeModelPickerModel) OpenCodeModelPickerModel {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return newModel.(OpenCodeModelPickerModel)
}

func TestOpenCodeModelPicker_EnterOnPhase_AdvancesToProviderSelect(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	// cursor is at row 0 (Orchestrator); press enter
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavProviderSelect {
		t.Errorf("expected navLevel==ocNavProviderSelect, got %d", m.navLevel)
	}

	view := m.View()
	// View should list provider names
	for _, p := range m.providers {
		if !strings.Contains(view, p.Name) {
			t.Errorf("provider select view should contain provider %q", p.Name)
		}
	}
}

func TestOpenCodeModelPicker_EnterOnProvider_AdvancesToModelSelect(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	// Enter phase list → provider select
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	// Provider select: cursor is at anthropic (index 0); press enter
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavModelSelect {
		t.Errorf("expected navLevel==ocNavModelSelect, got %d", m.navLevel)
	}

	view := m.View()
	// Should only show anthropic models
	if !strings.Contains(view, "Claude") {
		t.Error("model select view should contain anthropic model names")
	}
}

func TestOpenCodeModelPicker_TextFilter_NarrowsModels(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	// Navigate to model select for anthropic
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // phase → provider
	m = newModel.(OpenCodeModelPickerModel)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // provider → model
	m = newModel.(OpenCodeModelPickerModel)

	// Type "haiku" to filter
	for _, r := range "haiku" {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = newModel.(OpenCodeModelPickerModel)
	}

	filtered := m.filteredModels()
	for _, me := range filtered {
		if !strings.Contains(strings.ToLower(me.Name), "haiku") {
			t.Errorf("filtered model %q should contain 'haiku'", me.Name)
		}
	}

	view := m.View()
	if !strings.Contains(view, "haiku") {
		t.Error("view should show 'haiku' as part of the filter indicator")
	}
	// search field should be set correctly
	if m.search != "haiku" {
		t.Errorf("search field should be 'haiku', got %q", m.search)
	}
}

func TestOpenCodeModelPicker_NonEffortModel_SavesAndReturnsToPhaseList(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	// Navigate to model select for anthropic (claude models don't support effort)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	// Confirm first model (claude-opus-4-8, no effort)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavPhaseList {
		t.Errorf("after non-effort model selection, navLevel should be ocNavPhaseList, got %d", m.navLevel)
	}

	// Assignment should be saved for the selected phase (specai-orchestrator)
	if _, ok := m.assignments["specai-orchestrator"]; !ok {
		t.Error("assignment should be saved for 'specai-orchestrator'")
	}
}

func TestOpenCodeModelPicker_EffortModel_AdvancesToEffortSelect(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	// Navigate to model select for openai (o3 supports effort, index 1 in providers)
	// First find openai provider index
	openaiIdx := -1
	for i, p := range m.providers {
		if p.ID == "openai" {
			openaiIdx = i
			break
		}
	}
	if openaiIdx < 0 {
		t.Fatal("openai provider not found")
	}

	// Phase list → enter
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	// Provider select → navigate to openai
	for m.cursor < openaiIdx {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(OpenCodeModelPickerModel)
	}
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	// Model select: navigate to o3 (find its index)
	o3Idx := -1
	for i, me := range m.providers[openaiIdx].Models {
		if me.ID == "o3" {
			o3Idx = i
			break
		}
	}
	if o3Idx < 0 {
		t.Fatal("o3 model not found in openai provider")
	}

	// At ModelSelect, j/k are search characters; use arrow keys for navigation.
	for m.cursor < o3Idx {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(OpenCodeModelPickerModel)
	}

	// Enter on o3
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavEffortSelect {
		t.Errorf("expected navLevel==ocNavEffortSelect for o3, got %d", m.navLevel)
	}

	view := m.View()
	for _, effort := range effortOptions {
		if !strings.Contains(view, effort) {
			t.Errorf("effort select view should contain %q", effort)
		}
	}
}

func TestOpenCodeModelPicker_EffortSelect_SavesFullAssignment(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	openaiIdx := 1 // openai is at index 1

	// Navigate to EffortSelect for o3
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // phase list → provider select
	m = newModel.(OpenCodeModelPickerModel)
	for m.cursor < openaiIdx {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(OpenCodeModelPickerModel)
	}
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // provider → model select
	m = newModel.(OpenCodeModelPickerModel)

	// Navigate to o3 in model list — at ModelSelect, use arrow keys (j/k are search chars).
	o3Idx := -1
	for i, me := range m.selectedProvider.Models {
		if me.ID == "o3" {
			o3Idx = i
			break
		}
	}
	for m.cursor < o3Idx {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(OpenCodeModelPickerModel)
	}
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // model → effort select
	m = newModel.(OpenCodeModelPickerModel)

	// Select "medium" effort (index 1)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(OpenCodeModelPickerModel)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm medium
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavPhaseList {
		t.Errorf("after effort select, navLevel should be ocNavPhaseList, got %d", m.navLevel)
	}

	a, ok := m.assignments["specai-orchestrator"]
	if !ok {
		t.Fatal("assignment should be saved for 'specai-orchestrator'")
	}
	if a.ModelID != "o3" {
		t.Errorf("assignment model ID should be o3, got %q", a.ModelID)
	}
	if a.Effort != "medium" {
		t.Errorf("assignment effort should be 'medium', got %q", a.Effort)
	}
}

func TestOpenCodeModelPicker_EscAtProviderSelect_ReturnsToPhaseList(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavProviderSelect {
		t.Fatal("expected to be at provider select")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavPhaseList {
		t.Errorf("esc at provider select should return to phase list, got navLevel=%d", m.navLevel)
	}
}

func TestOpenCodeModelPicker_EscAtPhaseList_EmitsSelectedMsg(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc at phase list should produce a cmd")
	}
	msg := cmd()
	if _, ok := msg.(OpenCodeModelsSelectedMsg); !ok {
		t.Fatalf("expected OpenCodeModelsSelectedMsg from esc at phase list, got %T", msg)
	}
}

func TestOpenCodeModelPicker_SetAll_AppliesAssignmentToAllSDDPhases(t *testing.T) {
	m := NewOpenCodeModelPickerModel()

	// Navigate to "Set all SDD phases" (index 1 in phases)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(OpenCodeModelPickerModel)

	if m.cursor != 1 {
		t.Fatalf("cursor should be on Set-all row (index 1), got %d", m.cursor)
	}

	// Enter to advance to provider select
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavProviderSelect {
		t.Fatalf("expected provider select after Set-all enter, got navLevel=%d", m.navLevel)
	}

	// Choose first provider (anthropic)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	// Choose first model (claude-opus-4-8, no effort)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(OpenCodeModelPickerModel)

	if m.navLevel != ocNavPhaseList {
		t.Fatalf("expected phase list after Set-all assignment, got navLevel=%d", m.navLevel)
	}

	// All SDD phase keys should now have an assignment
	for _, key := range orderedSddPhaseKeys {
		if _, ok := m.assignments[key]; !ok {
			t.Errorf("Set-all: assignment missing for phase %q", key)
		}
	}
}

func TestOpenCodeModelPicker_SkipsSeparatorRows(t *testing.T) {
	m := NewOpenCodeModelPickerModel()
	// Find the separator row
	sepIdx := -1
	for i, row := range m.phases {
		if row.IsSeparator {
			sepIdx = i
			break
		}
	}
	if sepIdx < 0 {
		t.Fatal("no separator row found")
	}

	// Navigate cursor to just before separator
	for m.cursor < sepIdx-1 {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(OpenCodeModelPickerModel)
	}

	// One more j should skip over the separator
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(OpenCodeModelPickerModel)

	if m.cursor == sepIdx {
		t.Error("cursor should not land on a separator row")
	}
}

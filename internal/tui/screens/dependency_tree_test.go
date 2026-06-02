package screens

import (
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/planner"
	tea "github.com/charmbracelet/bubbletea"
)

// makeResolvedPlan creates a minimal resolved plan for testing.
func makeResolvedPlan(ordered []model.ComponentID, autoDeps []model.ComponentID) planner.ResolvedPlan {
	return planner.ResolvedPlan{
		OrderedComponents: ordered,
		AddedDependencies: autoDeps,
		Agents:            []model.AgentID{model.AgentClaudeCode},
	}
}

// stubResolver is a test double for the resolver function that can be configured
// to return a specific plan based on which components are selected.
func stubResolver(plan planner.ResolvedPlan) ResolveFunc {
	return func(_ model.Selection) (planner.ResolvedPlan, error) {
		return plan, nil
	}
}

// stubResolverWithSDD returns ComponentSDDMemory as auto-dep when ComponentSDD is selected.
func stubResolverWithSDD() ResolveFunc {
	return func(sel model.Selection) (planner.ResolvedPlan, error) {
		ordered := make([]model.ComponentID, 0, len(sel.Components)+1)
		autoDeps := []model.ComponentID{}

		hasSDD := false
		for _, c := range sel.Components {
			if c == model.ComponentSDD {
				hasSDD = true
			}
		}

		if hasSDD {
			// auto-add SDDMemory
			autoDeps = append(autoDeps, model.ComponentSDDMemory)
			ordered = append(ordered, model.ComponentSDDMemory)
		}
		for _, c := range sel.Components {
			ordered = append(ordered, c)
		}

		return planner.ResolvedPlan{
			OrderedComponents: ordered,
			AddedDependencies: autoDeps,
			Agents:            sel.Agents,
		}, nil
	}
}

func TestDependencyTree_ReadOnly_ViewHasBadges(t *testing.T) {
	resolved := makeResolvedPlan(
		[]model.ComponentID{model.ComponentSDDMemory, model.ComponentSDD},
		[]model.ComponentID{model.ComponentSDDMemory},
	)
	m := NewDependencyTreeModel(model.PresetFull, resolved, nil, nil)
	view := m.View()

	// Should have badges
	if !strings.Contains(view, "[included]") && !strings.Contains(view, "[auto]") {
		t.Error("read-only view should contain badge labels")
	}

	// Should NOT have checkbox characters
	if strings.Contains(view, "[x]") || strings.Contains(view, "[ ]") {
		t.Error("read-only view should not contain checkbox characters")
	}
}

func TestDependencyTree_ReadOnly_EnterEmitsConfirmedMsg(t *testing.T) {
	ordered := []model.ComponentID{model.ComponentSDDMemory, model.ComponentSDD}
	resolved := makeResolvedPlan(ordered, nil)
	m := NewDependencyTreeModel(model.PresetFull, resolved, nil, nil)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a cmd")
	}
	msg := cmd()
	cm, ok := msg.(DependencyTreeConfirmedMsg)
	if !ok {
		t.Fatalf("expected DependencyTreeConfirmedMsg, got %T", msg)
	}
	if len(cm.Components) != len(ordered) {
		t.Errorf("expected %d components, got %d", len(ordered), len(cm.Components))
	}
}

func TestDependencyTree_Custom_ViewHasCheckboxes(t *testing.T) {
	resolved := planner.ResolvedPlan{}
	m := NewDependencyTreeModel(model.PresetCustom, resolved, nil, stubResolver(resolved))
	view := m.View()

	if !strings.Contains(view, "[ ]") && !strings.Contains(view, "[x]") {
		t.Error("custom mode view should contain checkbox characters")
	}
}

func TestDependencyTree_Custom_ToggleSDD_AutoAddsMemory(t *testing.T) {
	resolved := planner.ResolvedPlan{}
	agents := []model.AgentID{model.AgentOpenCode}
	m := NewDependencyTreeModel(model.PresetCustom, resolved, agents, stubResolverWithSDD())

	// Find the index of ComponentSDD in items
	sddIdx := -1
	for i, e := range m.components {
		if e.ID == model.ComponentSDD {
			sddIdx = i
			break
		}
	}
	if sddIdx < 0 {
		t.Fatal("ComponentSDD not found in custom mode items")
	}

	// Navigate to SDD
	for m.cursor < sddIdx {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(DependencyTreeModel)
	}

	// Toggle SDD on (it starts unchecked in custom mode)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = newModel.(DependencyTreeModel)

	// SDDMemory should now be auto-dependency
	memoryFound := false
	for _, e := range m.components {
		if e.ID == model.ComponentSDDMemory {
			if e.Badge == "auto" && e.Checked {
				memoryFound = true
			}
			break
		}
	}
	if !memoryFound {
		t.Error("ComponentSDDMemory should be marked as auto-dependency after ComponentSDD is toggled on")
	}

	view := m.View()
	if !strings.Contains(view, "[auto]") {
		t.Error("view should contain '[auto]' badge after re-resolve")
	}
}

func TestDependencyTree_Custom_AutoDepCannotBeUnchecked(t *testing.T) {
	resolved := planner.ResolvedPlan{}
	agents := []model.AgentID{model.AgentOpenCode}
	m := NewDependencyTreeModel(model.PresetCustom, resolved, agents, stubResolverWithSDD())

	// Toggle SDD on so SDDMemory becomes auto
	sddIdx := -1
	for i, e := range m.components {
		if e.ID == model.ComponentSDD {
			sddIdx = i
			break
		}
	}
	for m.cursor < sddIdx {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(DependencyTreeModel)
	}
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = newModel.(DependencyTreeModel)

	// Find SDDMemory index
	memIdx := -1
	for i, e := range m.components {
		if e.ID == model.ComponentSDDMemory {
			memIdx = i
			break
		}
	}
	if memIdx < 0 {
		t.Fatal("ComponentSDDMemory not found")
	}

	// Navigate to SDDMemory
	for m.cursor != memIdx {
		if m.cursor < memIdx {
			newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		} else {
			newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		}
		m = newModel.(DependencyTreeModel)
	}

	// Attempt to toggle the auto-dep — should be a no-op
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = newModel.(DependencyTreeModel)

	for _, e := range m.components {
		if e.ID == model.ComponentSDDMemory {
			if !e.Checked || e.Badge != "auto" {
				t.Error("auto-dependency should remain checked and marked auto after toggle attempt")
			}
			return
		}
	}
	t.Error("ComponentSDDMemory not found after navigation")
}

func TestDependencyTree_EscEmitsBackMsg(t *testing.T) {
	resolved := planner.ResolvedPlan{}
	m := NewDependencyTreeModel(model.PresetCustom, resolved, nil, stubResolver(resolved))

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc should produce a cmd")
	}
	msg := cmd()
	if _, ok := msg.(BackMsg); !ok {
		t.Fatalf("expected BackMsg from esc, got %T", msg)
	}
}

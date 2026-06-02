package screens

import (
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/planner"
	tea "github.com/charmbracelet/bubbletea"
)

func makeFullReviewPayload() planner.ReviewPayload {
	return planner.ReviewPayload{
		Agents:  []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
		Persona: model.PersonaArgentina,
		Preset:  model.PresetFull,
		Components: []planner.ComponentAction{
			{ID: model.ComponentSDDMemory, Action: "selected"},
			{ID: model.ComponentSDD, Action: "selected"},
			{ID: model.ComponentPersona, Action: "auto-dependency"},
		},
		AddedDependencies: []model.ComponentID{model.ComponentPersona},
		Skills:            []model.SkillID{model.SkillSDDApply, model.SkillGoTesting},
		StrictTDD:         true,
		HasSDD:            true,
	}
}

func TestReview_ViewRendersAllSections(t *testing.T) {
	m := NewReviewModel(makeFullReviewPayload())
	view := m.View()

	checks := []struct {
		name  string
		token string
	}{
		{"agent claude-code", "claude-code"},
		{"agent opencode", "opencode"},
		{"persona argentina", "argentina"},
		{"preset full", "full"},
		{"component sdd-memory", "sdd-memory"},
		{"skill sdd-apply", "sdd-apply"},
		{"strict TDD heading", "Strict TDD"},
		{"Install action", "Install"},
		{"Back action", "Back"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(view, c.token) {
				t.Errorf("view does not contain %q", c.token)
			}
		})
	}
}

func TestReview_AutoDependencyBadgeShown(t *testing.T) {
	m := NewReviewModel(makeFullReviewPayload())
	view := m.View()

	// persona is in AddedDependencies — should show auto-dependency badge
	if !strings.Contains(view, "auto-dependency") {
		t.Error("view should contain 'auto-dependency' badge for auto-added components")
	}
}

func TestReview_EnterOnInstall_EmitsConfirmedMsg(t *testing.T) {
	m := NewReviewModel(makeFullReviewPayload())
	// cursor defaults to 0 (Install)

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = newModel

	if cmd == nil {
		t.Fatal("expected a cmd from enter on Install")
	}
	msg := cmd()
	if _, ok := msg.(ReviewConfirmedMsg); !ok {
		t.Fatalf("expected ReviewConfirmedMsg, got %T", msg)
	}
}

func TestReview_EnterOnBack_EmitsBackMsg(t *testing.T) {
	m := NewReviewModel(makeFullReviewPayload())
	// move cursor to Back (index 1)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(ReviewModel)

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = newModel

	if cmd == nil {
		t.Fatal("expected a cmd from enter on Back")
	}
	msg := cmd()
	if _, ok := msg.(ReviewBackMsg); !ok {
		t.Fatalf("expected ReviewBackMsg, got %T", msg)
	}
}

func TestReview_EscEmitsBackMsg(t *testing.T) {
	m := NewReviewModel(makeFullReviewPayload())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected a cmd from esc")
	}
	msg := cmd()
	if _, ok := msg.(ReviewBackMsg); !ok {
		t.Fatalf("expected ReviewBackMsg from esc, got %T", msg)
	}
}

func TestReview_JKCursorMovement(t *testing.T) {
	m := NewReviewModel(makeFullReviewPayload())

	if m.cursor != 0 {
		t.Fatalf("initial cursor should be 0, got %d", m.cursor)
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(ReviewModel)
	if m.cursor != 1 {
		t.Fatalf("cursor after j should be 1, got %d", m.cursor)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = newModel.(ReviewModel)
	if m.cursor != 0 {
		t.Fatalf("cursor after k should be 0, got %d", m.cursor)
	}
}

// TestReview_SDDModeRendered verifies that SDDMode is shown when HasSDD is true
// and SDDMode is set.
func TestReview_SDDModeRendered(t *testing.T) {
	payload := makeFullReviewPayload()
	payload.SDDMode = "multi"
	m := NewReviewModel(payload)
	view := m.View()

	if !strings.Contains(view, "SDD Mode: multi") {
		t.Errorf("view should contain 'SDD Mode: multi', got:\n%s", view)
	}
}

// TestReview_SDDModeNotRenderedWhenHasSDDFalse verifies SDDMode line is absent
// when HasSDD is false.
func TestReview_SDDModeNotRenderedWhenHasSDDFalse(t *testing.T) {
	payload := makeFullReviewPayload()
	payload.HasSDD = false
	payload.SDDMode = "multi"
	m := NewReviewModel(payload)
	view := m.View()

	if strings.Contains(view, "SDD Mode: multi") {
		t.Errorf("view should NOT contain 'SDD Mode: multi' when HasSDD=false")
	}
}

func TestReview_CursorDoesNotWrapBeyondBounds(t *testing.T) {
	m := NewReviewModel(makeFullReviewPayload())

	// cursor is 0; pressing k should stay at 0
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = newModel.(ReviewModel)
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 after k from 0, got %d", m.cursor)
	}

	// move to 1; pressing j should stay at 1
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(ReviewModel)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(ReviewModel)
	if m.cursor != 1 {
		t.Errorf("cursor should stay at 1 after j from 1, got %d", m.cursor)
	}
}

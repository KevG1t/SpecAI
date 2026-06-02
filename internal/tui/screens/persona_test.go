package screens

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

// TestPersona_JKeySync_SelectedFollowsCursor verifies that pressing j moves
// m.selected to personaOptions[m.cursor] so the radio marker tracks cursor position.
func TestPersona_JKeySync_SelectedFollowsCursor(t *testing.T) {
	m := NewPersonaModel()
	// Initial state: cursor=0, selected=PersonaArgentina (options[0])
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}
	if m.selected != model.PersonaArgentina {
		t.Fatalf("initial selected = %q, want PersonaArgentina", m.selected)
	}

	// Press j — cursor moves to 1, selected must become options[1]=PersonaNeutral
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(PersonaModel)

	if m.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m.cursor)
	}
	if m.selected != personaOptions[1] {
		t.Errorf("selected after j = %q, want %q", m.selected, personaOptions[1])
	}
}

// TestPersona_KKeySync_SelectedFollowsCursor verifies that pressing k moves
// selected backwards to personaOptions[m.cursor].
func TestPersona_KKeySync_SelectedFollowsCursor(t *testing.T) {
	m := NewPersonaModel()
	// Move to index 2 first
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(PersonaModel)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(PersonaModel)

	if m.cursor != 2 {
		t.Fatalf("cursor after 2x j = %d, want 2", m.cursor)
	}

	// Press k — cursor moves to 1, selected must become options[1]
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = newModel.(PersonaModel)

	if m.cursor != 1 {
		t.Errorf("cursor after k = %d, want 1", m.cursor)
	}
	if m.selected != personaOptions[1] {
		t.Errorf("selected after k = %q, want %q", m.selected, personaOptions[1])
	}
}

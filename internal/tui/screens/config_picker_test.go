package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigPickerModelInit(t *testing.T) {
	m := NewConfigPickerModel()
	if len(m.fields) == 0 {
		t.Fatal("ConfigPickerModel must have at least one field")
	}
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}
}

func TestConfigPickerModelDefaultValues(t *testing.T) {
	m := NewConfigPickerModel()

	// Theme field.
	theme := m.fields[0].Options[m.fields[0].Selected]
	if theme == "" {
		t.Fatal("Theme field must have a non-empty default value")
	}

	// PermissionsLevel field.
	perms := m.fields[1].Options[m.fields[1].Selected]
	if perms == "" {
		t.Fatal("PermissionsLevel field must have a non-empty default value")
	}

	// EditorMode field.
	editor := m.fields[2].Options[m.fields[2].Selected]
	if editor == "" {
		t.Fatal("EditorMode field must have a non-empty default value")
	}
}

func TestConfigPickerModelCursorNavigation(t *testing.T) {
	m := NewConfigPickerModel()
	m.cursor = 0

	// Move down.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(ConfigPickerModel)
	if m.cursor != 1 {
		t.Fatalf("cursor after j = %d, want 1", m.cursor)
	}

	// Move up.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(ConfigPickerModel)
	if m.cursor != 0 {
		t.Fatalf("cursor after k = %d, want 0", m.cursor)
	}
}

func TestConfigPickerModelOptionCycling(t *testing.T) {
	m := NewConfigPickerModel()
	m.cursor = 0 // Theme field

	initialSelected := m.fields[0].Selected
	totalOptions := len(m.fields[0].Options)

	// Cycle forward with 'l'.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(ConfigPickerModel)

	expected := (initialSelected + 1) % totalOptions
	if m.fields[0].Selected != expected {
		t.Fatalf("after l: Selected = %d, want %d", m.fields[0].Selected, expected)
	}

	// Cycle backward with 'h'.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = updated.(ConfigPickerModel)

	if m.fields[0].Selected != initialSelected {
		t.Fatalf("after h: Selected = %d, want %d (back to original)", m.fields[0].Selected, initialSelected)
	}
}

func TestConfigPickerModelConfirmEmitsConfigSelectedMsg(t *testing.T) {
	m := NewConfigPickerModel()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter key must return a non-nil Cmd")
	}

	msg := cmd()
	configMsg, ok := msg.(ConfigSelectedMsg)
	if !ok {
		t.Fatalf("expected ConfigSelectedMsg, got %T", msg)
	}

	// Default values must be non-empty.
	if configMsg.Theme == "" {
		t.Fatal("ConfigSelectedMsg.Theme must not be empty")
	}
	if configMsg.PermissionsLevel == "" {
		t.Fatal("ConfigSelectedMsg.PermissionsLevel must not be empty")
	}
	if configMsg.EditorMode == "" {
		t.Fatal("ConfigSelectedMsg.EditorMode must not be empty")
	}
}

func TestConfigPickerModelModifiedFieldsPropagateToMsg(t *testing.T) {
	m := NewConfigPickerModel()

	// Cycle Theme field once (from index 0 to 1 = "kanagawa").
	m.cursor = 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = updated.(ConfigPickerModel)

	expectedTheme := m.fields[0].Options[1] // After one forward cycle.

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()

	configMsg, ok := msg.(ConfigSelectedMsg)
	if !ok {
		t.Fatalf("expected ConfigSelectedMsg, got %T", msg)
	}

	if configMsg.Theme != expectedTheme {
		t.Fatalf("ConfigSelectedMsg.Theme = %q, want %q", configMsg.Theme, expectedTheme)
	}
}

func TestConfigPickerModelBackEmitsBackMsg(t *testing.T) {
	m := NewConfigPickerModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc key must return a non-nil Cmd")
	}

	msg := cmd()
	if _, ok := msg.(BackMsg); !ok {
		t.Fatalf("expected BackMsg from esc, got %T", msg)
	}
}

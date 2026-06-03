package screens

import (
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMCPPickerModelInit(t *testing.T) {
	m := NewMCPPickerModel()
	if len(m.items) == 0 {
		t.Fatal("MCPPickerModel must have at least one item")
	}
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}
}

func TestMCPPickerModelContext7StartsChecked(t *testing.T) {
	m := NewMCPPickerModel()
	found := false
	for _, it := range m.items {
		if it.ID == model.ComponentContext7 {
			found = true
			if !it.Checked {
				t.Fatal("Context7 must start checked by default")
			}
		}
	}
	if !found {
		t.Fatal("Context7 not found in MCPPickerModel items")
	}
}

func TestMCPPickerModelToggle(t *testing.T) {
	m := NewMCPPickerModel()
	// Find Notion index.
	notionIdx := -1
	for i, it := range m.items {
		if it.ID == model.ComponentNotion {
			notionIdx = i
			break
		}
	}
	if notionIdx < 0 {
		t.Fatal("Notion not found in MCP items")
	}

	// Move cursor to Notion.
	m.cursor = notionIdx
	initialChecked := m.items[notionIdx].Checked

	// Toggle via space key.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(MCPPickerModel)

	if m.items[notionIdx].Checked == initialChecked {
		t.Fatal("space key must toggle the focused item")
	}
}

func TestMCPPickerModelCursorNavigation(t *testing.T) {
	m := NewMCPPickerModel()
	m.cursor = 0

	// Move down.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(MCPPickerModel)
	if m.cursor != 1 {
		t.Fatalf("cursor after j = %d, want 1", m.cursor)
	}

	// Move up.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(MCPPickerModel)
	if m.cursor != 0 {
		t.Fatalf("cursor after k = %d, want 0", m.cursor)
	}

	// Cannot go above 0.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(MCPPickerModel)
	if m.cursor != 0 {
		t.Fatalf("cursor should not go below 0, got %d", m.cursor)
	}
}

func TestMCPPickerModelConfirmEmitsMCPServersSelectedMsg(t *testing.T) {
	m := NewMCPPickerModel()
	// Context7 starts checked; confirm should emit MCPServersSelectedMsg.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter key must return a non-nil Cmd")
	}

	msg := cmd()
	selectedMsg, ok := msg.(MCPServersSelectedMsg)
	if !ok {
		t.Fatalf("expected MCPServersSelectedMsg, got %T", msg)
	}

	// At least Context7 should be in the result (it starts checked).
	found := false
	for _, id := range selectedMsg.ComponentIDs {
		if id == model.ComponentContext7 {
			found = true
		}
	}
	if !found {
		t.Fatal("MCPServersSelectedMsg should include Context7 (checked by default)")
	}
}

func TestMCPPickerModelBackEmitsBackMsg(t *testing.T) {
	m := NewMCPPickerModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc key must return a non-nil Cmd")
	}

	msg := cmd()
	if _, ok := msg.(BackMsg); !ok {
		t.Fatalf("expected BackMsg from esc, got %T", msg)
	}
}

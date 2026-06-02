package screens

import (
	"strings"
	"testing"

	"github.com/KevG1t/SpecAI/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSkillPicker_ViewShowsAllSkillIDs(t *testing.T) {
	m := NewSkillPickerModel(nil) // nil => all skills checked
	view := m.View()

	for _, item := range allSkills {
		if !strings.Contains(view, string(item.ID)) {
			t.Errorf("view should contain skill ID %q", item.ID)
		}
	}
}

func TestSkillPicker_PreSelectedSkillsAreChecked(t *testing.T) {
	pre := []model.SkillID{model.SkillSDDApply, model.SkillGoTesting}
	m := NewSkillPickerModel(pre)

	checkedCount := 0
	for _, it := range m.items {
		if it.Checked {
			checkedCount++
			found := false
			for _, pid := range pre {
				if it.ID == pid {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("item %q is checked but was not in preSelected", it.ID)
			}
		}
	}
	if checkedCount != len(pre) {
		t.Errorf("expected %d checked items, got %d", len(pre), checkedCount)
	}
}

func TestSkillPicker_SpaceToggleCheck(t *testing.T) {
	m := NewSkillPickerModel(nil) // all checked
	// cursor is 0; toggling with 2+ skills checked should uncheck
	initialChecked := m.items[0].Checked
	if !initialChecked {
		t.Fatal("item 0 should start checked (nil preSelected)")
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = newModel.(SkillPickerModel)

	if m.items[0].Checked {
		t.Error("item 0 should be unchecked after space toggle")
	}

	// Toggle back on
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = newModel.(SkillPickerModel)
	if !m.items[0].Checked {
		t.Error("item 0 should be checked again after second space toggle")
	}
}

func TestSkillPicker_LastSkillGuard(t *testing.T) {
	// Create picker with exactly one checked skill.
	m := NewSkillPickerModel([]model.SkillID{model.SkillSDDApply})

	// cursor is 0, which maps to the first skill (SkillSDDInit), not SkillSDDApply
	// We need to navigate to the checked item.
	// Find index of SkillSDDApply
	applyIdx := -1
	for i, it := range m.items {
		if it.ID == model.SkillSDDApply {
			applyIdx = i
			break
		}
	}
	if applyIdx < 0 {
		t.Fatal("SkillSDDApply not found in items")
	}

	// Move cursor to applyIdx
	for i := 0; i < applyIdx; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(SkillPickerModel)
	}

	// Attempt to uncheck the only checked skill — should be a no-op
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = newModel.(SkillPickerModel)

	checkedCount := 0
	for _, it := range m.items {
		if it.Checked {
			checkedCount++
		}
	}
	if checkedCount != 1 {
		t.Errorf("last-skill guard failed: expected 1 checked skill, got %d", checkedCount)
	}
}

func TestSkillPicker_EnterEmitsSkillsSelectedMsg(t *testing.T) {
	pre := []model.SkillID{model.SkillSDDApply, model.SkillGoTesting}
	m := NewSkillPickerModel(pre)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a cmd")
	}
	msg := cmd()
	sm, ok := msg.(SkillsSelectedMsg)
	if !ok {
		t.Fatalf("expected SkillsSelectedMsg, got %T", msg)
	}

	if len(sm.Skills) != len(pre) {
		t.Errorf("expected %d skills, got %d", len(pre), len(sm.Skills))
	}
	for _, expected := range pre {
		found := false
		for _, got := range sm.Skills {
			if got == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected skill %q in result", expected)
		}
	}
}

func TestSkillPicker_EscEmitsBackMsg(t *testing.T) {
	m := NewSkillPickerModel(nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc should produce a cmd")
	}
	msg := cmd()
	if _, ok := msg.(BackMsg); !ok {
		t.Fatalf("expected BackMsg from esc, got %T", msg)
	}
}

func TestSkillPicker_CursorMovement(t *testing.T) {
	m := NewSkillPickerModel(nil)
	if m.cursor != 0 {
		t.Fatalf("initial cursor should be 0, got %d", m.cursor)
	}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(SkillPickerModel)
	if m.cursor != 1 {
		t.Errorf("cursor after j should be 1, got %d", m.cursor)
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = newModel.(SkillPickerModel)
	if m.cursor != 0 {
		t.Errorf("cursor after k should be 0, got %d", m.cursor)
	}
}

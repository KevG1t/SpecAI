package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// KiroModelsSelectedMsg is emitted when the user confirms Kiro model assignments.
type KiroModelsSelectedMsg struct {
	Assignments map[string]model.ClaudeModelAlias
}

// KiroModelPickerState reuses the same phase-assignment mechanics as Claude
// aliases (opus|sonnet|haiku), but remains a separate UI flow and persisted map.
type KiroModelPickerState struct {
	Preset            ClaudeModelPreset
	CustomAssignments map[string]model.ClaudeModelAlias
	InCustomMode      bool
}

func NewKiroModelPickerState() KiroModelPickerState {
	return KiroModelPickerState{
		Preset:            ClaudePresetBalanced,
		CustomAssignments: model.ClaudeModelPresetBalanced(),
	}
}

func NewKiroModelPickerStateFromAssignments(assignments map[string]model.ClaudeModelAlias) KiroModelPickerState {
	if len(assignments) == 0 {
		return NewKiroModelPickerState()
	}
	for preset, constructor := range claudePresetConstructors {
		if claudeAssignmentsEqual(constructor(), assignments) {
			return KiroModelPickerState{
				Preset:            preset,
				CustomAssignments: claudeCopyAssignments(assignments),
			}
		}
	}
	return KiroModelPickerState{
		Preset:            ClaudePresetCustom,
		CustomAssignments: claudeCopyAssignments(assignments),
	}
}

// KiroModelPickerModel is a standalone BubbleTea model for Kiro model picker.
type KiroModelPickerModel struct {
	state  KiroModelPickerState
	cursor int
}

func NewKiroModelPickerModel(existing map[string]model.ClaudeModelAlias) KiroModelPickerModel {
	return KiroModelPickerModel{
		state: NewKiroModelPickerStateFromAssignments(existing),
	}
}

func (m KiroModelPickerModel) Init() tea.Cmd { return nil }

func (m KiroModelPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		optCount := kiroPickerOptionCount(m.state)

		switch key {
		case "j", "down":
			if m.cursor < optCount-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "esc":
			if m.state.InCustomMode {
				m.state.InCustomMode = false
				m.cursor = 0
				return m, nil
			}
			return m, func() tea.Msg { return BackMsg{} }
		case "enter":
			handled, assignments := HandleKiroModelPickerNav(key, &m.state, m.cursor)
			if !handled {
				return m, func() tea.Msg { return BackMsg{} }
			}
			if assignments != nil {
				return m, func() tea.Msg { return KiroModelsSelectedMsg{Assignments: assignments} }
			}
			m.cursor = 0
		}
	}
	return m, nil
}

func (m KiroModelPickerModel) View() string {
	return RenderKiroModelPicker(m.state, m.cursor)
}

// HandleKiroModelPickerNav processes navigation for the Kiro model picker.
func HandleKiroModelPickerNav(
	key string,
	state *KiroModelPickerState,
	cursor int,
) (handled bool, assignments map[string]model.ClaudeModelAlias) {
	bridge := ClaudeModelPickerState{
		Preset:            state.Preset,
		CustomAssignments: state.CustomAssignments,
		InCustomMode:      state.InCustomMode,
	}
	handled, assignments = HandleClaudeModelPickerNav(key, &bridge, cursor)
	state.Preset = bridge.Preset
	state.CustomAssignments = bridge.CustomAssignments
	state.InCustomMode = bridge.InCustomMode
	return handled, assignments
}

func kiroPickerOptionCount(state KiroModelPickerState) int {
	if state.InCustomMode {
		return len(claudePhases) + 2
	}
	return len(claudePresetOrder) + 1
}

// RenderKiroModelPicker renders the Kiro model picker screen.
func RenderKiroModelPicker(state KiroModelPickerState, cursor int) string {
	if state.InCustomMode {
		return renderKiroCustomPhaseList(state, cursor)
	}
	return renderKiroPresetList(state, cursor)
}

func renderKiroPresetList(state KiroModelPickerState, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Kiro Model Assignments"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Choose how Kiro models are assigned to each SDD execution phase (explore → apply → archive):"))
	b.WriteString("\n\n")

	for idx, preset := range claudePresetOrder {
		isSelected := preset == state.Preset
		focused := idx == cursor
		b.WriteString(renderRadio(string(preset), isSelected, focused))
		b.WriteString(styles.SubtextStyle.Render("    "+claudePresetDescriptions[preset]) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"← Back"}, cursor-len(claudePresetOrder)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

func renderKiroCustomPhaseList(state KiroModelPickerState, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Custom Kiro Model Assignments"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Press enter on a phase to cycle: opus → sonnet → haiku"))
	b.WriteString("\n\n")

	for idx, phase := range claudePhases {
		focused := idx == cursor
		alias := state.CustomAssignments[phase]
		if alias == "" {
			alias = model.ClaudeModelSonnet
		}

		label := fmt.Sprintf("%-20s %s", claudePhaseLabels[phase], claudeAliasTag(alias))

		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+label) + "\n")
		}
	}

	b.WriteString("\n")
	actionCursor := cursor - len(claudePhases)
	b.WriteString(renderOptions([]string{"Confirm", "← Back"}, actionCursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: cycle/select • esc: back"))

	return b.String()
}

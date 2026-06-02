package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// ClaudeModelsSelectedMsg is emitted when the user confirms Claude model assignments.
type ClaudeModelsSelectedMsg struct {
	Assignments map[string]model.ClaudeModelAlias
}

// ClaudeModelPreset represents a named preset for Claude model assignments.
type ClaudeModelPreset string

const (
	ClaudePresetBalanced    ClaudeModelPreset = "balanced"
	ClaudePresetPerformance ClaudeModelPreset = "performance"
	ClaudePresetEconomy     ClaudeModelPreset = "economy"
	ClaudePresetDiversity   ClaudeModelPreset = "diversity"
	ClaudePresetCustom      ClaudeModelPreset = "custom"
)

var claudePresetDescriptions = map[ClaudeModelPreset]string{
	ClaudePresetBalanced:    "Smart defaults: opus for architecture, sonnet for most phases, haiku for archiving",
	ClaudePresetPerformance: "Maximum quality: opus for architecture, planning & verification phases",
	ClaudePresetEconomy:     "Cost-optimised: sonnet for all phases, haiku for archiving",
	ClaudePresetDiversity:   "Diversity: Opus for Judge A, Haiku for Judge B, Sonnet for fixes",
	ClaudePresetCustom:      "Pick the model alias for each SDD phase, JD agent, and general delegation entry individually",
}

var claudePresetOrder = []ClaudeModelPreset{
	ClaudePresetBalanced,
	ClaudePresetPerformance,
	ClaudePresetEconomy,
	ClaudePresetDiversity,
	ClaudePresetCustom,
}

var claudePhases = []string{
	"sdd-explore",
	"sdd-propose",
	"sdd-spec",
	"sdd-design",
	"sdd-tasks",
	"sdd-apply",
	"sdd-verify",
	"sdd-archive",
	"sdd-onboard",
	"jd-judge-a",
	"jd-judge-b",
	"jd-fix-agent",
	"default",
}

var claudePhaseLabels = map[string]string{
	"sdd-explore":  "Explore",
	"sdd-propose":  "Propose",
	"sdd-spec":     "Spec",
	"sdd-design":   "Design",
	"sdd-tasks":    "Tasks",
	"sdd-apply":    "Apply",
	"sdd-verify":   "Verify",
	"sdd-archive":  "Archive",
	"sdd-onboard":  "Onboard",
	"jd-judge-a":   "JD Judge A",
	"jd-judge-b":   "JD Judge B",
	"jd-fix-agent": "JD Fix Agent",
	"default":      "General delegation",
}

var claudeAliasOrder = []model.ClaudeModelAlias{
	model.ClaudeModelOpus,
	model.ClaudeModelSonnet,
	model.ClaudeModelHaiku,
}

// ClaudeModelPickerState holds navigation state for the picker screen.
type ClaudeModelPickerState struct {
	Preset            ClaudeModelPreset
	CustomAssignments map[string]model.ClaudeModelAlias
	InCustomMode      bool
}

func NewClaudeModelPickerState() ClaudeModelPickerState {
	return ClaudeModelPickerState{
		Preset:            ClaudePresetBalanced,
		CustomAssignments: model.ClaudeModelPresetBalanced(),
	}
}

func NewClaudeModelPickerStateFromAssignments(assignments map[string]model.ClaudeModelAlias) ClaudeModelPickerState {
	if len(assignments) == 0 {
		return NewClaudeModelPickerState()
	}
	for preset, constructor := range claudePresetConstructors {
		if claudeAssignmentsEqual(constructor(), assignments) {
			return ClaudeModelPickerState{
				Preset:            preset,
				CustomAssignments: claudeCopyAssignments(assignments),
			}
		}
	}
	return ClaudeModelPickerState{
		Preset:            ClaudePresetCustom,
		CustomAssignments: claudeCopyAssignments(assignments),
	}
}

var claudePresetConstructors = map[ClaudeModelPreset]func() map[string]model.ClaudeModelAlias{
	ClaudePresetBalanced:    model.ClaudeModelPresetBalanced,
	ClaudePresetPerformance: model.ClaudeModelPresetPerformance,
	ClaudePresetEconomy:     model.ClaudeModelPresetEconomy,
	ClaudePresetDiversity:   model.ClaudeModelPresetDiversity,
}

func claudeAssignmentsEqual(a, b map[string]model.ClaudeModelAlias) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func claudeCopyAssignments(m map[string]model.ClaudeModelAlias) map[string]model.ClaudeModelAlias {
	out := make(map[string]model.ClaudeModelAlias, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// ClaudeModelPickerModel is a standalone BubbleTea model for Claude model picker.
type ClaudeModelPickerModel struct {
	state  ClaudeModelPickerState
	cursor int
}

func NewClaudeModelPickerModel(existing map[string]model.ClaudeModelAlias) ClaudeModelPickerModel {
	return ClaudeModelPickerModel{
		state: NewClaudeModelPickerStateFromAssignments(existing),
	}
}

func (m ClaudeModelPickerModel) Init() tea.Cmd { return nil }

func (m ClaudeModelPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		optCount := claudePickerOptionCount(m.state)

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
			handled, assignments := HandleClaudeModelPickerNav(key, &m.state, m.cursor)
			if !handled {
				// Back option
				return m, func() tea.Msg { return BackMsg{} }
			}
			if assignments != nil {
				return m, func() tea.Msg { return ClaudeModelsSelectedMsg{Assignments: assignments} }
			}
			// Entered custom mode or cycled alias — stay on screen
			m.cursor = 0
		}
	}
	return m, nil
}

func (m ClaudeModelPickerModel) View() string {
	return RenderClaudeModelPicker(m.state, m.cursor)
}

// HandleClaudeModelPickerNav processes a key press on the Claude model picker screen.
func HandleClaudeModelPickerNav(
	key string,
	state *ClaudeModelPickerState,
	cursor int,
) (handled bool, assignments map[string]model.ClaudeModelAlias) {
	if !state.InCustomMode {
		return claudeHandlePresetNav(key, state, cursor)
	}
	return claudeHandleCustomPhaseNav(key, state, cursor)
}

func claudeHandlePresetNav(
	key string,
	state *ClaudeModelPickerState,
	cursor int,
) (bool, map[string]model.ClaudeModelAlias) {
	if key != "enter" {
		return false, nil
	}
	if cursor >= len(claudePresetOrder) {
		return false, nil
	}

	selected := claudePresetOrder[cursor]
	state.Preset = selected

	if selected == ClaudePresetCustom {
		state.InCustomMode = true
		if state.CustomAssignments == nil {
			state.CustomAssignments = model.ClaudeModelPresetBalanced()
		}
		return true, nil
	}

	constructor := claudePresetConstructors[selected]
	assignments := constructor()
	state.CustomAssignments = assignments
	return true, assignments
}

func claudeHandleCustomPhaseNav(
	key string,
	state *ClaudeModelPickerState,
	cursor int,
) (bool, map[string]model.ClaudeModelAlias) {
	switch key {
	case "esc":
		state.InCustomMode = false
		return true, nil
	case "enter":
		if cursor < len(claudePhases) {
			phase := claudePhases[cursor]
			current := state.CustomAssignments[phase]
			state.CustomAssignments[phase] = claudeNextAlias(current)
			return true, nil
		}
		if cursor == len(claudePhases) {
			return true, state.CustomAssignments
		}
		state.InCustomMode = false
		return true, nil
	}
	return false, nil
}

func claudeNextAlias(current model.ClaudeModelAlias) model.ClaudeModelAlias {
	for i, a := range claudeAliasOrder {
		if a == current {
			return claudeAliasOrder[(i+1)%len(claudeAliasOrder)]
		}
	}
	return model.ClaudeModelSonnet
}

func claudePickerOptionCount(state ClaudeModelPickerState) int {
	if state.InCustomMode {
		return len(claudePhases) + 2 // phases + Confirm + Back
	}
	return len(claudePresetOrder) + 1 // presets + Back
}

// RenderClaudeModelPicker renders the Claude model picker screen.
func RenderClaudeModelPicker(state ClaudeModelPickerState, cursor int) string {
	if state.InCustomMode {
		return renderClaudeCustomPhaseList(state, cursor)
	}
	return renderClaudePresetList(state, cursor)
}

func renderClaudePresetList(state ClaudeModelPickerState, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Claude Model Assignments"))
	b.WriteString("\n")
	b.WriteString(styles.SubtextStyle.Render("Current: " + string(state.Preset)))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Choose how Claude models are assigned to each SDD phase:"))
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

func renderClaudeCustomPhaseList(state ClaudeModelPickerState, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Custom Model Assignments"))
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
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: cycle model / confirm • esc: back to presets"))

	return b.String()
}

func claudeAliasTag(alias model.ClaudeModelAlias) string {
	switch alias {
	case model.ClaudeModelOpus:
		return styles.WarningStyle.Render("[opus]")
	case model.ClaudeModelHaiku:
		return styles.SubtextStyle.Render("[haiku]")
	default:
		return styles.SuccessStyle.Render("[sonnet]")
	}
}

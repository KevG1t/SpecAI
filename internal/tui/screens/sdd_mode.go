package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// SDDModeSelectedMsg is emitted when the user confirms an SDD mode choice.
type SDDModeSelectedMsg struct {
	Mode model.SDDModeID
}

var sddModeOptions = []model.SDDModeID{
	model.SDDModeSingle,
	model.SDDModeMulti,
}

var sddModeDescriptions = map[model.SDDModeID]string{
	model.SDDModeSingle: "Single orchestrator — one agent handles all SDD phases",
	model.SDDModeMulti:  "Multi-agent — dedicated sub-agent per SDD phase (9 hidden agents)",
}

// SDDModeModel is a standalone BubbleTea model for SDD mode selection.
type SDDModeModel struct {
	cursor   int
	selected model.SDDModeID
}

func NewSDDModeModel() SDDModeModel {
	return SDDModeModel{
		selected: model.SDDModeSingle,
	}
}

func (m SDDModeModel) Init() tea.Cmd { return nil }

func (m SDDModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(sddModeOptions) {
				m.cursor++
				if m.cursor < len(sddModeOptions) {
					m.selected = sddModeOptions[m.cursor]
				}
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < len(sddModeOptions) {
					m.selected = sddModeOptions[m.cursor]
				}
			}
		case "enter":
			if m.cursor < len(sddModeOptions) {
				m.selected = sddModeOptions[m.cursor]
				mode := m.selected
				return m, func() tea.Msg { return SDDModeSelectedMsg{Mode: mode} }
			}
			return m, func() tea.Msg { return BackMsg{} }
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m SDDModeModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Select SDD Mode"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("How should the SDD orchestrator be configured for OpenCode?"))
	b.WriteString("\n\n")

	for idx, mode := range sddModeOptions {
		isSelected := mode == m.selected
		focused := idx == m.cursor
		b.WriteString(renderRadio(string(mode), isSelected, focused))
		b.WriteString(styles.SubtextStyle.Render("    "+sddModeDescriptions[mode]) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, m.cursor-len(sddModeOptions)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// StrictTDDSelectedMsg is emitted when the user confirms a Strict TDD choice.
type StrictTDDSelectedMsg struct {
	Enabled bool
}

const (
	StrictTDDOptionEnable  = 0
	StrictTDDOptionDisable = 1
)

// StrictTDDModel is a standalone BubbleTea model for Strict TDD selection.
type StrictTDDModel struct {
	cursor  int
	enabled bool
}

func NewStrictTDDModel() StrictTDDModel {
	return StrictTDDModel{
		enabled: true,
	}
}

func (m StrictTDDModel) Init() tea.Cmd { return nil }

func (m StrictTDDModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	options := []string{"Enable", "Disable"}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(options) {
				m.cursor++
				if m.cursor < len(options) {
					m.enabled = m.cursor == StrictTDDOptionEnable
				}
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < len(options) {
					m.enabled = m.cursor == StrictTDDOptionEnable
				}
			}
		case "enter":
			if m.cursor < len(options) {
				enabled := m.cursor == StrictTDDOptionEnable
				return m, func() tea.Msg { return StrictTDDSelectedMsg{Enabled: enabled} }
			}
			return m, func() tea.Msg { return BackMsg{} }
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m StrictTDDModel) View() string {
	var b strings.Builder
	options := []string{"Enable", "Disable"}

	b.WriteString(styles.TitleStyle.Render("STRICT TDD MODE"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Should agents follow Strict TDD (RED → GREEN → REFACTOR) for every task?"))
	b.WriteString("\n")
	b.WriteString(styles.SubtextStyle.Render("When enabled, the sdd-apply agent writes tests first, confirms failure,"))
	b.WriteString("\n")
	b.WriteString(styles.SubtextStyle.Render("then implements the minimum code to pass before refactoring."))
	b.WriteString("\n\n")

	for idx, opt := range options {
		isSelected := (idx == StrictTDDOptionEnable && m.enabled) || (idx == StrictTDDOptionDisable && !m.enabled)
		focused := idx == m.cursor
		b.WriteString(renderRadio(opt, isSelected, focused))
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, m.cursor-len(options)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

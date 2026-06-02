package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/planner"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// ReviewConfirmedMsg is emitted when the user confirms installation on the review screen.
type ReviewConfirmedMsg struct{}

// ReviewBackMsg is emitted when the user chooses to go back from the review screen.
type ReviewBackMsg struct{}

// ReviewModel is a standalone BubbleTea model for the review and confirm screen.
type ReviewModel struct {
	payload planner.ReviewPayload
	cursor  int // 0=Install, 1=Back
}

// NewReviewModel constructs a ReviewModel with the given payload.
func NewReviewModel(payload planner.ReviewPayload) ReviewModel {
	return ReviewModel{payload: payload}
}

// Init returns nil — no async initialization needed.
func (m ReviewModel) Init() tea.Cmd { return nil }

// Update handles key navigation and confirmation.
func (m ReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	const maxCursor = 1
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < maxCursor {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.cursor == 0 {
				return m, func() tea.Msg { return ReviewConfirmedMsg{} }
			}
			return m, func() tea.Msg { return ReviewBackMsg{} }
		case "esc":
			return m, func() tea.Msg { return ReviewBackMsg{} }
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the review summary and action selection.
func (m ReviewModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Review Installation"))
	b.WriteString("\n\n")

	p := m.payload

	// Agents
	b.WriteString(styles.HeadingStyle.Render("Agents"))
	b.WriteString("\n")
	for _, agent := range p.Agents {
		b.WriteString(fmt.Sprintf("  %s\n", string(agent)))
	}
	b.WriteString("\n")

	// Persona
	b.WriteString(styles.HeadingStyle.Render("Persona"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", string(p.Persona)))

	// Preset
	b.WriteString(styles.HeadingStyle.Render("Preset"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", string(p.Preset)))

	// Components
	autoAdded := make(map[string]struct{}, len(p.AddedDependencies))
	for _, dep := range p.AddedDependencies {
		autoAdded[string(dep)] = struct{}{}
	}

	b.WriteString(styles.HeadingStyle.Render("Components"))
	b.WriteString("\n")
	for _, comp := range p.Components {
		label := string(comp.ID)
		if _, isAuto := autoAdded[string(comp.ID)]; isAuto {
			b.WriteString(fmt.Sprintf("  %s %s\n", label, styles.SubtextStyle.Render("[auto-dependency]")))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", label))
		}
	}
	b.WriteString("\n")

	// Skills
	if len(p.Skills) > 0 {
		b.WriteString(styles.HeadingStyle.Render("Skills"))
		b.WriteString("\n")
		for _, skill := range p.Skills {
			b.WriteString(fmt.Sprintf("  %s\n", string(skill)))
		}
		b.WriteString("\n")
	}

	// SDD Mode
	b.WriteString(styles.HeadingStyle.Render("SDD Mode"))
	b.WriteString("\n")
	if p.HasSDD {
		b.WriteString("  enabled\n")
		if p.SDDMode != "" {
			b.WriteString(fmt.Sprintf("  SDD Mode: %s\n", string(p.SDDMode)))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("  disabled\n\n")
	}

	// Strict TDD
	b.WriteString(styles.HeadingStyle.Render("Strict TDD"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", boolStr(p.StrictTDD)))

	// Actions
	actions := []string{"Install", "Back"}
	b.WriteString(renderOptions(actions, m.cursor))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

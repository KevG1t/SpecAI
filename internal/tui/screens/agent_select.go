package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/catalog"
	"github.com/KevG1t/SpecAI/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AgentsSelectedMsg is emitted when the user confirms their agent selection.
type AgentsSelectedMsg struct {
	Agents []model.AgentID
}

// agentEntry holds display data for each row in the selection list.
type agentEntry struct {
	ID   model.AgentID
	Name string
}

// AgentSelectModel is a bubbletea model that shows a multi-select list of all
// supported agents. Detected (installed) agents are pre-checked.
type AgentSelectModel struct {
	agents  []agentEntry
	checked map[model.AgentID]bool
	cursor  int
	err     string
}

// NewAgentSelectModel builds the model. detected is the slice of AgentIDs that were
// found installed on the system — they will be pre-checked.
func NewAgentSelectModel(detected []model.AgentID) AgentSelectModel {
	all := catalog.AllAgents()

	entries := make([]agentEntry, len(all))
	checked := make(map[model.AgentID]bool, len(all))

	// Build a fast lookup set for detected agents.
	detectedSet := make(map[model.AgentID]struct{}, len(detected))
	for _, id := range detected {
		detectedSet[id] = struct{}{}
	}

	for i, a := range all {
		entries[i] = agentEntry{ID: a.ID, Name: a.Name}
		_, isDetected := detectedSet[a.ID]
		checked[a.ID] = isDetected
	}

	return AgentSelectModel{
		agents:  entries,
		checked: checked,
	}
}

func (m AgentSelectModel) Init() tea.Cmd {
	return nil
}

func (m AgentSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.agents)-1 {
				m.cursor++
			}
		case tea.KeyRunes:
			if string(msg.Runes) == " " {
				// Toggle current item
				id := m.agents[m.cursor].ID
				m.checked[id] = !m.checked[id]
				m.err = "" // clear error on interaction
			}
		case tea.KeyEnter:
			// Collect selected agents
			var selected []model.AgentID
			for _, a := range m.agents {
				if m.checked[a.ID] {
					selected = append(selected, a.ID)
				}
			}
			if len(selected) == 0 {
				m.err = "Please select at least one agent."
				return m, nil
			}
			m.err = ""
			return m, func() tea.Msg {
				return AgentsSelectedMsg{Agents: selected}
			}
		case tea.KeyEsc:
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

var (
	agentCheckedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	agentUncheckedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	agentCursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("62")).Bold(true).Padding(0, 1)
	agentErrorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	agentTitleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).MarginBottom(1)
)

func (m AgentSelectModel) View() string {
	var b strings.Builder

	b.WriteString(agentTitleStyle.Render("Select agents to install:"))
	b.WriteString("\n\n")

	for i, agent := range m.agents {
		cursor := "  "
		name := agent.Name
		if m.checked[agent.ID] {
			name = agentCheckedStyle.Render("[✓] " + agent.Name)
		} else {
			name = agentUncheckedStyle.Render("[ ] " + agent.Name)
		}

		if m.cursor == i {
			cursor = "> "
			// Build the raw checkbox prefix so it is visible in the highlighted row.
			checkboxPrefix := "[ ] "
			if m.checked[agent.ID] {
				checkboxPrefix = "[✓] "
			}
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, agentCursorStyle.Render(checkboxPrefix+agent.Name)))
		} else {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, name))
		}
	}

	if m.err != "" {
		b.WriteString("\n" + agentErrorStyle.Render(m.err) + "\n")
	}

	b.WriteString("\nArrows to navigate, Space to toggle, Enter to confirm, Esc to go back.\n")
	return b.String()
}

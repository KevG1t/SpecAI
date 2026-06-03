package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// MCPServersSelectedMsg is emitted when the user confirms MCP server selection.
type MCPServersSelectedMsg struct {
	ComponentIDs []model.ComponentID
}

// MCPItem holds display state for one MCP server checkbox row.
type MCPItem struct {
	ID      model.ComponentID
	Label   string
	Checked bool
}

// allMCPServers is the ordered list of available MCP servers.
var allMCPServers = []MCPItem{
	{ID: model.ComponentContext7, Label: "Context7 — latest framework docs (via npx)", Checked: true},
	{ID: model.ComponentNotion, Label: "Notion — workspace context (requires token)", Checked: false},
	{ID: model.ComponentJira, Label: "Jira — issue tracking (requires token)", Checked: false},
}

// MCPPickerModel is a standalone BubbleTea model for MCP server selection.
// It is shown only in the Custom preset flow, after SkillPicker.
type MCPPickerModel struct {
	items  []MCPItem
	cursor int
}

// NewMCPPickerModel constructs an MCPPickerModel with default selections.
func NewMCPPickerModel() MCPPickerModel {
	items := make([]MCPItem, len(allMCPServers))
	copy(items, allMCPServers)
	return MCPPickerModel{items: items}
}

// Init returns nil — no async initialization needed.
func (m MCPPickerModel) Init() tea.Cmd { return nil }

// Update handles cursor movement, toggling, confirmation, and back navigation.
func (m MCPPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case " ":
			// Toggle the focused item.
			if m.cursor < len(m.items) {
				m.items[m.cursor].Checked = !m.items[m.cursor].Checked
			}
		case "enter":
			selected := m.selectedComponents()
			return m, func() tea.Msg { return MCPServersSelectedMsg{ComponentIDs: selected} }
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the MCP server checkbox list.
func (m MCPPickerModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Select MCP Servers"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Choose which MCP servers to install. You can skip all if needed."))
	b.WriteString("\n\n")

	for idx, item := range m.items {
		b.WriteString(renderCheckbox(string(item.ID)+" — "+item.Label, item.Checked, idx == m.cursor))
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • space: toggle • enter: confirm • esc: back"))

	return b.String()
}

// selectedComponents returns all currently checked component IDs.
func (m MCPPickerModel) selectedComponents() []model.ComponentID {
	out := make([]model.ComponentID, 0, len(m.items))
	for _, it := range m.items {
		if it.Checked {
			out = append(out, it.ID)
		}
	}
	return out
}

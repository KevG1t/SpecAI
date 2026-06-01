package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var mascotLines = []string{
	`       ╭───────╮`,
	`       │ ⚆   ⚆ │`,
	`       │   ▱   │`,
	`       ╰───┬───╯`,
	`      ╭────┴────╮`,
	`      │         │`,
	`    ──│  SpecAI │──`,
	`      │         │`,
	`      ╰────┬────╯`,
	`          / \`,
	`         /   \`,
}

var gradientColors = []lipgloss.Color{
	lipgloss.Color("51"),  // Cyan
	lipgloss.Color("45"),  // Light Blue
	lipgloss.Color("39"),  // Blue
	lipgloss.Color("99"),  // Purple
	lipgloss.Color("201"), // Magenta
}

// RenderMascot returns the ASCII mascot with a top-to-bottom gradient.
func RenderMascot() string {
	total := len(mascotLines)
	if total == 0 {
		return ""
	}

	bands := len(gradientColors)
	var b strings.Builder

	for i, line := range mascotLines {
		bandIdx := (i * bands) / total
		if bandIdx >= bands {
			bandIdx = bands - 1
		}
		style := lipgloss.NewStyle().Foreground(gradientColors[bandIdx])
		b.WriteString(style.Render(line))
		if i < total-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// OptionSelectedMsg is sent when the user selects a menu item.
type OptionSelectedMsg struct {
	Option string
}

type WelcomeModel struct {
	choices  []string
	cursor   int
	selected string
}

func NewWelcomeModel() WelcomeModel {
	return WelcomeModel{
		choices: []string{
			"Install",
			"Setup Local (Inyectar en este Repo)",
			"Upgrade",
			"Sync",
			"Upgrade + Sync",
			"Backup",
			"Uninstall",
			"Salir",
		},
	}
}

func (m WelcomeModel) Init() tea.Cmd {
	return nil
}

func (m WelcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.choices[m.cursor]
			if m.selected == "Salir" {
				return m, tea.Quit
			}
			return m, func() tea.Msg {
				return OptionSelectedMsg{Option: m.selected}
			}
		}
	}
	return m, nil
}

var (
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).MarginBottom(1)
	itemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("62")).Bold(true).Padding(0, 1)
)

func (m WelcomeModel) View() string {
	var b strings.Builder
	b.WriteString(RenderMascot())
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("SpecAI - Bienvenido!"))
	b.WriteString("\n\n")

	for i, choice := range m.choices {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">"
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, selectedItemStyle.Render(choice)))
		} else {
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, itemStyle.Render(choice)))
		}
	}

	b.WriteString("\nUsa las flechas para moverte, Enter para seleccionar, q para salir.\n")
	return b.String()
}

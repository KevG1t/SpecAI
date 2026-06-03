package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// ConfigSelectedMsg is emitted when the user confirms config customization.
type ConfigSelectedMsg struct {
	Theme           string // "kanagawa" | "default"
	PermissionsLevel string // "strict" | "standard" | "permissive"
	EditorMode      string // "vim" | "emacs" | "default"
}

// configField represents a single form field with cycling options.
type configField struct {
	Label    string
	Options  []string
	Selected int
}

// configDefaultFields returns the initial set of config fields with defaults.
func configDefaultFields() []configField {
	return []configField{
		{
			Label:    "Theme",
			Options:  []string{"default", "kanagawa"},
			Selected: 0,
		},
		{
			Label:    "Permissions Level",
			Options:  []string{"standard", "strict", "permissive"},
			Selected: 0,
		},
		{
			Label:    "Editor Mode",
			Options:  []string{"default", "vim", "emacs"},
			Selected: 0,
		},
	}
}

// ConfigPickerModel is a standalone BubbleTea model for config customization.
// It is shown only in the Custom preset flow, after MCPPicker.
type ConfigPickerModel struct {
	fields []configField
	cursor int
}

// NewConfigPickerModel constructs a ConfigPickerModel with system defaults.
func NewConfigPickerModel() ConfigPickerModel {
	return ConfigPickerModel{
		fields: configDefaultFields(),
	}
}

// Init returns nil — no async initialization needed.
func (m ConfigPickerModel) Init() tea.Cmd { return nil }

// Update handles cursor navigation, option cycling, confirmation, and back navigation.
func (m ConfigPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.fields)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "l", "right", " ":
			// Cycle to the next option for the focused field.
			if m.cursor < len(m.fields) {
				f := &m.fields[m.cursor]
				f.Selected = (f.Selected + 1) % len(f.Options)
			}
		case "h", "left":
			// Cycle to the previous option for the focused field.
			if m.cursor < len(m.fields) {
				f := &m.fields[m.cursor]
				f.Selected = (f.Selected - 1 + len(f.Options)) % len(f.Options)
			}
		case "enter":
			return m, func() tea.Msg {
				return ConfigSelectedMsg{
					Theme:            m.fields[0].Options[m.fields[0].Selected],
					PermissionsLevel: m.fields[1].Options[m.fields[1].Selected],
					EditorMode:       m.fields[2].Options[m.fields[2].Selected],
				}
			}
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the config customization form.
func (m ConfigPickerModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Config Customization"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Customize your installation settings. Press enter to confirm."))
	b.WriteString("\n\n")

	for idx, field := range m.fields {
		focused := idx == m.cursor
		current := field.Options[field.Selected]

		label := field.Label + ": " + current
		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label))
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • h/l or space: cycle options • enter: confirm • esc: back"))

	return b.String()
}

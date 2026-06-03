package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// PersonaSelectedMsg is emitted when the user confirms a persona choice.
type PersonaSelectedMsg struct {
	Persona model.PersonaID
}

var personaOptions = []model.PersonaID{
	model.PersonaArgentina,
	model.PersonaNicaragua,
	model.PersonaNeutral,
	model.PersonaCustom,
}

var personaDescriptions = map[model.PersonaID]string{
	model.PersonaArgentina: "Managed Argentina persona with teaching-first guidance and Rioplatense tone",
	model.PersonaNicaragua: "Managed Nicaragua persona with teaching-first guidance and Central American tone",
	model.PersonaNeutral:   "Managed neutral persona with the same guidance and less regional tone",
	model.PersonaCustom:    "Keep your existing persona unmanaged; SpecAI does not inject a persona",
}

// PersonaModel is a standalone BubbleTea model for persona selection.
type PersonaModel struct {
	cursor   int
	selected model.PersonaID
}

func NewPersonaModel() PersonaModel {
	return PersonaModel{
		selected: model.PersonaArgentina,
	}
}

func (m PersonaModel) Init() tea.Cmd { return nil }

func (m PersonaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(personaOptions) {
				m.cursor++
				if m.cursor < len(personaOptions) {
					m.selected = personaOptions[m.cursor]
				}
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < len(personaOptions) {
					m.selected = personaOptions[m.cursor]
				}
			}
		case "enter":
			if m.cursor < len(personaOptions) {
				m.selected = personaOptions[m.cursor]
				persona := m.selected
				return m, func() tea.Msg { return PersonaSelectedMsg{Persona: persona} }
			}
			// Back option
			return m, func() tea.Msg { return BackMsg{} }
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m PersonaModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Choose your Persona"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Your own Argentina! teaches before it solves."))
	b.WriteString("\n\n")

	for idx, persona := range personaOptions {
		isSelected := persona == m.selected
		focused := idx == m.cursor
		b.WriteString(renderRadio(string(persona), isSelected, focused))
		b.WriteString(styles.SubtextStyle.Render("    " + personaDescriptions[persona]))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, m.cursor-len(personaOptions)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

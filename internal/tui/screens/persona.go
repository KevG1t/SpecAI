package screens

import (
	"strings"

	"github.com/KevG1t/specai/internal/model"
	"github.com/KevG1t/specai/internal/tui/styles"
)

func PersonaOptions() []model.PersonaID {
	ids := make([]model.PersonaID, 0, len(model.Personas))
	for _, p := range model.Personas {
		ids = append(ids, p.ID)
	}
	return ids
}

var personaDescriptions = func() map[model.PersonaID]string {
	m := make(map[model.PersonaID]string, len(model.Personas))
	for _, p := range model.Personas {
		m[p.ID] = p.Description
	}
	return m
}()

func RenderPersona(selected model.PersonaID, cursor int) string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Choose your Persona"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("SpecAI teaches before it solves."))
	b.WriteString("\n\n")

	for idx, persona := range PersonaOptions() {
		isSelected := persona == selected
		focused := idx == cursor
		b.WriteString(renderRadio(string(persona), isSelected, focused))
		b.WriteString(styles.SubtextStyle.Render("    " + personaDescriptions[persona]))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, cursor-len(PersonaOptions())))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

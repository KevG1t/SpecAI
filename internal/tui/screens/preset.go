package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// PresetSelectedMsg is emitted when the user confirms a preset choice.
type PresetSelectedMsg struct {
	Preset model.PresetID
}

var presetOptions = []model.PresetID{
	model.PresetFull,
	model.PresetEcosystemOnly,
	model.PresetMinimal,
	model.PresetCustom,
}

var presetDescriptions = map[model.PresetID]string{
	model.PresetMinimal:       "Just SDD Memory persistent memory across sessions",
	model.PresetEcosystemOnly: "Memory + SDD + skills + docs",
	model.PresetFull:          "Dev Stack + security gates, theme, and logo",
	model.PresetCustom:        "Choose components and skills manually; keep existing persona/settings unmanaged",
}

var presetLabels = map[model.PresetID]string{
	model.PresetMinimal:       "Memory Only",
	model.PresetEcosystemOnly: "Dev Stack",
	model.PresetFull:          "Dev Stack + Polish",
	model.PresetCustom:        "Custom",
}

// PresetModel is a standalone BubbleTea model for preset selection.
type PresetModel struct {
	cursor   int
	selected model.PresetID
}

func NewPresetModel() PresetModel {
	return PresetModel{
		selected: model.PresetFull,
	}
}

func (m PresetModel) Init() tea.Cmd { return nil }

func (m PresetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(presetOptions) {
				m.cursor++
				if m.cursor < len(presetOptions) {
					m.selected = presetOptions[m.cursor]
				}
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < len(presetOptions) {
					m.selected = presetOptions[m.cursor]
				}
			}
		case "enter":
			if m.cursor < len(presetOptions) {
				m.selected = presetOptions[m.cursor]
				preset := m.selected
				return m, func() tea.Msg { return PresetSelectedMsg{Preset: preset} }
			}
			// Back option
			return m, func() tea.Msg { return BackMsg{} }
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		}
	}
	return m, nil
}

func (m PresetModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Select Ecosystem Preset"))
	b.WriteString("\n\n")

	for idx, preset := range presetOptions {
		isSelected := preset == m.selected
		focused := idx == m.cursor
		b.WriteString(renderRadio(presetLabels[preset], isSelected, focused))
		b.WriteString(styles.SubtextStyle.Render("    "+presetDescriptions[preset]) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(renderOptions([]string{"Back"}, m.cursor-len(presetOptions)))
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))

	return b.String()
}

package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// SkillsSelectedMsg is emitted when the user confirms skill selection.
type SkillsSelectedMsg struct {
	Skills []model.SkillID
}

// SkillItem holds display state for one skill checkbox row.
type SkillItem struct {
	ID      model.SkillID
	Label   string
	Checked bool
}

// allSkills is the ordered list of all known skill IDs with display labels.
var allSkills = []SkillItem{
	{ID: model.SkillSDDInit, Label: "SDD Init"},
	{ID: model.SkillSDDApply, Label: "SDD Apply"},
	{ID: model.SkillSDDVerify, Label: "SDD Verify"},
	{ID: model.SkillSDDExplore, Label: "SDD Explore"},
	{ID: model.SkillSDDPropose, Label: "SDD Propose"},
	{ID: model.SkillSDDSpec, Label: "SDD Spec"},
	{ID: model.SkillSDDDesign, Label: "SDD Design"},
	{ID: model.SkillSDDTasks, Label: "SDD Tasks"},
	{ID: model.SkillSDDArchive, Label: "SDD Archive"},
	{ID: model.SkillSDDOnboard, Label: "SDD Onboard"},
	{ID: model.SkillGoTesting, Label: "Go Testing"},
	{ID: model.SkillCreator, Label: "Skill Creator"},
	{ID: model.SkillImprover, Label: "Skill Improver"},
	{ID: model.SkillJudgmentDay, Label: "Judgment Day"},
	{ID: model.SkillBranchPR, Label: "Branch PR"},
	{ID: model.SkillIssueCreation, Label: "Issue Creation"},
	{ID: model.SkillSkillRegistry, Label: "Skill Registry"},
	{ID: model.SkillChainedPR, Label: "Chained PR"},
	{ID: model.SkillCognitiveDoc, Label: "Cognitive Doc Design"},
	{ID: model.SkillCommentWriter, Label: "Comment Writer"},
	{ID: model.SkillWorkUnitCommits, Label: "Work Unit Commits"},
}

// SkillPickerModel is a standalone BubbleTea model for custom skill selection.
type SkillPickerModel struct {
	items  []SkillItem
	cursor int
}

// NewSkillPickerModel constructs a SkillPickerModel.
// preSelected holds the IDs that should start checked. If nil, all skills start checked.
func NewSkillPickerModel(preSelected []model.SkillID) SkillPickerModel {
	items := make([]SkillItem, len(allSkills))
	copy(items, allSkills)

	if preSelected == nil {
		// No pre-selection provided: check all skills by default.
		for i := range items {
			items[i].Checked = true
		}
	} else {
		selectedSet := make(map[model.SkillID]struct{}, len(preSelected))
		for _, id := range preSelected {
			selectedSet[id] = struct{}{}
		}
		for i := range items {
			_, items[i].Checked = selectedSet[items[i].ID]
		}
	}

	return SkillPickerModel{items: items}
}

// Init returns nil — no async initialization needed.
func (m SkillPickerModel) Init() tea.Cmd { return nil }

// Update handles cursor movement, toggling, confirmation and quit.
func (m SkillPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Toggle the focused item, but guard against unchecking the last checked skill.
			if m.cursor < len(m.items) {
				if m.items[m.cursor].Checked {
					// Only uncheck if there is more than one currently checked item.
					checkedCount := 0
					for _, it := range m.items {
						if it.Checked {
							checkedCount++
						}
					}
					if checkedCount > 1 {
						m.items[m.cursor].Checked = false
					}
				} else {
					m.items[m.cursor].Checked = true
				}
			}
		case "enter":
			selected := m.selectedSkills()
			return m, func() tea.Msg { return SkillsSelectedMsg{Skills: selected} }
		case "esc":
			return m, func() tea.Msg { return BackMsg{} }
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the skill checkbox list.
func (m SkillPickerModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Select Skills"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Choose the skills to include. At least one must remain selected."))
	b.WriteString("\n\n")

	for idx, item := range m.items {
		b.WriteString(renderCheckbox(string(item.ID)+" — "+item.Label, item.Checked, idx == m.cursor))
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • space: toggle • enter: confirm • esc: back"))

	return b.String()
}

// selectedSkills returns all currently checked skill IDs.
func (m SkillPickerModel) selectedSkills() []model.SkillID {
	out := make([]model.SkillID, 0, len(m.items))
	for _, it := range m.items {
		if it.Checked {
			out = append(out, it.ID)
		}
	}
	return out
}

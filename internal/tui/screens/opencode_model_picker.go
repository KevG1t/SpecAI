package screens

import (
	"strings"

	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

// OpenCodeModelsSelectedMsg is emitted when the user finishes assigning models
// to OpenCode SDD phases.
type OpenCodeModelsSelectedMsg struct {
	Assignments map[string]model.ModelAssignment
}

// ocNavLevel is the 4-level navigation state for the OpenCode model picker.
type ocNavLevel int

const (
	ocNavPhaseList     ocNavLevel = iota
	ocNavProviderSelect            // nolint:deadcode,varcheck
	ocNavModelSelect               // nolint:deadcode,varcheck
	ocNavEffortSelect              // nolint:deadcode,varcheck
)

// PhaseRow holds one row in the phase list.
type PhaseRow struct {
	Key         string // e.g. "sdd-apply"
	Label       string // e.g. "Apply"
	IsSetAll    bool   // true for the "Set all SDD phases" meta-row
	IsSeparator bool   // true for visual divider rows
}

// orderedSddPhaseKeys lists the SDD phase keys that "Set all" applies to.
// This matches the phase list order, excluding orchestrator and judgment-day entries.
var orderedSddPhaseKeys = []string{
	"sdd-init",
	"sdd-explore",
	"sdd-propose",
	"sdd-spec",
	"sdd-design",
	"sdd-tasks",
	"sdd-apply",
	"sdd-verify",
	"sdd-archive",
	"sdd-onboard",
	"jd-judge-a",
	"jd-judge-b",
	"jd-fix-agent",
}

// buildPhaseRows returns the canonical phase row list.
func buildPhaseRows() []PhaseRow {
	return []PhaseRow{
		{Key: "gentle-orchestrator", Label: "Orchestrator"},
		{Key: "__set_all__", Label: "Set all SDD phases", IsSetAll: true},
		{Key: "sdd-init", Label: "SDD Init"},
		{Key: "sdd-explore", Label: "SDD Explore"},
		{Key: "sdd-propose", Label: "SDD Propose"},
		{Key: "sdd-spec", Label: "SDD Spec"},
		{Key: "sdd-design", Label: "SDD Design"},
		{Key: "sdd-tasks", Label: "SDD Tasks"},
		{Key: "sdd-apply", Label: "SDD Apply"},
		{Key: "sdd-verify", Label: "SDD Verify"},
		{Key: "sdd-archive", Label: "SDD Archive"},
		{Key: "sdd-onboard", Label: "SDD Onboard"},
		{Key: "", Label: "--- Judgment Day ---", IsSeparator: true},
		{Key: "jd-judge-a", Label: "JD Judge A"},
		{Key: "jd-judge-b", Label: "JD Judge B"},
		{Key: "jd-fix-agent", Label: "JD Fix Agent"},
	}
}

// OpenCodeModelPickerModel is a standalone BubbleTea model for the OpenCode model picker.
type OpenCodeModelPickerModel struct {
	navLevel         ocNavLevel
	phases           []PhaseRow
	providers        []model.ProviderEntry
	assignments      map[string]model.ModelAssignment
	search           string
	cursor           int
	selectedPhase    string
	selectedProvider model.ProviderEntry
	selectedModel    model.ModelEntry // set when transitioning from ModelSelect → EffortSelect
}

// NewOpenCodeModelPickerModel constructs a fresh OpenCodeModelPickerModel.
func NewOpenCodeModelPickerModel() OpenCodeModelPickerModel {
	return OpenCodeModelPickerModel{
		navLevel:    ocNavPhaseList,
		phases:      buildPhaseRows(),
		providers:   model.OpenCodeProviders,
		assignments: make(map[string]model.ModelAssignment),
	}
}

// Init returns nil — no async initialization needed.
func (m OpenCodeModelPickerModel) Init() tea.Cmd { return nil }

// Update handles the 4-level navigation state machine.
func (m OpenCodeModelPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.navLevel {
		case ocNavPhaseList:
			return m.updatePhaseList(msg)
		case ocNavProviderSelect:
			return m.updateProviderSelect(msg)
		case ocNavModelSelect:
			return m.updateModelSelect(msg)
		case ocNavEffortSelect:
			return m.updateEffortSelect(msg)
		}
	}
	return m, nil
}

// updatePhaseList handles key input at the PhaseList level.
func (m OpenCodeModelPickerModel) updatePhaseList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.cursor = m.nextNavigablePhase(m.cursor, 1)
	case "k", "up":
		m.cursor = m.nextNavigablePhase(m.cursor, -1)
	case "enter":
		if m.cursor >= len(m.phases) {
			break
		}
		row := m.phases[m.cursor]
		if row.IsSeparator {
			break
		}
		if row.IsSetAll {
			m.selectedPhase = "__set_all__"
		} else {
			m.selectedPhase = row.Key
		}
		m.navLevel = ocNavProviderSelect
		m.cursor = 0
		m.search = ""
	case "esc", "s":
		// Emit assignments and exit
		assignments := m.copyAssignments()
		return m, func() tea.Msg { return OpenCodeModelsSelectedMsg{Assignments: assignments} }
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// updateProviderSelect handles key input at the ProviderSelect level.
func (m OpenCodeModelPickerModel) updateProviderSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.providers)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.cursor < len(m.providers) {
			m.selectedProvider = m.providers[m.cursor]
			m.navLevel = ocNavModelSelect
			m.cursor = 0
			m.search = ""
		}
	case "esc":
		m.navLevel = ocNavPhaseList
		m.cursor = 0
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// updateModelSelect handles key input at the ModelSelect level.
// At this level j/k are treated as search characters (not navigation) so that
// model names containing those letters can be typed without triggering cursor
// movement. Arrow keys (down/up) are used for navigation instead.
func (m OpenCodeModelPickerModel) updateModelSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.filteredModels()

	switch msg.Type {
	case tea.KeyDown:
		if m.cursor < len(filtered)-1 {
			m.cursor++
		}
		return m, nil
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	}

	switch msg.String() {
	case "enter":
		if m.cursor < len(filtered) {
			chosen := filtered[m.cursor]
			if chosen.SupportsEffort {
				m.selectedModel = chosen
				m.navLevel = ocNavEffortSelect
				m.cursor = 0
			} else {
				m.saveAssignment(m.selectedPhase, m.selectedProvider, chosen, "")
				m.navLevel = ocNavPhaseList
				m.cursor = 0
				m.search = ""
			}
		}
	case "backspace":
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.cursor = 0
		}
	case "esc":
		m.navLevel = ocNavProviderSelect
		m.cursor = 0
		m.search = ""
	case "q", "ctrl+c":
		return m, tea.Quit
	default:
		// Any printable rune appends to the search filter.
		if len(msg.Runes) > 0 {
			m.search += string(msg.Runes)
			m.cursor = 0
		}
	}
	return m, nil
}

var effortOptions = []string{"low", "medium", "high"}

// updateEffortSelect handles key input at the EffortSelect level.
func (m OpenCodeModelPickerModel) updateEffortSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(effortOptions)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.cursor < len(effortOptions) {
			effort := effortOptions[m.cursor]
			m.saveAssignment(m.selectedPhase, m.selectedProvider, m.selectedModel, effort)
			m.navLevel = ocNavPhaseList
			m.cursor = 0
			m.search = ""
		}
	case "esc":
		m.navLevel = ocNavModelSelect
		m.cursor = 0
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// saveAssignment stores the ModelAssignment for a phase (or all SDD phases for __set_all__).
func (m *OpenCodeModelPickerModel) saveAssignment(
	phase string,
	provider model.ProviderEntry,
	me model.ModelEntry,
	effort string,
) {
	assignment := model.ModelAssignment{
		ProviderID: provider.ID,
		ModelID:    me.ID,
		Effort:     effort,
	}

	if phase == "__set_all__" {
		for _, key := range orderedSddPhaseKeys {
			m.assignments[key] = assignment
		}
		return
	}
	m.assignments[phase] = assignment
}

// filteredModels returns the models of selectedProvider filtered by the current search string.
func (m OpenCodeModelPickerModel) filteredModels() []model.ModelEntry {
	return m.filteredModelsForProvider(m.selectedProvider)
}

// filteredModelsForProvider filters a provider's models by m.search.
func (m OpenCodeModelPickerModel) filteredModelsForProvider(p model.ProviderEntry) []model.ModelEntry {
	searchLower := strings.ToLower(m.search)

	if searchLower == "" {
		out := make([]model.ModelEntry, len(p.Models))
		copy(out, p.Models)
		return out
	}

	out := make([]model.ModelEntry, 0, len(p.Models))
	for _, me := range p.Models {
		if strings.Contains(strings.ToLower(me.Name), searchLower) ||
			strings.Contains(strings.ToLower(me.ID), searchLower) {
			out = append(out, me)
		}
	}
	return out
}

// nextNavigablePhase moves the cursor in the given direction, skipping separator rows.
func (m OpenCodeModelPickerModel) nextNavigablePhase(current, dir int) int {
	next := current + dir
	for next >= 0 && next < len(m.phases) {
		if !m.phases[next].IsSeparator {
			return next
		}
		next += dir
	}
	return current
}

// copyAssignments returns a shallow copy of the assignments map.
func (m OpenCodeModelPickerModel) copyAssignments() map[string]model.ModelAssignment {
	out := make(map[string]model.ModelAssignment, len(m.assignments))
	for k, v := range m.assignments {
		out[k] = v
	}
	return out
}

// View renders the current navigation level.
func (m OpenCodeModelPickerModel) View() string {
	switch m.navLevel {
	case ocNavPhaseList:
		return m.viewPhaseList()
	case ocNavProviderSelect:
		return m.viewProviderSelect()
	case ocNavModelSelect:
		return m.viewModelSelect()
	case ocNavEffortSelect:
		return m.viewEffortSelect()
	}
	return ""
}

func (m OpenCodeModelPickerModel) viewPhaseList() string {
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render("OpenCode Model Assignments"))
	b.WriteString("\n\n")
	b.WriteString(styles.SubtextStyle.Render("Select a phase to assign a provider/model."))
	b.WriteString("\n\n")

	for idx, row := range m.phases {
		if row.IsSeparator {
			b.WriteString(styles.SubtextStyle.Render("  "+row.Label) + "\n")
			continue
		}

		label := row.Label
		if a, ok := m.assignments[row.Key]; ok && row.Key != "__set_all__" {
			label += " " + styles.SuccessStyle.Render("["+a.ProviderID+"/"+a.ModelID+"]")
		}

		focused := idx == m.cursor
		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+label) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select phase • esc/s: done"))
	return b.String()
}

func (m OpenCodeModelPickerModel) viewProviderSelect() string {
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render("Select Provider"))
	b.WriteString("\n\n")

	for idx, p := range m.providers {
		focused := idx == m.cursor
		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+p.Name) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+p.Name) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))
	return b.String()
}

func (m OpenCodeModelPickerModel) viewModelSelect() string {
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render("Select Model — " + m.selectedProvider.Name))
	b.WriteString("\n")

	if m.search != "" {
		b.WriteString(styles.SubtextStyle.Render("Filter: "+m.search))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	filtered := m.filteredModels()
	for idx, me := range filtered {
		label := me.Name
		if me.SupportsEffort {
			label += " " + styles.WarningStyle.Render("[effort]")
		}
		focused := idx == m.cursor
		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+label) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+label) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • type to filter • enter: select • esc: back"))
	return b.String()
}

func (m OpenCodeModelPickerModel) viewEffortSelect() string {
	var b strings.Builder
	b.WriteString(styles.TitleStyle.Render("Select Effort Level"))
	b.WriteString("\n\n")

	for idx, effort := range effortOptions {
		focused := idx == m.cursor
		if focused {
			b.WriteString(styles.SelectedStyle.Render(styles.Cursor+effort) + "\n")
		} else {
			b.WriteString(styles.UnselectedStyle.Render("  "+effort) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("j/k: navigate • enter: select • esc: back"))
	return b.String()
}

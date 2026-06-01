package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type SetupLocalModel struct {
	State     InstallState // Reusing InstallState since it's the same flow
	Spinner   spinner.Model
	Err       error
}

func NewSetupLocalModel() SetupLocalModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.WarningStyle

	return SetupLocalModel{
		State:   InstallStateConfirm,
		Spinner: s,
	}
}

func (m SetupLocalModel) Init() tea.Cmd {
	return nil
}

func (m SetupLocalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.State == InstallStateConfirm || m.State == InstallStateResult {
				return m, func() tea.Msg { return BackMsg{} }
			}
		case "enter":
			if m.State == InstallStateConfirm {
				m.State = InstallStateRunning
				return m, tea.Batch(m.Spinner.Tick, func() tea.Msg { return StartPipelineMsg{Action: "SetupLocal"} })
			} else if m.State == InstallStateResult {
				return m, func() tea.Msg { return BackMsg{} }
			}
		}
	case PipelineFinishedMsg:
		m.State = InstallStateResult
		m.Err = msg.Err
		return m, nil
	case spinner.TickMsg:
		if m.State == InstallStateRunning {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m SetupLocalModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Setup Local (Inyectar en este Repo)"))
	b.WriteString("\n\n")

	switch m.State {
	case InstallStateRunning:
		b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("%s  Configurando el repositorio local...", m.Spinner.View())))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Por favor espera..."))
	case InstallStateResult:
		if m.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Falla en la configuración local"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render(m.Err.Error()))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Configuración local completada"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render("Las reglas de SpecAI han sido inyectadas en el directorio actual."))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: volver al menú • esc: atrás"))
	case InstallStateConfirm:
		b.WriteString(styles.UnselectedStyle.Render("Esta operación instalará la configuración de SpecAI"))
		b.WriteString("\n")
		b.WriteString(styles.UnselectedStyle.Render("únicamente en el directorio del repositorio actual."))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Presiona Enter para continuar"))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: confirmar • esc: atrás"))
	}

	return b.String()
}

package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type UpgradeModel struct {
	State     InstallState
	Spinner   spinner.Model
	Err       error
}

func NewUpgradeModel() UpgradeModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.WarningStyle

	return UpgradeModel{
		State:   InstallStateConfirm,
		Spinner: s,
	}
}

func (m UpgradeModel) Init() tea.Cmd {
	return nil
}

func (m UpgradeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, tea.Batch(m.Spinner.Tick, func() tea.Msg { return StartPipelineMsg{Action: "Upgrade"} })
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

func (m UpgradeModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Upgrade SpecAI"))
	b.WriteString("\n\n")

	switch m.State {
	case InstallStateRunning:
		b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("%s  Descargando e instalando actualización...", m.Spinner.View())))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Por favor espera..."))
	case InstallStateResult:
		if m.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Falla en la actualización"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render(m.Err.Error()))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Actualización completada exitosamente"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render("SpecAI se ha actualizado a la última versión disponible."))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: volver al menú • esc: atrás"))
	case InstallStateConfirm:
		b.WriteString(styles.UnselectedStyle.Render("Verificaremos si hay actualizaciones disponibles en GitHub."))
		b.WriteString("\n")
		b.WriteString(styles.UnselectedStyle.Render("Si hay una nueva versión, se descargará y reemplazará el binario actual."))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Presiona Enter para buscar y actualizar"))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: confirmar • esc: atrás"))
	}

	return b.String()
}

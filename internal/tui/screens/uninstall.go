package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type UninstallModel struct {
	State     InstallState
	Spinner   spinner.Model
	Err       error
}

func NewUninstallModel() UninstallModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.WarningStyle

	return UninstallModel{
		State:   InstallStateConfirm,
		Spinner: s,
	}
}

func (m UninstallModel) Init() tea.Cmd {
	return nil
}

func (m UninstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, tea.Batch(m.Spinner.Tick, func() tea.Msg { return StartPipelineMsg{Action: "Uninstall"} })
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

func (m UninstallModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Uninstall SpecAI"))
	b.WriteString("\n\n")

	switch m.State {
	case InstallStateRunning:
		b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("%s  Desinstalando configuraciones...", m.Spinner.View())))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Por favor espera..."))
	case InstallStateResult:
		if m.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Falla en la desinstalación"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render(m.Err.Error()))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Desinstalación completada"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render("Todas las reglas gestionadas por SpecAI han sido removidas."))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: volver al menú • esc: atrás"))
	case InstallStateConfirm:
		b.WriteString(styles.ErrorStyle.Render("Esta operación eliminará todas las configuraciones"))
		b.WriteString("\n")
		b.WriteString(styles.ErrorStyle.Render("gestionadas por SpecAI en todos los entornos detectados."))
		b.WriteString("\n\n")
		b.WriteString(styles.WarningStyle.Render("Se creará un backup automáticamente antes de modificar cualquier archivo."))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Presiona Enter para proceder con la desinstalación"))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: desinstalar • esc: cancelar y volver"))
	}

	return b.String()
}

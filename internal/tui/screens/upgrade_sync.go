package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type UpgradeSyncModel struct {
	State     InstallState
	Spinner   spinner.Model
	Err       error
}

func NewUpgradeSyncModel() UpgradeSyncModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.WarningStyle

	return UpgradeSyncModel{
		State:   InstallStateConfirm,
		Spinner: s,
	}
}

func (m UpgradeSyncModel) Init() tea.Cmd {
	return nil
}

func (m UpgradeSyncModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, tea.Batch(m.Spinner.Tick, func() tea.Msg { return StartPipelineMsg{Action: "UpgradeSync"} })
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

func (m UpgradeSyncModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Upgrade + Sync"))
	b.WriteString("\n\n")

	switch m.State {
	case InstallStateRunning:
		b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("%s  Actualizando y sincronizando...", m.Spinner.View())))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Por favor espera..."))
	case InstallStateResult:
		if m.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Falla en la operación"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render(m.Err.Error()))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Actualización y Sincronización completadas"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render("SpecAI fue actualizado y todos los IDEs detectados fueron sincronizados."))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: volver al menú • esc: atrás"))
	case InstallStateConfirm:
		b.WriteString(styles.UnselectedStyle.Render("Esta operación actualizará SpecAI a su última versión"))
		b.WriteString("\n")
		b.WriteString(styles.UnselectedStyle.Render("y luego aplicará la sincronización en tus IDEs globales."))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Presiona Enter para iniciar"))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: confirmar • esc: atrás"))
	}

	return b.String()
}

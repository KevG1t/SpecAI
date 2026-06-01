package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type InstallState int

const (
	InstallStateConfirm InstallState = iota
	InstallStateRunning
	InstallStateResult
)

type InstallModel struct {
	State     InstallState
	Spinner   spinner.Model
	Err       error
}

func NewInstallModel() InstallModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.WarningStyle

	return InstallModel{
		State:   InstallStateConfirm,
		Spinner: s,
	}
}

func (m InstallModel) Init() tea.Cmd {
	return nil
}

func (m InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, tea.Batch(m.Spinner.Tick, func() tea.Msg { return StartPipelineMsg{Action: "Install"} })
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

func (m InstallModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Install SpecAI Globally"))
	b.WriteString("\n\n")

	switch m.State {
	case InstallStateRunning:
		b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("%s  Instalando configuración global...", m.Spinner.View())))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Por favor espera..."))
	case InstallStateResult:
		if m.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Falla en la instalación"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render(m.Err.Error()))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Instalación completada exitosamente"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render("Las reglas globales y los skills han sido inyectados."))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: volver al menú • esc: atrás"))
	case InstallStateConfirm:
		b.WriteString(styles.UnselectedStyle.Render("Esta operación instalará SpecAI de forma global en tu entorno."))
		b.WriteString("\n")
		b.WriteString(styles.UnselectedStyle.Render("Detectará todos tus IDEs y configurará los dotfiles globales."))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Presiona Enter para instalar"))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: confirmar • esc: atrás"))
	}

	return b.String()
}

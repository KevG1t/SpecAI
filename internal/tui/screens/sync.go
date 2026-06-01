package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

type SyncState int

const (
	SyncStateConfirm SyncState = iota
	SyncStateRunning
	SyncStateResult
)

type SyncModel struct {
	State     SyncState
	Spinner   spinner.Model
	Err       error
}

func NewSyncModel() SyncModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.WarningStyle

	return SyncModel{
		State:   SyncStateConfirm,
		Spinner: s,
	}
}

func (m SyncModel) Init() tea.Cmd {
	return nil
}

func (m SyncModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.State == SyncStateConfirm || m.State == SyncStateResult {
				return m, func() tea.Msg { return BackMsg{} }
			}
		case "enter":
			if m.State == SyncStateConfirm {
				m.State = SyncStateRunning
				return m, tea.Batch(m.Spinner.Tick, func() tea.Msg { return StartPipelineMsg{Action: "Sync"} })
			} else if m.State == SyncStateResult {
				return m, func() tea.Msg { return BackMsg{} }
			}
		}
	case PipelineFinishedMsg:
		m.State = SyncStateResult
		m.Err = msg.Err
		return m, nil
	case spinner.TickMsg:
		if m.State == SyncStateRunning {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m SyncModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Sync Configurations"))
	b.WriteString("\n\n")

	switch m.State {
	case SyncStateRunning:
		b.WriteString(styles.WarningStyle.Render(fmt.Sprintf("%s  Sincronizando configuraciones...", m.Spinner.View())))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Por favor espera..."))
	case SyncStateResult:
		if m.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Falla en la sincronización"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render(m.Err.Error()))
			b.WriteString("\n\n")
			b.WriteString(styles.HelpStyle.Render("Revisa la configuración y vuelve a intentar."))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Sincronización completada"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render("Todas las reglas y configuraciones han sido sincronizadas."))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: volver al menú • esc: atrás"))
	case SyncStateConfirm:
		b.WriteString(styles.UnselectedStyle.Render("Sync re-aplicará las configuraciones globales"))
		b.WriteString("\n")
		b.WriteString(styles.UnselectedStyle.Render("a todos los IDEs detectados en tu máquina."))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Presiona Enter para sincronizar"))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: confirmar • esc: atrás"))
	}

	return b.String()
}

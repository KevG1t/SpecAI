package screens

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type InstallModel struct {
	Started bool
	Done    bool
	Err     error
	Spinner spinner.Model
	Steps   []string
	Status  map[string]string
}

func NewInstallModel() InstallModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.WarningStyle

	return InstallModel{
		Spinner: s,
		Status:  make(map[string]string),
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
			if !m.Started || m.Done {
				return m, func() tea.Msg { return BackMsg{} }
			}
		case "enter":
			if !m.Started {
				m.Started = true
				return m, tea.Batch(m.Spinner.Tick, func() tea.Msg { return StartPipelineMsg{Action: "Install"} })
			} else if m.Done {
				return m, func() tea.Msg { return BackMsg{} }
			}
		}
	case PipelineFinishedMsg:
		m.Done = true
		m.Err = msg.Err
		return m, nil
	case spinner.TickMsg:
		if m.Started && !m.Done {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	case ProgressMsg:
		if _, exists := m.Status[msg.TaskName]; !exists {
			m.Steps = append(m.Steps, msg.TaskName)
		}
		m.Status[msg.TaskName] = msg.Status
		return m, nil
	}
	return m, nil
}

func (m InstallModel) View() string {
	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("Install SpecAI Globally"))
	b.WriteString("\n\n")

	if !m.Started {
		b.WriteString(styles.UnselectedStyle.Render("Esta operación instalará SpecAI de forma global en tu entorno."))
		b.WriteString("\n")
		b.WriteString(styles.UnselectedStyle.Render("Detectará todos tus IDEs y configurará los dotfiles globales."))
		b.WriteString("\n\n")
		b.WriteString(styles.HeadingStyle.Render("Presiona Enter para instalar"))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: confirmar • esc: atrás"))
		return b.String()
	}

	for _, step := range m.Steps {
		b.WriteString(m.renderStep(step) + "\n")
	}


	if !m.Done {
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("Por favor espera..."))
	} else {
		b.WriteString("\n")
		if m.Err != nil {
			b.WriteString(styles.ErrorStyle.Render("✗ Falla en la instalación"))
			b.WriteString("\n\n")
			b.WriteString(styles.SubtextStyle.Render(m.Err.Error()))
		} else {
			b.WriteString(styles.SuccessStyle.Render("✓ Instalación completada exitosamente"))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("enter: volver al menú • esc: atrás"))
	}

	return b.String()
}

func (m InstallModel) renderStep(step string) string {
	status := m.Status[step]
	icon := " "
	if status == "running" {
		icon = m.Spinner.View()
	} else if status == "succeeded" {
		icon = styles.SuccessStyle.Render("✓")
	} else if status == "failed" {
		icon = styles.ErrorStyle.Render("✗")
	}
	return fmt.Sprintf("%s %s", icon, step)
}

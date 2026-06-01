package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KevG1t/SpecAI/internal/backup"
	tea "github.com/charmbracelet/bubbletea"
)

// RestoreDoneMsg is sent when a restore finishes.
type RestoreDoneMsg struct {
	Err error
}

func ListBackups() []backup.Manifest {
	root, err := backup.BackupRootFn()
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}

	var manifests []backup.Manifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(root, entry.Name(), "manifest.json")
		m, err := backup.ReadManifest(manifestPath)
		if err == nil {
			manifests = append(manifests, m)
		}
	}

	// Sort by CreatedAt descending
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].CreatedAt.After(manifests[j].CreatedAt)
	})

	return manifests
}

func (m MainModel) UpdateBackups(msg tea.Msg) (MainModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.currentScreen {
		case ScreenBackups:
			switch msg.String() {
			case "esc", "q":
				m.currentScreen = ScreenWelcome
				return m, nil
			case "up", "k":
				if m.SelectedBackup > 0 {
					m.SelectedBackup--
					if m.SelectedBackup < m.BackupScroll {
						m.BackupScroll = m.SelectedBackup
					}
				}
			case "down", "j":
				if m.SelectedBackup < len(m.Backups)-1 {
					m.SelectedBackup++
					if m.SelectedBackup >= m.BackupScroll+10 {
						m.BackupScroll = m.SelectedBackup - 9
					}
				}
			case "d":
				if len(m.Backups) > 0 {
					m.currentScreen = ScreenDeleteConfirm
				}
			case "p":
				if len(m.Backups) > 0 {
					sel := m.Backups[m.SelectedBackup]
					_ = backup.TogglePin(sel)
					m.Backups = ListBackups()
				}
			case "enter", " ":
				if len(m.Backups) > 0 {
					m.currentScreen = ScreenRestoreConfirm
				}
			}

		case ScreenDeleteConfirm:
			switch msg.String() {
			case "y", "s":
				sel := m.Backups[m.SelectedBackup]
				_ = backup.DeleteBackup(sel)
				m.Backups = ListBackups()
				if m.SelectedBackup >= len(m.Backups) && m.SelectedBackup > 0 {
					m.SelectedBackup--
				}
				m.currentScreen = ScreenBackups
			case "n", "esc":
				m.currentScreen = ScreenBackups
			}

		case ScreenRestoreConfirm:
			switch msg.String() {
			case "y", "s":
				sel := m.Backups[m.SelectedBackup]
				m.currentScreen = ScreenBackupResult
				m.RestoreMsg = "Restaurando..."
				m.BackupErr = nil
				return m, func() tea.Msg {
					err := backup.RestoreService{}.Restore(sel)
					return RestoreDoneMsg{Err: err}
				}
			case "n", "esc":
				m.currentScreen = ScreenBackups
			}

		case ScreenBackupResult:
			switch msg.String() {
			case "enter", "esc", "q":
				m.currentScreen = ScreenBackups
				m.Backups = ListBackups()
			}
		}

	case RestoreDoneMsg:
		m.BackupErr = msg.Err
		if msg.Err == nil {
			m.RestoreMsg = "Restauración completada con éxito."
		} else {
			m.RestoreMsg = fmt.Sprintf("Error al restaurar: %v", msg.Err)
		}
	}
	return m, nil
}

func (m MainModel) ViewBackups() string {
	var b strings.Builder
	b.WriteString("\n\n")

	switch m.currentScreen {
	case ScreenBackups:
		b.WriteString("  [ Gestión de Backups ]\n\n")
		if len(m.Backups) == 0 {
			b.WriteString("  No se encontraron backups.\n")
		} else {
			start := m.BackupScroll
			end := start + 10
			if end > len(m.Backups) {
				end = len(m.Backups)
			}
			for i := start; i < end; i++ {
				bk := m.Backups[i]
				cursor := "  "
				if m.SelectedBackup == i {
					cursor = "> "
				}
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, bk.DisplayLabel()))
			}
			if len(m.Backups) > 10 {
				b.WriteString(fmt.Sprintf("\n  Mostrando %d-%d de %d backups\n", start+1, end, len(m.Backups)))
			}
		}
		b.WriteString("\n  [↑/↓] Navegar   [Enter] Restaurar   [d] Eliminar   [p] Pin/Unpin   [Esc] Volver\n")

	case ScreenRestoreConfirm:
		b.WriteString("  [ Confirmar Restauración ]\n\n")
		sel := m.Backups[m.SelectedBackup]
		b.WriteString(fmt.Sprintf("  ¿Estás seguro de restaurar el backup?\n  %s\n\n", sel.DisplayLabel()))
		b.WriteString("  [y/s] Sí   [n/Esc] No\n")

	case ScreenDeleteConfirm:
		b.WriteString("  [ Confirmar Eliminación ]\n\n")
		sel := m.Backups[m.SelectedBackup]
		b.WriteString(fmt.Sprintf("  ¿Estás seguro de eliminar este backup?\n  %s\n\n", sel.DisplayLabel()))
		b.WriteString("  [y/s] Sí   [n/Esc] No\n")

	case ScreenBackupResult:
		b.WriteString("  [ Resultado de Restauración ]\n\n")
		b.WriteString(fmt.Sprintf("  %s\n\n", m.RestoreMsg))
		b.WriteString("  [Enter/Esc] Volver\n")
	}

	return b.String()
}

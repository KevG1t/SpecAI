package steps

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/templates"
)

// SyncContext holds shared state for the sync pipeline steps.
type SyncContext struct {
	IDEs    []system.IDEAdapter
	HomeDir string
	Changed int // populated by sync steps: files refreshed
	Skipped int // populated by sync steps: files already up to date
}

// NewSyncContext creates a SyncContext with the current user's home directory.
func NewSyncContext() (*SyncContext, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el directorio home: %w", err)
	}
	return &SyncContext{HomeDir: homeDir}, nil
}

// StepScanIDEsForSync scans for installed IDEs and populates the context.
type StepScanIDEsForSync struct {
	ctx *SyncContext
}

// NewStepScanIDEsForSync returns a new StepScanIDEsForSync.
func NewStepScanIDEsForSync(ctx *SyncContext) *StepScanIDEsForSync {
	return &StepScanIDEsForSync{ctx: ctx}
}

// ID implements pipeline.Step.
func (s *StepScanIDEsForSync) ID() string {
	return "Detectando IDEs instalados"
}

// Run implements pipeline.Step.
func (s *StepScanIDEsForSync) Run() error {
	ides, err := system.DetectInstalledIDEs()
	if err != nil {
		return err
	}
	if len(ides) == 0 {
		return fmt.Errorf("no se detectaron IDEs instalados")
	}
	s.ctx.IDEs = ides
	return nil
}

// StepSnapshotBeforeSync creates a backup of all files that sync will touch.
type StepSnapshotBeforeSync struct {
	ctx *SyncContext
}

// NewStepSnapshotBeforeSync returns a new StepSnapshotBeforeSync.
func NewStepSnapshotBeforeSync(ctx *SyncContext) *StepSnapshotBeforeSync {
	return &StepSnapshotBeforeSync{ctx: ctx}
}

// ID implements pipeline.Step.
func (s *StepSnapshotBeforeSync) ID() string {
	return "Creando backup previo al sync"
}

// Run implements pipeline.Step. Backup failure is non-fatal.
func (s *StepSnapshotBeforeSync) Run() error {
	var paths []string

	embeddedSkills, err := templates.WalkFiles("base/skills")
	if err != nil {
		log.Printf("sync backup: enumerating embedded skills: %v", err)
	}

	for _, ide := range s.ctx.IDEs {
		// Rules file
		rulesDir := ide.GlobalRulesDir(s.ctx.HomeDir)
		var rulePath string
		if ide.Name() == "Cursor" {
			rulePath = filepath.Join(rulesDir, "specai.mdc")
		} else {
			rulePath = filepath.Join(rulesDir, "specai.md")
		}
		paths = append(paths, rulePath)

		// Skill files
		skillsDir := ide.GlobalSkillsDir(s.ctx.HomeDir)
		for _, ef := range embeddedSkills {
			paths = append(paths, filepath.Join(skillsDir, ef.RelPath))
		}
	}

	snapshotDir := filepath.Join(
		s.ctx.HomeDir,
		".specai", "backups",
		time.Now().UTC().Format("20060102-150405"),
	)

	snapshotter := backup.NewSnapshotter()
	if _, err := snapshotter.Create(snapshotDir, paths); err != nil {
		log.Printf("sync backup: no se pudo crear el snapshot (no bloqueante): %v", err)
	}
	return nil
}

// StepSyncGlobalRules syncs the embedded persona rules to global IDE config dirs.
// It writes only when the on-disk content differs from the embedded content.
type StepSyncGlobalRules struct {
	ctx *SyncContext
}

// NewStepSyncGlobalRules returns a new StepSyncGlobalRules.
func NewStepSyncGlobalRules(ctx *SyncContext) *StepSyncGlobalRules {
	return &StepSyncGlobalRules{ctx: ctx}
}

// ID implements pipeline.Step.
func (s *StepSyncGlobalRules) ID() string {
	return "Sincronizando reglas globales"
}

// Run implements pipeline.Step.
func (s *StepSyncGlobalRules) Run() error {
	if len(s.ctx.IDEs) == 0 {
		return fmt.Errorf("no se detectaron IDEs para sincronizar las reglas globales")
	}

	embedded := []byte(templates.MustRead("base/persona.md"))

	for _, ide := range s.ctx.IDEs {
		rulesDir := ide.GlobalRulesDir(s.ctx.HomeDir)

		var rulePath string
		if ide.Name() == "Cursor" {
			rulePath = filepath.Join(rulesDir, "specai.mdc")
		} else {
			rulePath = filepath.Join(rulesDir, "specai.md")
		}

		existing, readErr := os.ReadFile(rulePath)
		if readErr == nil && bytes.Equal(existing, embedded) {
			s.ctx.Skipped++
			continue
		}
		// Missing file or content differs — write it.
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear el directorio de reglas %s: %w", rulesDir, err)
		}
		if err := os.WriteFile(rulePath, embedded, 0644); err != nil {
			return fmt.Errorf("no se pudo escribir la regla en %s: %w", rulePath, err)
		}
		s.ctx.Changed++
	}
	return nil
}

// StepSyncGlobalSkills syncs embedded skill files to global IDE config dirs.
// It writes only when the on-disk content differs from the embedded content.
type StepSyncGlobalSkills struct {
	ctx *SyncContext
}

// NewStepSyncGlobalSkills returns a new StepSyncGlobalSkills.
func NewStepSyncGlobalSkills(ctx *SyncContext) *StepSyncGlobalSkills {
	return &StepSyncGlobalSkills{ctx: ctx}
}

// ID implements pipeline.Step.
func (s *StepSyncGlobalSkills) ID() string {
	return "Sincronizando skills globales"
}

// Run implements pipeline.Step.
func (s *StepSyncGlobalSkills) Run() error {
	if len(s.ctx.IDEs) == 0 {
		return fmt.Errorf("no se detectaron IDEs para sincronizar los skills globales")
	}

	embeddedSkills, err := templates.WalkFiles("base/skills")
	if err != nil {
		return fmt.Errorf("no se pudieron enumerar los skills embebidos: %w", err)
	}

	for _, ide := range s.ctx.IDEs {
		skillsDir := ide.GlobalSkillsDir(s.ctx.HomeDir)

		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear el directorio de skills %s: %w", skillsDir, err)
		}

		for _, ef := range embeddedSkills {
			destPath := filepath.Join(skillsDir, ef.RelPath)

			existing, readErr := os.ReadFile(destPath)
			if readErr == nil && bytes.Equal(existing, ef.Content) {
				s.ctx.Skipped++
				continue
			}
			// Missing or outdated — write it.
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("no se pudo crear el directorio para el skill %s: %w", destPath, err)
			}
			if err := os.WriteFile(destPath, ef.Content, 0644); err != nil {
				return fmt.Errorf("no se pudo escribir el skill %s: %w", destPath, err)
			}
			s.ctx.Changed++
		}
	}
	return nil
}

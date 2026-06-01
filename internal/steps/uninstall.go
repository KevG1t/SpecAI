package steps

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/templates"
)

type UninstallContext struct {
	IDEs    []system.IDEAdapter
	HomeDir string
}

func NewUninstallContext() (*UninstallContext, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el directorio home: %w", err)
	}
	return &UninstallContext{
		HomeDir: homeDir,
	}, nil
}

// StepScanIDEsForUninstall detects global IDEs.
type StepScanIDEsForUninstall struct {
	ctx *UninstallContext
}

func NewStepScanIDEsForUninstall(ctx *UninstallContext) *StepScanIDEsForUninstall {
	return &StepScanIDEsForUninstall{ctx: ctx}
}

func (s *StepScanIDEsForUninstall) ID() string {
	return "Detectando IDEs instalados para desinstalación"
}

func (s *StepScanIDEsForUninstall) Run() error {
	ides, err := system.DetectInstalledIDEs()
	if err != nil {
		return err
	}
	s.ctx.IDEs = ides
	return nil
}

// StepSnapshotBeforeUninstall creates a snapshot of the files to be deleted.
type StepSnapshotBeforeUninstall struct {
	ctx *UninstallContext
}

func NewStepSnapshotBeforeUninstall(ctx *UninstallContext) *StepSnapshotBeforeUninstall {
	return &StepSnapshotBeforeUninstall{ctx: ctx}
}

func (s *StepSnapshotBeforeUninstall) ID() string {
	return "Creando snapshot previo a la desinstalación"
}

func (s *StepSnapshotBeforeUninstall) Run() error {
	if len(s.ctx.IDEs) == 0 {
		return nil // Nothing to snapshot
	}

	var paths []string

	embeddedSkills, err := templates.WalkFiles("base/skills")
	if err != nil {
		log.Printf("uninstall backup: enumerating embedded skills: %v", err)
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
		for _, sf := range embeddedSkills {
			paths = append(paths, filepath.Join(skillsDir, sf.RelPath))
		}
	}

	snapshotDir := filepath.Join(
		s.ctx.HomeDir,
		".specai", "backups",
		time.Now().UTC().Format("20060102-150405"),
	)

	snapshotter := backup.NewSnapshotter()
	manifest, err := snapshotter.Create(snapshotDir, paths)
	if err != nil {
		log.Printf("uninstall backup: no se pudo crear el snapshot (no bloqueante): %v", err)
		return nil
	}

	// Update the source to indicate this was an uninstall backup
	manifest.Source = backup.BackupSourceUninstall
	manifestPath := filepath.Join(snapshotDir, backup.ManifestFilename)
	if err := backup.WriteManifest(manifestPath, manifest); err != nil {
		log.Printf("uninstall backup: no se pudo actualizar el origen del manifiesto: %v", err)
	}

	return nil
}

// StepRemoveGlobalRules deletes the global rules for each IDE.
type StepRemoveGlobalRules struct {
	ctx *UninstallContext
}

func NewStepRemoveGlobalRules(ctx *UninstallContext) *StepRemoveGlobalRules {
	return &StepRemoveGlobalRules{ctx: ctx}
}

func (s *StepRemoveGlobalRules) ID() string {
	return "Eliminando reglas globales de SpecAI"
}

func (s *StepRemoveGlobalRules) Run() error {
	for _, ide := range s.ctx.IDEs {
		rulesDir := ide.GlobalRulesDir(s.ctx.HomeDir)
		var rulePath string
		if ide.Name() == "Cursor" {
			rulePath = filepath.Join(rulesDir, "specai.mdc")
		} else {
			rulePath = filepath.Join(rulesDir, "specai.md")
		}

		err := os.Remove(rulePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("no se pudo eliminar la regla %s: %w", rulePath, err)
		}
	}
	return nil
}

// StepRemoveGlobalSkills deletes the embedded skill files from each IDE's global skills directory.
type StepRemoveGlobalSkills struct {
	ctx *UninstallContext
}

func NewStepRemoveGlobalSkills(ctx *UninstallContext) *StepRemoveGlobalSkills {
	return &StepRemoveGlobalSkills{ctx: ctx}
}

func (s *StepRemoveGlobalSkills) ID() string {
	return "Eliminando skills globales de SpecAI"
}

func (s *StepRemoveGlobalSkills) Run() error {
	embeddedSkills, err := templates.WalkFiles("base/skills")
	if err != nil {
		return fmt.Errorf("error al leer skills embebidos: %w", err)
	}

	for _, ide := range s.ctx.IDEs {
		skillsDir := ide.GlobalSkillsDir(s.ctx.HomeDir)
		for _, sf := range embeddedSkills {
			targetPath := filepath.Join(skillsDir, sf.RelPath)
			err := os.Remove(targetPath)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("no se pudo eliminar el skill %s: %w", targetPath, err)
			}
		}

		// Try to remove the directory if it's empty
		// os.Remove on a directory only succeeds if it's empty
		_ = os.Remove(skillsDir)
	}
	return nil
}

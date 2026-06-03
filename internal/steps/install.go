package steps

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/KevG1t/SpecAI/internal/assets"
	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/model"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/templates"
)

type InstallContext struct {
	IDEs      []system.IDEAdapter
	TargetIDE system.IDEAdapter // Used for local setup
	HomeDir   string
	Persona   model.PersonaID // Persona selected by the user; defaults to PersonaArgentina

	// Model assignment maps persisted from the picker screens.
	ClaudeModelAssignments map[string]string // phase → alias (opus|sonnet|haiku)
	KiroModelAssignments   map[string]string // phase → alias

	// OpenCode-specific configuration.
	SDDMode   string // "single" | "multi"
	StrictTDD bool   // strict TDD mode enabled

	// Preset holds the ecosystem preset chosen by the user.
	Preset model.PresetID

	// ModelAssignments maps SDD phase names to OpenCode provider/model assignments.
	ModelAssignments map[string]model.ModelAssignment

	// SelectedComponents holds the component IDs confirmed by the user
	// (after dependency tree screen). Zero value is safe for existing consumers.
	SelectedComponents []model.ComponentID
}

func NewInstallContext() (*InstallContext, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &InstallContext{
		HomeDir: homeDir,
		Persona: model.PersonaArgentina,
	}, nil
}

// StepScanGlobalIDEs scans for installed IDEs and populates the context.
type StepScanGlobalIDEs struct {
	ctx *InstallContext
}

func NewStepScanGlobalIDEs(ctx *InstallContext) *StepScanGlobalIDEs {
	return &StepScanGlobalIDEs{ctx: ctx}
}

func (s *StepScanGlobalIDEs) ID() string {
	return "Detectando IDEs instalados en el sistema"
}

func (s *StepScanGlobalIDEs) Run() error {
	ides, err := system.DetectInstalledIDEs()
	if err != nil {
		return err
	}
	s.ctx.IDEs = ides
	return nil
}

// StepSnapshotBeforeInstall creates a backup of the files that install will
// overwrite. Failure is non-fatal — logged and skipped so the install proceeds.
type StepSnapshotBeforeInstall struct {
	ctx *InstallContext
}

func NewStepSnapshotBeforeInstall(ctx *InstallContext) *StepSnapshotBeforeInstall {
	return &StepSnapshotBeforeInstall{ctx: ctx}
}

func (s *StepSnapshotBeforeInstall) ID() string {
	return "Creando backup previo a la instalación"
}

func (s *StepSnapshotBeforeInstall) Run() error {
	var paths []string
	for _, ide := range s.ctx.IDEs {
		if f := ide.SystemPromptFile(s.ctx.HomeDir); f != "" {
			paths = append(paths, f)
		}
		if d := ide.SkillsDir(s.ctx.HomeDir); d != "" {
			paths = append(paths, d)
		}
		if d := ide.SubAgentsDir(s.ctx.HomeDir); d != "" {
			paths = append(paths, d)
		}
		if d := ide.CommandsDir(s.ctx.HomeDir); d != "" {
			paths = append(paths, d)
		}
	}

	snapshotDir := filepath.Join(
		s.ctx.HomeDir,
		".specai", "backups",
		time.Now().UTC().Format("20060102-150405"),
	)

	snapshotter := backup.NewSnapshotter()
	if _, err := snapshotter.Create(snapshotDir, paths); err != nil {
		log.Printf("install backup: no se pudo crear el snapshot (no bloqueante): %v", err)
	}
	return nil
}

// personaReadFileFS is an interface for reading persona files. It allows injection
// of a test double in unit tests without touching the real filesystem.
type personaReadFileFS interface {
	ReadFile(name string) ([]byte, error)
}

// resolvePersona implements the 3-tier persona fallback chain:
//  1. {adapter.AssetFolder()}/persona-argentina.md
//  2. generic/persona-argentina.md
//  3. generic/persona-neutral.md
//  4. Return error if none found
func resolvePersona(pFS personaReadFileFS, ide system.IDEAdapter) ([]byte, error) {
	candidates := []string{
		ide.AssetFolder() + "/persona-argentina.md",
		"generic/persona-argentina.md",
		"generic/persona-neutral.md",
	}

	for _, candidate := range candidates {
		data, err := pFS.ReadFile(candidate)
		if err == nil {
			return data, nil
		}
		if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
			continue
		}
		// Unexpected error
		return nil, fmt.Errorf("error reading persona file %s: %w", candidate, err)
	}

	return nil, fmt.Errorf(
		"no persona file found for agent %q (tried %v)",
		ide.AssetFolder(), candidates,
	)
}

// resolvePersonaNeutral resolves the neutral persona using a 2-tier fallback chain:
//  1. {adapter.AssetFolder()}/persona-neutral.md
//  2. generic/persona-neutral.md
func resolvePersonaNeutral(pFS personaReadFileFS, ide system.IDEAdapter) ([]byte, error) {
	candidates := []string{
		ide.AssetFolder() + "/persona-neutral.md",
		"generic/persona-neutral.md",
	}

	for _, candidate := range candidates {
		data, err := pFS.ReadFile(candidate)
		if err == nil {
			return data, nil
		}
		if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
			continue
		}
		return nil, fmt.Errorf("error reading persona file %s: %w", candidate, err)
	}

	return nil, fmt.Errorf(
		"no neutral persona file found for agent %q (tried %v)",
		ide.AssetFolder(), candidates,
	)
}

// defaultPersonaFS wraps assets.FS (the centralized embedded FS) with a fallback
// to the templates FS (base/persona.md).
type defaultPersonaFS struct{}

func (d defaultPersonaFS) ReadFile(name string) ([]byte, error) {
	// First try the centralized embedded assets FS (internal/assets).
	data, err := assets.FS.ReadFile(name)
	if err == nil {
		return data, nil
	}

	// Fallback: if looking for the neutral persona, use the base template.
	data, templateErr := templates.FS.ReadFile("base/persona.md")
	if templateErr == nil {
		return data, nil
	}

	return nil, err // return original error
}

// StepInstallGlobalRules installs rules for all detected IDEs.
type StepInstallGlobalRules struct {
	ctx     *InstallContext
	personaFS personaReadFileFS
}

func NewStepInstallGlobalRules(ctx *InstallContext) *StepInstallGlobalRules {
	return &StepInstallGlobalRules{ctx: ctx, personaFS: defaultPersonaFS{}}
}

func (s *StepInstallGlobalRules) ID() string {
	return "Inyectando reglas globales de SpecAI"
}

func (s *StepInstallGlobalRules) Run() error {
	if len(s.ctx.IDEs) == 0 {
		return fmt.Errorf("no se detectaron IDEs para instalar las reglas globales")
	}

	for _, ide := range s.ctx.IDEs {
		// PersonaCustom means the user manages their own persona — skip injection.
		if s.ctx.Persona == model.PersonaCustom {
			continue
		}

		var persona []byte
		var err error
		if s.ctx.Persona == model.PersonaNeutral {
			persona, err = resolvePersonaNeutral(s.personaFS, ide)
		} else {
			persona, err = resolvePersona(s.personaFS, ide)
		}
		if err != nil {
			return fmt.Errorf("failed to resolve persona for %s: %w", ide.Name(), err)
		}

		rulesDir := ide.GlobalRulesDir(s.ctx.HomeDir)

		// Ensure dir exists
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear el directorio de reglas %s: %w", rulesDir, err)
		}

		// Cursor rule file is specai.mdc
		var rulePath string
		if ide.Name() == "Cursor" {
			rulePath = filepath.Join(rulesDir, "specai.mdc")
		} else {
			// fallback to standard name
			rulePath = filepath.Join(rulesDir, "specai.md")
		}

		if err := os.WriteFile(rulePath, persona, 0644); err != nil {
			return fmt.Errorf("no se pudo escribir la regla en %s: %w", rulePath, err)
		}
	}
	return nil
}

// StepSetupLocalRules injects rules and skills inside the current working directory.
type StepSetupLocalRules struct {
	ctx *InstallContext
}

func NewStepSetupLocalRules(ctx *InstallContext) *StepSetupLocalRules {
	return &StepSetupLocalRules{ctx: ctx}
}

func (s *StepSetupLocalRules) ID() string {
	if s.ctx.TargetIDE != nil {
		return fmt.Sprintf("Configurando reglas locales para %s", s.ctx.TargetIDE.Name())
	}
	return "Configurando reglas locales"
}

func (s *StepSetupLocalRules) Run() error {
	var targets []system.IDEAdapter
	if s.ctx.TargetIDE != nil {
		targets = append(targets, s.ctx.TargetIDE)
	} else {
		targets = s.ctx.IDEs
	}

	if len(targets) == 0 {
		return fmt.Errorf("no se especificó ni detectó ningún IDE para el setup local")
	}

	persona := templates.MustRead("base/persona.md")

	for _, ide := range targets {
		// 1. Write local rules file
		rulesFile := ide.LocalRulesFile()
		if err := os.WriteFile(rulesFile, []byte(persona), 0644); err != nil {
			return fmt.Errorf("no se pudo escribir el archivo de reglas locales %s: %w", rulesFile, err)
		}

		// 2. Write local skills directory
		skillsDir := ide.LocalSkillsDir()
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear el directorio de skills local %s: %w", skillsDir, err)
		}

		if err := templates.CopyEmbeddedDir("base/skills", skillsDir); err != nil {
			return fmt.Errorf("error al copiar los skills locales a %s: %w", skillsDir, err)
		}
	}

	return nil
}

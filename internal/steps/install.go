package steps

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/templates"
)

type InstallContext struct {
	IDEs      []system.IDEAdapter
	TargetIDE system.IDEAdapter // Used for local setup
	HomeDir   string
}

func NewInstallContext() (*InstallContext, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &InstallContext{
		HomeDir: homeDir,
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

// StepInstallGlobalRules installs rules for all detected IDEs.
type StepInstallGlobalRules struct {
	ctx *InstallContext
}

func NewStepInstallGlobalRules(ctx *InstallContext) *StepInstallGlobalRules {
	return &StepInstallGlobalRules{ctx: ctx}
}

func (s *StepInstallGlobalRules) ID() string {
	return "Inyectando reglas globales de SpecAI"
}

func (s *StepInstallGlobalRules) Run() error {
	if len(s.ctx.IDEs) == 0 {
		return fmt.Errorf("no se detectaron IDEs para instalar las reglas globales")
	}

	persona := templates.MustRead("base/persona.md")

	for _, ide := range s.ctx.IDEs {
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

		if err := os.WriteFile(rulePath, []byte(persona), 0644); err != nil {
			return fmt.Errorf("no se pudo escribir la regla en %s: %w", rulePath, err)
		}
	}
	return nil
}

// StepInstallGlobalSkills installs skills for all detected IDEs.
type StepInstallGlobalSkills struct {
	ctx *InstallContext
}

func NewStepInstallGlobalSkills(ctx *InstallContext) *StepInstallGlobalSkills {
	return &StepInstallGlobalSkills{ctx: ctx}
}

func (s *StepInstallGlobalSkills) ID() string {
	return "Inyectando skills globales de SpecAI"
}

func (s *StepInstallGlobalSkills) Run() error {
	if len(s.ctx.IDEs) == 0 {
		return fmt.Errorf("no se detectaron IDEs para instalar los skills globales")
	}

	for _, ide := range s.ctx.IDEs {
		skillsDir := ide.GlobalSkillsDir(s.ctx.HomeDir)
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("no se pudo crear el directorio de skills %s: %w", skillsDir, err)
		}

		// Copy skills folder from embed FS
		if err := templates.CopyEmbeddedDir("base/skills", skillsDir); err != nil {
			return fmt.Errorf("error al copiar los skills globales a %s: %w", skillsDir, err)
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

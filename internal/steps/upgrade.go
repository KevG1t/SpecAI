package steps

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/update"
)

// UpgradeContext holds shared state for the upgrade pipeline steps.
type UpgradeContext struct {
	CurrentVersion string
	Profile        system.PlatformProfile
	Results        []update.UpdateResult
}

// NewUpgradeContext builds an UpgradeContext by detecting the current platform profile.
func NewUpgradeContext(currentVersion string) (*UpgradeContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	detection, err := system.Detect(ctx)
	if err != nil {
		return nil, fmt.Errorf("no se pudo detectar el perfil del sistema: %w", err)
	}

	return &UpgradeContext{
		CurrentVersion: currentVersion,
		Profile:        detection.System.Profile,
	}, nil
}

// StepCheckForUpdates checks all registered tools for available updates.
type StepCheckForUpdates struct {
	ctx *UpgradeContext
}

// NewStepCheckForUpdates returns a new StepCheckForUpdates.
func NewStepCheckForUpdates(ctx *UpgradeContext) *StepCheckForUpdates {
	return &StepCheckForUpdates{ctx: ctx}
}

// ID implements pipeline.Step.
func (s *StepCheckForUpdates) ID() string {
	return "Verificando actualizaciones disponibles"
}

// Run implements pipeline.Step. It fetches update results for all tools.
// Returns an error only when every single check failed; partial failures are
// tolerated so that one unreachable GitHub endpoint does not block the whole upgrade.
func (s *StepCheckForUpdates) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.ctx.Results = update.CheckAll(ctx, s.ctx.CurrentVersion, s.ctx.Profile)

	// Count failures; only error out when ALL checks failed.
	failed := 0
	for _, r := range s.ctx.Results {
		if r.Status == update.CheckFailed {
			failed++
		}
	}

	if len(s.ctx.Results) > 0 && failed == len(s.ctx.Results) {
		return fmt.Errorf("no se pudo verificar ninguna actualización: todos los checks fallaron")
	}

	return nil
}

// StepInstallUpdates installs all tools that have an available update.
type StepInstallUpdates struct {
	ctx *UpgradeContext
}

// NewStepInstallUpdates returns a new StepInstallUpdates.
func NewStepInstallUpdates(ctx *UpgradeContext) *StepInstallUpdates {
	return &StepInstallUpdates{ctx: ctx}
}

// ID implements pipeline.Step.
func (s *StepInstallUpdates) ID() string {
	return "Instalando actualizaciones"
}

// Run implements pipeline.Step. It upgrades every tool with UpdateAvailable status.
// Errors are collected across all tools (ContinueOnError semantics) and returned
// as a combined error at the end.
func (s *StepInstallUpdates) Run() error {
	var errs []string

	for _, result := range s.ctx.Results {
		if result.Status != update.UpdateAvailable {
			continue
		}

		if err := runUpgrade(result); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", result.Tool.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errores al instalar actualizaciones:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

// runUpgrade dispatches the upgrade to the correct strategy based on InstallMethod.
func runUpgrade(result update.UpdateResult) error {
	switch result.Tool.InstallMethod {
	case update.InstallOpenCodePlugin:
		// OpenCode manages its own plugin lifecycle on restart/reload; skip silently.
		return nil

	case update.InstallGoInstall:
		importPath := result.Tool.GoImportPath
		if importPath == "" {
			return fmt.Errorf("GoImportPath no configurado para %s", result.Tool.Name)
		}
		target := importPath + "@latest"
		cmd := exec.Command("go", "install", target) //nolint:gosec
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go install %s falló: %w", target, err)
		}
		return nil

	case update.InstallBrew:
		cmd := exec.Command("brew", "upgrade", result.Tool.Name) //nolint:gosec
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("brew upgrade %s falló: %w", result.Tool.Name, err)
		}
		return nil

	case update.InstallScript:
		hint := result.UpdateHint
		if hint == "" {
			return fmt.Errorf("no hay script de instalación configurado para %s", result.Tool.Name)
		}
		cmd := exec.Command("sh", "-c", hint) //nolint:gosec
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("script de instalación para %s falló: %w", result.Tool.Name, err)
		}
		return nil

	case update.InstallBinary:
		// TODO: binary download not yet implemented.
		// The release page is available at result.ReleaseURL.
		return fmt.Errorf(
			"descarga de binario no implementada aún para %s — descargá manualmente desde %s",
			result.Tool.Name,
			result.ReleaseURL,
		)

	default:
		return fmt.Errorf("método de instalación desconocido %q para %s", result.Tool.InstallMethod, result.Tool.Name)
	}
}

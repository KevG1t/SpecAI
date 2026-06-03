package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/KevG1t/SpecAI/internal/pipeline"
	"github.com/KevG1t/SpecAI/internal/skillregistry"
	"github.com/KevG1t/SpecAI/internal/steps"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/tui"
)

// version is set by GoReleaser via ldflags at build time.
var version = "dev"

// ResolveVersion determines the actual version, falling back to build info.
func ResolveVersion(v string) string {
	if v != "dev" && v != "" {
		return v
	}
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	v := ResolveVersion(version)
	tui.SetVersion(v)

	if len(os.Args) > 1 {
		if len(os.Args) > 2 && os.Args[1] == "skill-registry" && os.Args[2] == "refresh" {
			force := len(os.Args) > 3 && os.Args[3] == "--force"
			cwd, _ := os.Getwd()
			homeDir, _ := os.UserHomeDir()
			result, err := skillregistry.Regenerate(cwd, homeDir, force)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if result.Regenerated {
				fmt.Printf("Registry regenerated: %d skills → %s\n", result.SkillCount, result.Registry)
			} else {
				fmt.Printf("Registry up to date (%s)\n", result.Reason)
			}
			os.Exit(0)
		} else if os.Args[1] == "setup" {
			if len(os.Args) < 3 {
				fmt.Println("Error: Debes especificar un IDE (ej: specai setup cursor)")
				fmt.Println("IDEs soportados: cursor, windsurf, codex")
				os.Exit(1)
			}
			ideName := os.Args[2]
			err := runLocalSetup(ideName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error en el setup local: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("¡Setup local de %s completado con éxito!\n", ideName)
			os.Exit(0)
		} else {
			fmt.Println("Uso:")
			fmt.Println("  specai               - Lanza la interfaz interactiva (TUI)")
			fmt.Println("  specai setup <ide>   - Inyecta reglas locales en este repositorio")
			fmt.Println("\nIDEs soportados: cursor, windsurf, codex")
			os.Exit(0)
		}
	}

	err := tui.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runLocalSetup(ideName string) error {
	adapter := system.GetAdapterByName(ideName)
	if adapter == nil {
		return fmt.Errorf("IDE '%s' no soportado. IDEs válidos: cursor, windsurf, codex", ideName)
	}

	ctx, err := steps.NewInstallContext()
	if err != nil {
		return err
	}
	ctx.TargetIDE = adapter

	plan := pipeline.StagePlan{
		Apply: []pipeline.Step{
			steps.NewStepSetupLocalRules(ctx),
		},
	}

	orch := pipeline.NewOrchestrator(
		pipeline.DefaultRollbackPolicy(),
		pipeline.WithFailurePolicy(pipeline.StopOnError),
		pipeline.WithProgressFunc(func(ev pipeline.ProgressEvent) {
			if ev.Status == pipeline.StepStatusRunning {
				fmt.Printf("=> Ejecutando: %s...\n", ev.StepID)
			}
		}),
	)

	res := orch.Execute(plan)
	return res.Err
}

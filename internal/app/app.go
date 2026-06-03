package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/KevG1t/SpecAI/internal/backup"
	"github.com/KevG1t/SpecAI/internal/cli"
	"github.com/KevG1t/SpecAI/internal/skillregistry"
	"github.com/KevG1t/SpecAI/internal/system"
	"github.com/KevG1t/SpecAI/internal/tui"
	"github.com/KevG1t/SpecAI/internal/update"
	"github.com/KevG1t/SpecAI/internal/update/upgrade"
	"github.com/KevG1t/SpecAI/internal/verify"
)

// Version is set from main via ldflags at build time.
var Version = "dev"

var (
	updateCheckAll      = update.CheckAll
	updateCheckFiltered = update.CheckFiltered
	upgradeExecute      = upgrade.Execute
	selfUpdateFn        = selfUpdate
	ensureCurrentOSSupported = system.EnsureCurrentOSSupported
	detectSystem             = system.Detect
)

func Run() error {
	return RunArgs(os.Args[1:], os.Stdout)
}

func RunArgs(args []string, stdout io.Writer) error {
	// Propagate the build-time version to the CLI and upgrade layers so backup
	// manifests record which version of specai created them.
	cli.AppVersion = Version
	upgrade.AppVersion = Version

	// Info commands: no system detection, no self-update, no platform validation.
	if len(args) > 0 {
		switch args[0] {
		case "version", "--version", "-v":
			_, _ = fmt.Fprintf(stdout, "specai %s\n", Version)
			return nil
		case "help", "--help", "-h":
			printHelp(stdout, Version)
			return nil
		case "uninstall":
			_, err := cli.RunUninstall(args[1:], stdout)
			return err
		case "skill-registry":
			return runSkillRegistry(args[1:], stdout)
		}
	}

	if err := ensureCurrentOSSupported(); err != nil {
		return err
	}

	result, err := detectSystem(context.Background())
	if err != nil {
		return fmt.Errorf("detect system: %w", err)
	}

	if !result.System.Supported {
		return system.EnsureSupportedPlatform(result.System.Profile)
	}

	var (
		profile         system.PlatformProfile
		profileResolved bool
	)
	resolveProfile := func() system.PlatformProfile {
		if !profileResolved {
			profile = cli.ResolveInstallProfile(result)
			profileResolved = true
		}
		return profile
	}

	// Self-update: check for a newer specai release and apply it before
	// CLI/TUI dispatch. Errors are non-fatal — logged and swallowed.
	// Skip auto-upgrade on TUI entry (len(args) == 0) to avoid silently
	// replacing the binary while the user expects a clean TUI launch.
	isTUIFlow := len(args) == 0
	if !isTUIFlow && !isExplicitUpdateFlow(args) {
		if err := selfUpdateFn(context.Background(), Version, resolveProfile(), stdout); err != nil {
			_, _ = fmt.Fprintf(stdout, "Warning: self-update failed: %v\n", err)
		}
	}

	if len(args) == 0 {
		// Launch the interactive TUI.
		tui.SetVersion(Version)
		return tui.Start()
	}

	switch args[0] {
	case "update":
		return runUpdate(context.Background(), Version, resolveProfile(), stdout)
	case "upgrade":
		return runUpgrade(context.Background(), args[1:], result, stdout)
	case "install":
		installResult, err := cli.RunInstall(args[1:], result)
		if err != nil {
			return err
		}

		if installResult.DryRun {
			_, _ = fmt.Fprintln(stdout, cli.RenderDryRun(installResult))
		} else {
			_, _ = fmt.Fprint(stdout, verify.RenderReport(installResult.Verify))
		}

		return nil
	case "sync":
		syncResult, err := cli.RunSync(args[1:])
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(stdout, cli.RenderSyncReport(syncResult))
		return nil
	case "uninstall":
		uninstallResult, err := cli.RunUninstall(args[1:], stdout)
		if err != nil {
			// If a backup was created before the failure, surface it so
			// the user can restore safely.
			if uninstallResult.Manifest.ID != "" {
				_, _ = fmt.Fprintln(stdout, cli.RenderUninstallReport(uninstallResult))
			}
			return err
		}
		if uninstallResult.Manifest.ID != "" {
			_, _ = fmt.Fprintln(stdout, cli.RenderUninstallReport(uninstallResult))
		}
		return nil
	case "restore":
		return cli.RunRestore(args[1:], stdout)
	case "doctor":
		return cli.RunDoctor(context.Background(), stdout)
	case "backup":
		return cli.RunBackup(args[1:], stdout)
	case "repair":
		return cli.RunRepair(args[1:], stdout)
	case "status":
		return cli.RunStatus(args[1:], stdout)
	default:
		return fmt.Errorf("unknown command %q — run 'specai help' for available commands", args[0])
	}
}

func runSkillRegistry(args []string, stdout io.Writer) error {
	if len(args) == 0 || args[0] != "refresh" {
		return fmt.Errorf("usage: specai skill-registry refresh [--cwd <dir>] [--force] [--quiet] [--no-gitignore]")
	}

	cwd := ""
	force := false
	quiet := false
	ensureGitignore := true
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		case "--quiet", "-q":
			quiet = true
		case "--no-gitignore":
			ensureGitignore = false
		case "--cwd":
			if i+1 >= len(args) {
				return fmt.Errorf("--cwd requires a value")
			}
			cwd = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown skill-registry argument %q", args[i])
		}
	}
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve cwd: %w", err)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}
	if ensureGitignore {
		if err := skillregistry.EnsureATLIgnored(cwd); err != nil {
			return err
		}
	}
	result, err := skillregistry.Regenerate(cwd, home, force)
	if err != nil {
		return err
	}
	if !quiet {
		if result.Regenerated {
			_, _ = fmt.Fprintf(stdout, "Skill registry refreshed (%d skills): %s\n", result.SkillCount, result.Registry)
		} else {
			_, _ = fmt.Fprintf(stdout, "Skill registry up to date (%s): %s\n", result.Reason, result.Registry)
		}
	}
	return nil
}

func runUpdate(ctx context.Context, currentVersion string, profile system.PlatformProfile, stdout io.Writer) error {
	return runUpdateWithArgs(ctx, os.Args[2:], currentVersion, profile, stdout)
}

// runUpdateWithArgs parses --skills and --sdd-memory flags and limits the update scope accordingly.
func runUpdateWithArgs(ctx context.Context, args []string, currentVersion string, profile system.PlatformProfile, stdout io.Writer) error {
	var skillsOnly bool
	var sddMemoryOnly bool

	for _, arg := range args {
		switch arg {
		case "--skills":
			skillsOnly = true
		case "--sdd-memory":
			sddMemoryOnly = true
		}
	}

	if skillsOnly {
		results := updateCheckFiltered(ctx, currentVersion, profile, []string{"skills"})
		_, _ = fmt.Fprint(stdout, update.RenderCLI(results))
		return updateCheckError(results)
	}

	if sddMemoryOnly {
		results := updateCheckFiltered(ctx, currentVersion, profile, []string{"sdd-memory"})
		_, _ = fmt.Fprint(stdout, update.RenderCLI(results))
		return updateCheckError(results)
	}

	results := updateCheckAll(ctx, currentVersion, profile)
	_, _ = fmt.Fprint(stdout, update.RenderCLI(results))
	return updateCheckError(results)
}

// runUpgrade handles the `specai upgrade [--dry-run] [tool...]` command.
func runUpgrade(ctx context.Context, args []string, detection system.DetectionResult, stdout io.Writer) error {
	dryRun := false
	noBackup := false
	var toolFilter []string

	for _, arg := range args {
		switch {
		case arg == "--dry-run" || arg == "-n":
			dryRun = true
		case arg == "--no-backup":
			noBackup = true
		case !strings.HasPrefix(arg, "-"):
			toolFilter = append(toolFilter, arg)
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	profile := cli.ResolveInstallProfile(detection)

	// Check for available updates (filtered to requested tools if specified).
	sp := upgrade.NewSpinner(stdout, "Checking for updates")
	checkResults := updateCheckFiltered(ctx, Version, profile, toolFilter)
	checkErr := updateCheckError(checkResults)
	sp.Finish(checkErr == nil)
	if checkErr != nil {
		_, _ = fmt.Fprint(stdout, update.RenderCLI(checkResults))
		return checkErr
	}

	// Execute upgrades (no-op if nothing is UpdateAvailable).
	report := upgrade.ExecuteWithOptions(ctx, checkResults, profile, homeDir, dryRun, upgrade.ExecuteOptions{
		Progress:          stdout,
		BackupDiagnostics: stdout,
		SkipBackup:        noBackup,
	})

	_, _ = fmt.Fprint(stdout, upgrade.RenderUpgradeReport(report))

	// Return error only if any tool failed (not for skipped/manual).
	var errs []error
	for _, r := range report.Results {
		if r.Status == upgrade.UpgradeFailed && r.Err != nil {
			errs = append(errs, fmt.Errorf("upgrade failed for %q: %w", r.ToolName, r.Err))
		}
	}

	return errors.Join(errs...)
}

func updateCheckError(results []update.UpdateResult) error {
	failed := update.CheckFailures(results)
	if len(failed) == 0 {
		return nil
	}

	return fmt.Errorf("update check failed for: %s", strings.Join(failed, ", "))
}

// ListBackups returns all backup manifests from the backup directory.
func ListBackups() []backup.Manifest {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	backupRoot := filepath.Join(homeDir, ".specai", "backups")
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return nil
	}

	manifests := make([]backup.Manifest, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(backupRoot, entry.Name(), backup.ManifestFilename)
		manifest, err := backup.ReadManifest(manifestPath)
		if err != nil {
			continue
		}
		manifests = append(manifests, manifest)
	}

	// Sort by creation time (newest first) — the IDs are timestamps.
	for i := 0; i < len(manifests); i++ {
		for j := i + 1; j < len(manifests); j++ {
			if manifests[j].CreatedAt.After(manifests[i].CreatedAt) {
				manifests[i], manifests[j] = manifests[j], manifests[i]
			}
		}
	}

	return manifests
}

// isExplicitUpdateFlow reports whether the current invocation is already in the
// explicit update/upgrade path. In those cases, self-update must be skipped to
// avoid preempting the user's requested command behavior.
func isExplicitUpdateFlow(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return args[0] == "update" || args[0] == "upgrade"
}

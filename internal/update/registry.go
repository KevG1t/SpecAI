package update

import (
	"path/filepath"
)

// Tools is the static registry of managed tools that can be checked for updates.
//
// InstallMethod controls which upgrade strategy the executor uses:
//   - InstallBrew: managed via homebrew (macOS/Linux with brew)
//   - InstallGoInstall: installed via `go install <GoImportPath>@version`
//   - InstallBinary: downloaded binary from GitHub Releases (atomic replace)
//
// For brew-managed platforms the executor picks brew regardless of the
// field here; InstallMethod represents the non-brew fallback strategy.
var Tools = []ToolInfo{
	{
		Name:          "specai",
		Owner:         "KevG1t",
		Repo:          "SpecAI",
		DetectCmd:     nil, // version comes from build-time ldflags (app.Version)
		VersionPrefix: "v",
		// specai: brew on macOS, binary release download on Linux/Windows.
		// Self-upgrade of the running binary on Windows is deferred to Phase 2.
		InstallMethod: InstallBinary,
	},
	{
		Name:              "sdd-memory",
		Owner:             "KevG1t",
		Repo:              "sdd-memory",
		DetectCmd:         []string{"sdd-memory", "version"},
		VersionPrefix:     "v",
		ReleaseTagPattern: `^v[0-9]+\.[0-9]+\.[0-9]+$`,
		InstallMethod:     InstallBinary,
		FallbackPaths: func(homeDir, localAppData string) []string {
			var paths []string
			if localAppData != "" {
				paths = append(paths, filepath.Join(localAppData, "sdd-memory", "bin", "sdd-memory.exe"))
			} else if homeDir != "" {
				paths = append(paths, filepath.Join(homeDir, "AppData", "Local", "sdd-memory", "bin", "sdd-memory.exe"))
			}
			if homeDir != "" {
				paths = append(paths, filepath.Join(homeDir, ".local", "bin", "sdd-memory"))
			}
			return paths
		},
	},
	{
		Name:          "gga",
		Owner:         "Gentleman-Programming",
		Repo:          "gentleman-guardian-angel",
		DetectCmd:     []string{"gga", "--version"},
		VersionPrefix: "v",
		// gga: brew on macOS, install.sh script on Linux/Windows.
		// GGA does not publish pre-built release binary assets — only source archives.
		// Using InstallScript runs curl | bash via the project's install.sh.
		InstallMethod: InstallScript,
		// FallbackPaths covers the Windows stale-PATH scenario: gga installs a
		// PowerShell shim to ~/bin/gga.ps1, and the bash script to ~/.local/bin/gga.
		// Both locations may not be in PATH immediately after install.
		FallbackPaths: func(homeDir, localAppData string) []string {
			var paths []string
			if homeDir != "" {
				// Windows: ~/bin/gga.ps1 (PowerShell shim, callable as "gga" in PS)
				paths = append(paths, filepath.Join(homeDir, "bin", "gga.ps1"))
				// Linux/macOS: ~/.local/bin/gga
				paths = append(paths, filepath.Join(homeDir, ".local", "bin", "gga"))
				// Linux/macOS: ~/bin/gga
				paths = append(paths, filepath.Join(homeDir, "bin", "gga"))
			}
			return paths
		},
	},
	{
		Name:          "opencode-subagent-statusline",
		Owner:         "Joaquinvesapa",
		Repo:          "sub-agent-statusline",
		VersionPrefix: "v",
		InstallMethod: InstallOpenCodePlugin,
		NpmPackage:    "opencode-subagent-statusline",
	},
	{
		Name:          "opencode-sdd-memory-manage",
		Owner:         "j0k3r-dev-rgl",
		Repo:          "sdd-engram-plugin",
		VersionPrefix: "v",
		InstallMethod: InstallOpenCodePlugin,
		NpmPackage:    "opencode-sdd-engram-manage",
	},
}

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
		Repo:          "sdd-memory-plugin",
		VersionPrefix: "v",
		InstallMethod: InstallOpenCodePlugin,
		NpmPackage:    "opencode-sdd-memory-manage",
	},
}

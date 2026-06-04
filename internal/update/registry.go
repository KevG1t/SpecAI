package update

import (
	"path/filepath"
)

// Tools is the static registry of managed tools that can be checked for updates.
//
// InstallMethod controls which upgrade strategy the executor uses:
//   - InstallGoInstall: installed via `go install <GoImportPath>@version`
//   - InstallBinary: downloaded binary from GitHub Releases (atomic replace)
var Tools = []ToolInfo{
	{
		Name:          "specai",
		Owner:         "KevG1t",
		Repo:          "SpecAI",
		DetectCmd:     nil, // version comes from build-time ldflags (app.Version)
		VersionPrefix: "v",
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
		// FallbackPaths covers the Windows stale-PATH scenario (and Linux ~/.local/bin
		// when not yet in PATH): AddToUserPath updates the registry/profile but the
		// current process does not see the change until a new shell session starts.
		FallbackPaths: func(homeDir, localAppData string) []string {
			var paths []string
			// Windows: %LOCALAPPDATA%\sdd-memory\bin\sdd-memory.exe
			if localAppData != "" {
				paths = append(paths, filepath.Join(localAppData, "sdd-memory", "bin", "sdd-memory.exe"))
			} else if homeDir != "" {
				// LOCALAPPDATA is not set (e.g. restricted environment or CI on Windows).
				// Derive the standard path from homeDir for parity with the installer.
				paths = append(paths, filepath.Join(homeDir, "AppData", "Local", "sdd-memory", "bin", "sdd-memory.exe"))
			}
			// Linux/macOS: ~/.local/bin/sdd-memory (when /usr/local/bin is not writable,
			// the binary installer places it here, which may not be in PATH yet).
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

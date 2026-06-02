package update

import (
	"fmt"
	"strings"

	"github.com/KevG1t/SpecAI/internal/system"
)

// updateHint returns a platform-specific instruction string for updating the given tool.
func updateHint(tool ToolInfo, profile system.PlatformProfile) string {
	switch tool.Name {
	case "specai":
		return specAIHint(profile)
	case "sdd-memory":
		return sddMemoryHint(profile)
	case "opencode-subagent-statusline", "opencode-sdd-memory-manage":
		return "specai upgrade updates ~/.config/opencode npm deps, clears this plugin's @latest cache, then requires OpenCode restart/reload"
	default:
		return ""
	}
}

func openCodeRegisteredNotMaterializedHint(tool ToolInfo) string {
	pkg := strings.TrimSpace(tool.NpmPackage)
	if pkg == "" {
		pkg = tool.Name
	}
	return fmt.Sprintf("registered in ~/.config/opencode/tui.json; pending npm dependency materialization for %s. Run specai upgrade to install/update ~/.config/opencode dependencies, then restart or reload OpenCode; if it stays pending, check OpenCode logs for package or peer dependency errors.", pkg)
}

func specAIHint(profile system.PlatformProfile) string {
	switch profile.OS {
	case "darwin":
		return "brew upgrade specai"
	case "linux":
		return "curl -fsSL https://raw.githubusercontent.com/KevG1t/SpecAI/main/scripts/install.sh | bash"
	case "windows":
		return "irm https://raw.githubusercontent.com/KevG1t/SpecAI/main/scripts/install.ps1 | iex"
	default:
		return ""
	}
}

func sddMemoryHint(profile system.PlatformProfile) string {
	switch profile.PackageManager {
	case "brew":
		return "brew upgrade sdd-memory"
	default:
		return "specai upgrade (downloads pre-built binary)"
	}
}


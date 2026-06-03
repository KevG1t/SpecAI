package tui

// Screen represents an identifier for the different TUI screens.
type Screen string

const (
	ScreenWelcome        Screen = "welcome"
	ScreenLoading        Screen = "loading" // DEPRECATED: to be replaced by module-specific screens
	ScreenBackups        Screen = "backups"
	ScreenRestoreConfirm Screen = "restoreConfirm"
	ScreenDeleteConfirm  Screen = "deleteConfirm"
	ScreenBackupResult   Screen = "backupResult"
	ScreenInstall           Screen = "install"
	ScreenAgentSelect       Screen = "agentSelect"       // Agent selection before install pipeline
	ScreenPersona           Screen = "persona"           // Persona selection after agent select
	ScreenPreset            Screen = "preset"            // Ecosystem preset selection after persona
	ScreenClaudeModelPicker Screen = "claudeModelPicker" // Claude model assignments (when Claude selected)
	ScreenKiroModelPicker   Screen = "kiroModelPicker"   // Kiro model assignments (when Kiro selected)
	ScreenSDDMode           Screen = "sddMode"           // SDD mode for OpenCode
	ScreenStrictTDD         Screen = "strictTDD"         // Strict TDD for OpenCode+SDD
	ScreenSetupLocal        Screen = "setupLocal"
	ScreenUpgrade        Screen = "upgrade"
	ScreenSync           Screen = "sync"
	ScreenUpgradeSync    Screen = "upgradeSync"
	ScreenUninstall      Screen = "uninstall"
	ScreenComplete       Screen = "complete" // Post-install completion summary

	ScreenDetection           Screen = "detection"          // System detection before agent select
	ScreenReview              Screen = "review"             // Review and confirm before install
	ScreenDependencyTree      Screen = "dependencyTree"     // Component list with auto-dep badges
	ScreenOpenCodeModelPicker Screen = "openCodeModelPicker" // OpenCode 4-level model picker
	ScreenSkillPicker         Screen = "skillPicker"        // Custom preset skill selection
	ScreenMCPPicker           Screen = "mcpPicker"          // MCP server selection (Custom flow only)
	ScreenConfigPicker        Screen = "configPicker"       // Config customization (Custom flow only)
)

// Route defines the transitions from a screen.
type Route struct {
	Forward  Screen
	Backward Screen
}

// linearRoutes defines the standard navigation paths.
var linearRoutes = map[Screen]Route{
	ScreenWelcome: {
		Forward:  ScreenLoading,
		Backward: ScreenWelcome, // no backward from welcome
	},
	ScreenLoading: {
		Forward:  ScreenLoading, // stays here for now
		Backward: ScreenWelcome,
	},
}

// NextScreen returns the next screen in the linear flow.
func NextScreen(current Screen) Screen {
	if route, exists := linearRoutes[current]; exists {
		return route.Forward
	}
	return current
}

// PreviousScreen returns the previous screen in the linear flow.
func PreviousScreen(current Screen) Screen {
	if route, exists := linearRoutes[current]; exists {
		return route.Backward
	}
	return current
}

// ScreenChangeMsg is sent to the event loop to change screens.
type ScreenChangeMsg struct {
	Next Screen
}

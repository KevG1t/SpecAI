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
	ScreenInstall        Screen = "install"
	ScreenAgentSelect    Screen = "agentSelect" // Agent selection before install pipeline
	ScreenSetupLocal     Screen = "setupLocal"
	ScreenUpgrade        Screen = "upgrade"
	ScreenSync           Screen = "sync"
	ScreenUpgradeSync    Screen = "upgradeSync"
	ScreenUninstall      Screen = "uninstall"
	ScreenComplete       Screen = "complete" // Post-install completion summary
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

package gga

import (
	"github.com/KevG1t/specai/internal/installcmd"
	"github.com/KevG1t/specai/internal/model"
	"github.com/KevG1t/specai/internal/system"
)

func InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	return installcmd.NewResolver().ResolveComponentInstall(profile, model.ComponentGGA)
}

func ShouldInstall(enabled bool) bool {
	return enabled
}

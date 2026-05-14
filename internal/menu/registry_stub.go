//go:build !windows

package menu

import (
	"fmt"

	"rightmenu/internal/config"
)

type RegistryInstaller struct{}

func NewRegistryInstaller() RegistryInstaller { return RegistryInstaller{} }

func (RegistryInstaller) Install(config.Config, string) error {
	return fmt.Errorf("registry install is only supported on Windows")
}

func (RegistryInstaller) Uninstall() error {
	return fmt.Errorf("registry uninstall is only supported on Windows")
}

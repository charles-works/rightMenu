package menu

import "rightmenu/internal/config"

type Installer interface {
	Install(config.Config, string) error
	Uninstall() error
}

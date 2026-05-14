package menu

import (
	"fmt"
	"path/filepath"

	"rightmenu/internal/config"
)

const (
	ParentKey           = `Software\Classes\*\shell\DEVDebug`
	ExtendedCommandsKey = ParentKey + `\ExtendedSubCommandsKey\Shell`
	SettingsID          = "settings"
)

type CommandSpec struct {
	ID      string
	Title   string
	Command string
}

type RegistryPlan struct {
	ParentKey string
	Values    map[string]string
	Commands  []CommandSpec
}

func RenderRegistryPlan(cfg config.Config, pinnedExe string) RegistryPlan {
	commands := make([]CommandSpec, 0, len(cfg.Items)+1)
	for _, item := range cfg.Items {
		commands = append(commands, CommandSpec{
			ID:      item.ID,
			Title:   item.Title,
			Command: RenderRunCommand(pinnedExe, item.ID),
		})
	}
	commands = append(commands, CommandSpec{ID: SettingsID, Title: "设置", Command: RenderConfigCommand(pinnedExe)})
	return RegistryPlan{
		ParentKey: ParentKey,
		Values: map[string]string{
			"MUIVerb":          cfg.MenuTitle,
			"Icon":             filepath.Clean(pinnedExe),
			"MultiSelectModel": "Single",
		},
		Commands: commands,
	}
}

func RenderRunCommand(pinnedExe, itemID string) string {
	return fmt.Sprintf("%s run %s %s", quote(pinnedExe), quote(itemID), quote("%1"))
}

func RenderConfigCommand(pinnedExe string) string {
	return fmt.Sprintf("%s config", quote(pinnedExe))
}

func quote(s string) string { return `"` + s + `"` }

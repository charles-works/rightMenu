package menu

import (
	"strings"
	"testing"

	"rightmenu/internal/config"
)

func TestRenderRegistryPlan(t *testing.T) {
	pinned := `%LOCALAPPDATA%\RightMenu\rightmenu.exe`
	cfg := config.Config{MenuTitle: "DEV调试", Items: []config.Item{{ID: "aa", Title: "AA", Program: `C:\Tools\AA.exe`}}}
	plan := RenderRegistryPlan(cfg, pinned)
	if plan.ParentKey != `Software\Classes\*\shell\DEVDebug` {
		t.Fatalf("parent key = %q", plan.ParentKey)
	}
	if plan.Values["MUIVerb"] != "DEV调试" || plan.Values["MultiSelectModel"] != "Single" {
		t.Fatalf("unexpected values: %#v", plan.Values)
	}
	if len(plan.Commands) != 2 {
		t.Fatalf("commands = %#v", plan.Commands)
	}
	if got := plan.Commands[0].Command; got != `"%LOCALAPPDATA%\RightMenu\rightmenu.exe" run "aa" "%1"` {
		t.Fatalf("run command = %q", got)
	}
	if got := plan.Commands[1].Command; got != `"%LOCALAPPDATA%\RightMenu\rightmenu.exe" config` {
		t.Fatalf("settings command = %q", got)
	}
	for _, cmd := range plan.Commands {
		if strings.Contains(cmd.Command, `Software\RightMenu\MenuCommands`) {
			t.Fatalf("forbidden menu store in command: %s", cmd.Command)
		}
	}
}

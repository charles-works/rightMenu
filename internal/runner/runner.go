package runner

import (
	"fmt"
	"os/exec"

	"rightmenu/internal/config"
)

type Starter interface {
	Start(program string, args []string) error
}

type ExecStarter struct{}

func (ExecStarter) Start(program string, args []string) error {
	cmd := exec.Command(program, args...)
	return cmd.Start()
}

func BuildInvocation(cfg config.Config, itemID, file string) (string, []string, error) {
	if file == "" {
		return "", nil, fmt.Errorf("selected file path is required")
	}
	item, ok := cfg.Find(itemID)
	if !ok {
		return "", nil, fmt.Errorf("menu item %q not found", itemID)
	}
	return item.Program, item.ExpandedArgs(file), nil
}

func Run(cfg config.Config, itemID, file string, starter Starter) error {
	program, args, err := BuildInvocation(cfg, itemID, file)
	if err != nil {
		return err
	}
	if starter == nil {
		starter = ExecStarter{}
	}
	if err := starter.Start(program, args); err != nil {
		return fmt.Errorf("start %q: %w", program, err)
	}
	return nil
}

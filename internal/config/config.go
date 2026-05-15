package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	AppDirName     = "RightMenu"
	ConfigFileName = "config.json"
	PinnedExeName  = "rightmenu.exe"
	FileToken      = "{file}"
)

var itemIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// Config is the user-editable menu configuration.
type Config struct {
	MenuTitle string `json:"menuTitle"`
	Items     []Item `json:"items"`
}

// Item describes one DEV调试 submenu command.
type Item struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Program         string   `json:"program"`
	SpecifiedFolder string   `json:"specifiedFolder,omitempty"`
	Args            []string `json:"args,omitempty"`
}

// Paths contains canonical v1 filesystem locations.
type Paths struct {
	ConfigPath string
	InstallDir string
	PinnedExe  string
}

func DefaultPaths() (Paths, error) {
	appData, err := userConfigBase()
	if err != nil {
		return Paths{}, err
	}
	localAppData, err := userLocalBase()
	if err != nil {
		return Paths{}, err
	}
	installDir := filepath.Join(localAppData, AppDirName)
	return Paths{
		ConfigPath: filepath.Join(appData, AppDirName, ConfigFileName),
		InstallDir: installDir,
		PinnedExe:  filepath.Join(installDir, PinnedExeName),
	}, nil
}

func userConfigBase() (string, error) {
	if v := os.Getenv("APPDATA"); v != "" {
		return v, nil
	}
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(home, ".config"), nil
}

func userLocalBase() (string, error) {
	if v := os.Getenv("LOCALAPPDATA"); v != "" {
		return v, nil
	}
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve local app directory: %w", err)
	}
	return filepath.Join(home, ".local", "share"), nil
}

func DefaultConfig() Config {
	return Config{
		MenuTitle: "DEV调试",
		Items: []Item{{
			ID:              "aa",
			Title:           "AA",
			Program:         `C:\Tools\AA.exe`,
			SpecifiedFolder: `C:\DEV`,
		}},
	}
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Ensure(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat config %q: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	b, err := json.MarshalIndent(DefaultConfig(), "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write sample config %q: %w", path, err)
	}
	return nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.MenuTitle) == "" {
		return errors.New("config menuTitle is required")
	}
	if len(c.Items) == 0 {
		return errors.New("config requires at least one item")
	}
	seen := map[string]struct{}{}
	for i, item := range c.Items {
		idx := fmt.Sprintf("items[%d]", i)
		if strings.TrimSpace(item.ID) == "" {
			return fmt.Errorf("%s.id is required", idx)
		}
		if !itemIDPattern.MatchString(item.ID) {
			return fmt.Errorf("%s.id %q must match %s", idx, item.ID, itemIDPattern.String())
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate item id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
		if strings.TrimSpace(item.Title) == "" {
			return fmt.Errorf("%s.title is required", idx)
		}
		if strings.TrimSpace(item.Program) == "" {
			return fmt.Errorf("%s.program is required", idx)
		}
		if len(item.Args) == 0 && strings.TrimSpace(item.SpecifiedFolder) == "" {
			return fmt.Errorf("%s.specifiedFolder is required when args is omitted or empty", idx)
		}
	}
	return nil
}

func (c Config) Find(id string) (Item, bool) {
	for _, item := range c.Items {
		if item.ID == id {
			return item, true
		}
	}
	return Item{}, false
}

func (i Item) ArgsWithDefaults(file string) []string {
	if len(i.Args) == 0 {
		dir, name := splitSelectedFile(file)
		return []string{" ", "IF_A_000N", i.SpecifiedFolder, dir, name, "Q"}
	}
	out := make([]string, len(i.Args))
	copy(out, i.Args)
	return out
}

func (i Item) ExpandedArgs(file string) []string {
	args := i.ArgsWithDefaults(file)
	out := make([]string, len(args))
	for idx, arg := range args {
		out[idx] = strings.ReplaceAll(arg, FileToken, file)
	}
	return out
}

func splitSelectedFile(file string) (dir, name string) {
	idx := strings.LastIndexAny(file, `\/`)
	if idx < 0 {
		return "", file
	}
	return file[:idx], file[idx+1:]
}

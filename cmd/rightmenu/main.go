package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"rightmenu/internal/config"
	"rightmenu/internal/menu"
	"rightmenu/internal/runner"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printUsage(stdout)
		return nil
	}
	paths, err := config.DefaultPaths()
	if err != nil {
		return err
	}
	switch args[0] {
	case "install", "refresh":
		return install(paths, stdout)
	case "uninstall":
		return menu.NewRegistryInstaller().Uninstall()
	case "run":
		if len(args) != 3 {
			return fmt.Errorf("usage: rightmenu run <item-id> <selected-file-path>")
		}
		cfg, err := config.Load(paths.ConfigPath)
		if err != nil {
			return err
		}
		return runner.Run(cfg, args[1], args[2], nil)
	case "config":
		if err := config.Ensure(paths.ConfigPath); err != nil {
			return err
		}
		return openPath(paths.ConfigPath)
	case "paths":
		fmt.Fprintf(stdout, "config: %s\ninstallDir: %s\npinnedExe: %s\n", paths.ConfigPath, paths.InstallDir, paths.PinnedExe)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `rightmenu - DEV调试 Windows file context menu utility

Usage:
  rightmenu install      Pin the exe, create config if missing, and register DEV调试 menu
  rightmenu refresh      Rebuild registered menu items from config.json
  rightmenu uninstall    Remove the owned DEV调试 menu registry subtree
  rightmenu run <id> <file>
  rightmenu config       Open the config file
  rightmenu paths        Print canonical paths`)
}

func install(paths config.Paths, stdout io.Writer) error {
	if err := os.MkdirAll(paths.InstallDir, 0o755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}
	if err := pinExecutable(paths.PinnedExe); err != nil {
		return err
	}
	if err := config.Ensure(paths.ConfigPath); err != nil {
		return err
	}
	cfg, err := config.Load(paths.ConfigPath)
	if err != nil {
		return err
	}
	if err := menu.NewRegistryInstaller().Install(cfg, paths.PinnedExe); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "installed DEV调试 menu\nconfig: %s\npinned exe: %s\n", paths.ConfigPath, paths.PinnedExe)
	return nil
}

func pinExecutable(dest string) error {
	src, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}
	src, _ = filepath.Abs(src)
	dest, _ = filepath.Abs(dest)
	if samePath(src, dest) {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open current executable: %w", err)
	}
	defer in.Close()
	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create pinned executable: %w", err)
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("copy pinned executable: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close pinned executable: %w", closeErr)
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace pinned executable: %w", err)
	}
	return nil
}

func samePath(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	ai, aerr := os.Stat(a)
	bi, berr := os.Stat(b)
	return aerr == nil && berr == nil && os.SameFile(ai, bi)
}

func openPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open config %q: %w", path, err)
	}
	return nil
}

package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

type RunLogger interface {
	LogRun(entry LogEntry) error
}

type LogEntry struct {
	Time             string   `json:"time"`
	SelectedFile     string   `json:"selectedFile"`
	SelectedFileName string   `json:"selectedFileName"`
	Program          string   `json:"program"`
	Args             []string `json:"args"`
	Command          string   `json:"command"`
}

type FileLogger struct {
	Path string
	Now  func() time.Time
}

func (l FileLogger) LogRun(entry LogEntry) error {
	if strings.TrimSpace(l.Path) == "" {
		return nil
	}
	if entry.Time == "" {
		now := time.Now
		if l.Now != nil {
			now = l.Now
		}
		entry.Time = now().Format(time.RFC3339)
	}
	if entry.Command == "" {
		entry.Command = FormatCommand(entry.Program, entry.Args)
	}
	if err := os.MkdirAll(filepath.Dir(l.Path), 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}
	b = append(b, '\n')
	f, err := os.OpenFile(l.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log %q: %w", l.Path, err)
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("write log %q: %w", l.Path, err)
	}
	return nil
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
	return RunWithLogger(cfg, itemID, file, starter, nil)
}

func RunWithLogger(cfg config.Config, itemID, file string, starter Starter, logger RunLogger) error {
	program, args, err := BuildInvocation(cfg, itemID, file)
	if err != nil {
		return err
	}
	if logger != nil {
		if err := logger.LogRun(LogEntry{
			SelectedFile:     file,
			SelectedFileName: selectedFileName(file),
			Program:          program,
			Args:             append([]string(nil), args...),
			Command:          FormatCommand(program, args),
		}); err != nil {
			return err
		}
	}
	if starter == nil {
		starter = ExecStarter{}
	}
	if err := starter.Start(program, args); err != nil {
		return fmt.Errorf("start %q: %w", program, err)
	}
	return nil
}

func FormatCommand(program string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quoteCommandProgram(program))
	for _, arg := range args {
		parts = append(parts, quoteCommandPart(arg))
	}
	return strings.Join(parts, " ")
}

func quoteCommandProgram(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func quoteCommandPart(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, " \t\r\n\"") {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func selectedFileName(file string) string {
	idx := strings.LastIndexAny(file, `\/`)
	if idx < 0 {
		return file
	}
	return file[idx+1:]
}

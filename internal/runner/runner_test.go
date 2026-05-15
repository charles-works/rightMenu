package runner

import (
	"reflect"
	"strings"
	"testing"

	"rightmenu/internal/config"
)

type recordingStarter struct {
	program string
	args    []string
}

func (r *recordingStarter) Start(program string, args []string) error {
	r.program = program
	r.args = append([]string(nil), args...)
	return nil
}

func TestBuildInvocation(t *testing.T) {
	cfg := config.Config{MenuTitle: "DEV调试", Items: []config.Item{{ID: "aa", Title: "AA", Program: `C:\Tools\AA.exe`, SpecifiedFolder: `C:\Target Dir`}}}
	file := `C:\Temp\path with spaces\file.txt`
	program, args, err := BuildInvocation(cfg, "aa", file)
	if err != nil {
		t.Fatal(err)
	}
	if program != `C:\Tools\AA.exe` {
		t.Fatalf("program = %q", program)
	}
	if want := []string{" ", "IF_A_000N", `C:\Target Dir`, `C:\Temp\path with spaces`, "file.txt", "Q"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v want %#v", args, want)
	}
}

func TestBuildInvocationPreservesCustomArgsCompatibility(t *testing.T) {
	cfg := config.Config{MenuTitle: "DEV调试", Items: []config.Item{{ID: "aa", Title: "AA", Program: `C:\Tools\AA.exe`, Args: []string{"--open", config.FileToken}}}}
	file := `C:\Temp\path with spaces\file.txt`
	_, args, err := BuildInvocation(cfg, "aa", file)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"--open", file}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v want %#v", args, want)
	}
}

func TestRunUsesStarterWithoutShellConcatenation(t *testing.T) {
	cfg := config.Config{MenuTitle: "DEV调试", Items: []config.Item{{ID: "aa", Title: "AA", Program: "tool", Args: []string{config.FileToken}}}}
	starter := &recordingStarter{}
	file := `C:\Temp\path with spaces\file.txt`
	if err := Run(cfg, "aa", file, starter); err != nil {
		t.Fatal(err)
	}
	if starter.program != "tool" || !reflect.DeepEqual(starter.args, []string{file}) {
		t.Fatalf("recorded %#v %#v", starter.program, starter.args)
	}
}

func TestBuildInvocationErrors(t *testing.T) {
	cfg := config.Config{MenuTitle: "DEV调试", Items: []config.Item{{ID: "aa", Title: "AA", Program: "tool"}}}
	if _, _, err := BuildInvocation(cfg, "missing", "file"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("missing item err = %v", err)
	}
	if _, _, err := BuildInvocation(cfg, "aa", ""); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("missing file err = %v", err)
	}
}

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestValidateRejectsDuplicateAndInvalidIDs(t *testing.T) {
	cfg := Config{MenuTitle: "DEV调试", Items: []Item{
		{ID: "aa", Title: "AA", Program: "tool.exe", SpecifiedFolder: `C:\DEV`},
		{ID: "aa", Title: "BB", Program: "tool.exe", SpecifiedFolder: `C:\DEV`},
	}}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	cfg.Items[1].ID = "bad id"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "must match") {
		t.Fatalf("expected regex error, got %v", err)
	}
}

func TestEnsureLoadAndDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "RightMenu", "config.json")
	if err := Ensure(path); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MenuTitle != "DEV调试" || len(cfg.Items) != 1 || cfg.Items[0].ID != "aa" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if !cfg.LoggingEnabled() {
		t.Fatalf("default logging should be enabled")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestLoggingConfigDefaultsAndOverrides(t *testing.T) {
	cfg := Config{}
	if !cfg.LoggingEnabled() {
		t.Fatal("zero-value logging should be enabled")
	}
	if got := cfg.LogPath("default.log"); got != "default.log" {
		t.Fatalf("default log path = %q", got)
	}
	disabled := false
	cfg.Logging = Logging{Enabled: &disabled, Path: `C:\Logs\rightmenu.log`}
	if cfg.LoggingEnabled() {
		t.Fatal("logging should be disabled")
	}
	if got := cfg.LogPath("default.log"); got != `C:\Logs\rightmenu.log` {
		t.Fatalf("configured log path = %q", got)
	}
}

func TestExpandedArgsDefaultSixArgumentContract(t *testing.T) {
	item := Item{ID: "aa", Title: "AA", Program: "tool.exe", SpecifiedFolder: `C:\Target Dir`}
	file := `C:\Temp\路径 with spaces\sample file.txt`
	got := item.ExpandedArgs(file)
	want := []string{" ", "IF_A_000N", `C:\Target Dir`, `C:\Temp\路径 with spaces`, "sample file.txt", "Q"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	if got := item.ArgsWithDefaults(file); !reflect.DeepEqual(got, want) {
		t.Fatalf("default args = %#v want %#v", got, want)
	}
}

func TestExpandedArgsPreservesCustomArgsCompatibility(t *testing.T) {
	item := Item{ID: "aa", Title: "AA", Program: "tool.exe", Args: []string{"--file", FileToken, "prefix=" + FileToken}}
	file := `C:\Temp\路径 with spaces\sample file.txt`
	got := item.ExpandedArgs(file)
	want := []string{"--file", file, "prefix=" + file}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	if got := item.ArgsWithDefaults(file); !reflect.DeepEqual(got, []string{"--file", FileToken, "prefix=" + FileToken}) {
		t.Fatalf("custom args = %#v", got)
	}
}

func TestValidateRejectsRequiredFields(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "missing menu title",
			cfg:  Config{Items: []Item{{ID: "aa", Title: "AA", Program: "tool.exe"}}},
			want: "menuTitle",
		},
		{
			name: "missing items",
			cfg:  Config{MenuTitle: "DEV调试"},
			want: "at least one item",
		},
		{
			name: "missing item id",
			cfg:  Config{MenuTitle: "DEV调试", Items: []Item{{Title: "AA", Program: "tool.exe"}}},
			want: "id is required",
		},
		{
			name: "missing item title",
			cfg:  Config{MenuTitle: "DEV调试", Items: []Item{{ID: "aa", Program: "tool.exe"}}},
			want: "title is required",
		},
		{
			name: "missing item program",
			cfg:  Config{MenuTitle: "DEV调试", Items: []Item{{ID: "aa", Title: "AA", SpecifiedFolder: `C:\DEV`}}},
			want: "program is required",
		},
		{
			name: "missing specified folder without args override",
			cfg:  Config{MenuTitle: "DEV调试", Items: []Item{{ID: "aa", Title: "AA", Program: "tool.exe"}}},
			want: "specifiedFolder",
		},
		{
			name: "empty args still requires specified folder",
			cfg:  Config{MenuTitle: "DEV调试", Items: []Item{{ID: "aa", Title: "AA", Program: "tool.exe", Args: []string{}}}},
			want: "specifiedFolder",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

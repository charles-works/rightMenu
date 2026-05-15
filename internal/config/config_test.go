package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestValidateRejectsDuplicateAndInvalidIDs(t *testing.T) {
	cfg := Config{MenuTitle: "DEV调试", Items: []Item{{ID: "aa", Title: "AA", Program: "tool.exe"}, {ID: "aa", Title: "BB", Program: "tool.exe"}}}
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
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestExpandedArgsPreservesSpacesAndUnicode(t *testing.T) {
	item := Item{ID: "aa", Title: "AA", Program: "tool.exe", Args: []string{"--file", FileToken, "prefix=" + FileToken}}
	file := `C:\Temp\路径 with spaces\sample file.txt`
	got := item.ExpandedArgs(file)
	want := []string{"--file", file, "prefix=" + file}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	if got := (Item{}).ArgsWithDefaults(); !reflect.DeepEqual(got, []string{FileToken}) {
		t.Fatalf("default args = %#v", got)
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
			cfg:  Config{MenuTitle: "DEV调试", Items: []Item{{ID: "aa", Title: "AA"}}},
			want: "program is required",
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

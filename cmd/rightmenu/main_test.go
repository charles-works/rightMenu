package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "rightmenu install") {
		t.Fatalf("help output missing install: %s", out.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"bogus"}, &out); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown command error, got %v", err)
	}
}

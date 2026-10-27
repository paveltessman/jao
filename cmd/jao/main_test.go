package main

import (
	"strings"
	"testing"
)

func TestDoFailsWithoutSubcommand(t *testing.T) {
	err := do(t.Context(), nil)
	if err == nil {
		t.Fatal("do returned no error, want one")
	}
}

func TestDoFailsOnUnknownSubcommand(t *testing.T) {
	err := do(t.Context(), []string{"nope"})
	if err == nil {
		t.Fatal("do returned no error, want one")
	}
	if !strings.Contains(err.Error(), `"nope"`) {
		t.Errorf("error = %q, want it to name the subcommand", err)
	}
}

func TestKeyFailsOnExtraArgs(t *testing.T) {
	err := do(t.Context(), []string{"key", "a", "b"})
	if err == nil {
		t.Fatal("do returned no error, want one")
	}
	if !strings.Contains(err.Error(), "jao key [name]") {
		t.Errorf("error = %q, want the usage", err)
	}
}

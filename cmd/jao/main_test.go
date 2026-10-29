package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/paveltessman/jao/keys"
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
	if !strings.Contains(err.Error(), "jao key <protocol name>") {
		t.Errorf("error = %q, want the usage", err)
	}
}

func TestListFailsOnArgs(t *testing.T) {
	err := do(t.Context(), []string{"list", "a"})
	if err == nil {
		t.Fatal("do returned no error, want one")
	}
	if !strings.Contains(err.Error(), "jao list") {
		t.Errorf("error = %q, want the usage", err)
	}
}

func TestPrintInstalled(t *testing.T) {
	var out bytes.Buffer
	err := printInstalled(&out, []keys.Installed{
		{Name: "jao-awg", Protocol: "awg"},
		{Name: "jao-awg-second", Protocol: "awg"},
	})
	if err != nil {
		t.Fatalf("print: %v", err)
	}

	want := "NAME            PROTOCOL\n" +
		"jao-awg         awg\n" +
		"jao-awg-second  awg\n"
	if out.String() != want {
		t.Errorf("output:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestPrintInstalledWithoutContainers(t *testing.T) {
	var out bytes.Buffer
	if err := printInstalled(&out, nil); err != nil {
		t.Fatalf("print: %v", err)
	}
	if out.String() != "NAME  PROTOCOL\n" {
		t.Errorf("output = %q, want the header only", out.String())
	}
}

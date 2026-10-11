package settings

import (
	"testing"
)

func load(t *testing.T, env map[string]string) Settings {

	for key, value := range env {
		t.Setenv(key, value)
	}

	settings := New()
	return settings
}

func TestNew(t *testing.T) {

	workDir := "/opt/hoba"

	settings := load(t, map[string]string{envWorkDir: workDir})

	if settings.WorkDir != workDir {
		t.Logf("workdir: want %s, got %s", workDir, settings.WorkDir)
		t.Fail()
	}
}

func TestNewSetsDefaults(t *testing.T) {

	settings := load(t, map[string]string{})

	if settings.WorkDir != defaultWorkDir {
		t.Logf("want default walue for workdir, got %s", settings.WorkDir)
		t.Fail()
	}
}

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
	publicHost := "203.0.113.7"

	settings := load(t, map[string]string{envWorkDir: workDir, envPublicHost: publicHost})

	if settings.WorkDir != workDir {
		t.Logf("workdir: want %s, got %s", workDir, settings.WorkDir)
		t.Fail()
	}
	if settings.PublicHost != publicHost {
		t.Logf("public host: want %s, got %s", publicHost, settings.PublicHost)
		t.Fail()
	}
}

func TestNewSetsDefaults(t *testing.T) {

	// Empty values, so that the env of the shell does not change the result.
	settings := load(t, map[string]string{envWorkDir: "", envPublicHost: ""})

	if settings.WorkDir != defaultWorkDir {
		t.Logf("want default walue for workdir, got %s", settings.WorkDir)
		t.Fail()
	}
	if settings.PublicHost != "" {
		t.Logf("want no default for public host, got %s", settings.PublicHost)
		t.Fail()
	}
}

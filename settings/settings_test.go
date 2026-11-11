package settings

import (
	"errors"
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
	httpAddr := "123.456.789.012:1234"

	settings := load(t, map[string]string{
		envWorkDir:    workDir,
		envPublicHost: publicHost,
		envHttpAddr:   httpAddr,
		envSecretKey:  "opa",
	})

	if settings.WorkDir != workDir {
		t.Logf("workdir: want %s, got %s", workDir, settings.WorkDir)
		t.Fail()
	}
	if settings.PublicHost != publicHost {
		t.Logf("public host: want %s, got %s", publicHost, settings.PublicHost)
		t.Fail()
	}

	if settings.HttpListenAddr != httpAddr {
		t.Errorf("http listen address: want %s, got %s", httpAddr, settings.HttpListenAddr)
	}
}

func TestNewSetsDefaults(t *testing.T) {

	settings := load(t, map[string]string{envWorkDir: "", envPublicHost: "", envSecretKey: "opa"})

	if settings.WorkDir != defaultWorkDir {
		t.Logf("want default walue for workdir, got %s", settings.WorkDir)
		t.Fail()
	}
	if settings.PublicHost != "" {
		t.Logf("want no default for public host, got %s", settings.PublicHost)
		t.Fail()
	}

	if settings.HttpListenAddr != "127.0.0.1:8001" {
		t.Errorf("wrong default http listen address: got %s", settings.HttpListenAddr)
	}

}

func TestNewPanicsWithEmptySecret(t *testing.T) {

	defer func() {
		v := recover()

		if v == nil {
			t.Errorf("want panic, got nil")
		}

		err := v.(error)

		if !errors.Is(err, ErrSecretNotSet) {
			t.Errorf("want the secret key error, got %q", err)
		}

	}()

	_ = load(t, map[string]string{envSecretKey: ""})

}

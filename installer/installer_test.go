package installer

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/protocol/awg"
)

type fakeContainerAPI struct {
	buildErr error

	builtTag  string
	created   []container.ContainerSpec
	startedID string
}

func (f *fakeContainerAPI) ImageBuild(_ context.Context, _ fs.FS, imageTag string) error {
	f.builtTag = imageTag
	return f.buildErr
}

func (f *fakeContainerAPI) ContainerCreate(_ context.Context, spec container.ContainerSpec) (string, error) {
	f.created = append(f.created, spec)
	return "ctr-1", nil
}

func (f *fakeContainerAPI) ContainerStart(_ context.Context, id string) error {
	f.startedID = id
	return nil
}

func (f *fakeContainerAPI) ContainerExec(_ context.Context, _ string, _ []string) (string, error) {
	return "", nil
}

func TestInstallAwgCreatesContainerFromConfig(t *testing.T) {

	ctr := &fakeContainerAPI{}
	workDir := t.TempDir()
	cfg := awg.DefaultConfig()
	cfg.Name = "awg-test"
	cfg.Port = 40001

	if err := New(ctr, workDir).InstallAwg(t.Context(), cfg); err != nil {
		t.Fatalf("install: %v", err)
	}

	if len(ctr.created) != 1 {
		t.Fatalf("want 1 container, got %d", len(ctr.created))
	}
	spec := ctr.created[0]

	if spec.Image != ctr.builtTag {
		t.Errorf("image: want built tag %s, got %s", ctr.builtTag, spec.Image)
	}
	if spec.Name != "awg-test" {
		t.Errorf("name: want awg-test, got %s", spec.Name)
	}
	if ctr.startedID != "ctr-1" {
		t.Errorf("start: want ctr-1, got %q", ctr.startedID)
	}

	wantEnv := map[string]string{
		"AWG_SERVER_PORT":       "40001",
		"AWG_SUBNET_IP":         "10.8.1.0",
		"WIREGUARD_SUBNET_CIDR": "24",
		"HEADER_PROTECTION_KEY": cfg.HeaderProtectionKey,
		"REKEY_AFTER_TIME":      "100-120",
		"DISABLE_COOKIES":       "on",
	}
	for key, want := range wantEnv {
		if got := spec.Env[key]; got != want {
			t.Errorf("env %s: want %q, got %q", key, want, got)
		}
	}

	wantPort := container.PortBinding{HostPort: 40001, ContainerPort: 40001, Protocol: container.UDP}
	if !slices.Equal(spec.Ports, []container.PortBinding{wantPort}) {
		t.Errorf("ports: want [%v], got %v", wantPort, spec.Ports)
	}

	wantMount := container.Mount{Source: filepath.Join(workDir, "awg-test"), Target: "/opt/amnezia/awg"}
	if !slices.Equal(spec.Mounts, []container.Mount{wantMount}) {
		t.Errorf("mounts: want [%v], got %v", wantMount, spec.Mounts)
	}
	info, err := os.Stat(wantMount.Source)
	if err != nil {
		t.Fatalf("data dir: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("data dir mode: want 0700, got %o", perm)
	}

	if !slices.Contains(spec.CapAdd, "NET_ADMIN") {
		t.Errorf("caps: want NET_ADMIN, got %v", spec.CapAdd)
	}
	if !slices.Equal(spec.Devices, []string{"/dev/net/tun"}) {
		t.Errorf("devices: want [/dev/net/tun], got %v", spec.Devices)
	}
	if spec.Sysctls["net.ipv4.ip_forward"] != "1" {
		t.Errorf("sysctls: want ip_forward=1, got %v", spec.Sysctls)
	}
	if spec.Restart != container.RestartUnlessStopped {
		t.Errorf("restart: want unless-stopped, got %q", spec.Restart)
	}
}

func TestInstallAwgStopsOnBuildError(t *testing.T) {

	buildErr := errors.New("no network")
	ctr := &fakeContainerAPI{buildErr: buildErr}

	err := New(ctr, t.TempDir()).InstallAwg(t.Context(), awg.DefaultConfig())
	if !errors.Is(err, buildErr) {
		t.Errorf("want build error, got %v", err)
	}
	if len(ctr.created) != 0 {
		t.Errorf("want no container after a failed build, got %d", len(ctr.created))
	}
}

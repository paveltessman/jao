//go:build integration

package container

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func newService(t *testing.T) *Service {
	t.Helper()

	svc, err := New()
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Errorf("close service: %v", err)
		}
	})

	return svc
}

func TestPing(t *testing.T) {

	svc := newService(t)

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	ping, err := svc.Ping(ctx)
	if err != nil {
		t.Fatalf("ping: %v", err)
	}

	if ping.APIVersion == "" {
		t.Errorf("want API version, got empty string")
	}
	if ping.OSType == "" {
		t.Errorf("want OS type, got empty string")
	}
}

const testImage = "jao/container-test:latest"

// capNetAdmin is the bit of CAP_NET_ADMIN in the capability sets.
const capNetAdmin = 12

func startTestContainer(t *testing.T, svc *Service, spec ContainerSpec) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	workdir := fstest.MapFS{
		"Dockerfile": {Data: []byte("FROM alpine:3.22\nCMD [\"sleep\", \"infinity\"]\n")},
	}
	if err := svc.ImageBuild(ctx, workdir, testImage); err != nil {
		t.Fatalf("build image: %v", err)
	}
	t.Cleanup(func() {
		_, err := svc.client.ImageRemove(context.Background(), testImage, client.ImageRemoveOptions{Force: true})
		if err != nil {
			t.Errorf("remove image: %v", err)
		}
	})

	spec.Image = testImage
	id, err := svc.ContainerCreate(ctx, spec)
	if err != nil {
		t.Fatalf("create container: %v", err)
	}
	t.Cleanup(func() {
		if err := svc.ContainerRemove(context.Background(), id); err != nil {
			t.Errorf("remove container: %v", err)
		}
	})

	if err := svc.ContainerStart(ctx, id); err != nil {
		t.Fatalf("start container: %v", err)
	}

	return id
}

func freeUDPPort(t *testing.T) uint16 {
	t.Helper()

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	defer func() { _ = conn.Close() }()

	return uint16(conn.LocalAddr().(*net.UDPAddr).Port)
}

func exec(t *testing.T, svc *Service, id string, cmd ...string) string {
	t.Helper()

	out, err := svc.ContainerExec(t.Context(), id, cmd)
	if err != nil {
		t.Fatalf("exec %v: %v", cmd, err)
	}
	return strings.TrimSpace(out)
}

func TestContainerCreateAppliesSpec(t *testing.T) {

	svc := newService(t)

	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "hello"), []byte("from host"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	hostPort := freeUDPPort(t)

	id := startTestContainer(t, svc, ContainerSpec{
		Name:    "jao-container-test",
		Env:     map[string]string{"FOO": "bar"},
		CapAdd:  []string{"NET_ADMIN"},
		Devices: []string{"/dev/net/tun"},
		Ports:   []PortBinding{{HostPort: hostPort, ContainerPort: 4000, Protocol: UDP}},
		Mounts:  []Mount{{Source: dataDir, Target: "/data"}},
		Sysctls: map[string]string{"net.ipv4.ip_forward": "1"},
		Restart: RestartUnlessStopped,
	})

	if got := exec(t, svc, id, "printenv", "FOO"); got != "bar" {
		t.Errorf("env FOO: want bar, got %q", got)
	}
	if got := exec(t, svc, id, "cat", "/data/hello"); got != "from host" {
		t.Errorf("mount: want %q, got %q", "from host", got)
	}
	if got := exec(t, svc, id, "cat", "/proc/sys/net/ipv4/ip_forward"); got != "1" {
		t.Errorf("sysctl ip_forward: want 1, got %q", got)
	}
	exec(t, svc, id, "test", "-c", "/dev/net/tun")

	capLine := exec(t, svc, id, "grep", "CapEff", "/proc/1/status")
	capEff, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(capLine, "CapEff:")), 16, 64)
	if err != nil {
		t.Fatalf("parse %q: %v", capLine, err)
	}
	if capEff&(1<<capNetAdmin) == 0 {
		t.Errorf("want NET_ADMIN in effective caps, got %s", capLine)
	}

	inspect, err := svc.client.ContainerInspect(t.Context(), id, client.ContainerInspectOptions{})
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	hostCfg := inspect.Container.HostConfig
	if hostCfg.RestartPolicy.Name != container.RestartPolicyUnlessStopped {
		t.Errorf("restart policy: want unless-stopped, got %q", hostCfg.RestartPolicy.Name)
	}
	if inspect.Container.Name != "/jao-container-test" {
		t.Errorf("name: want /jao-container-test, got %q", inspect.Container.Name)
	}

	var bound []string
	for port, bindings := range hostCfg.PortBindings {
		for _, b := range bindings {
			bound = append(bound, b.HostPort+"->"+port.String())
		}
	}
	want := strconv.Itoa(int(hostPort)) + "->4000/udp"
	if len(bound) != 1 || bound[0] != want {
		t.Errorf("port bindings: want [%s], got %v", want, bound)
	}
}

func TestContainerExecFailsOnNonZeroExit(t *testing.T) {

	svc := newService(t)
	id := startTestContainer(t, svc, ContainerSpec{})

	_, err := svc.ContainerExec(t.Context(), id, []string{"sh", "-c", "echo oops >&2; exit 3"})
	if !errors.Is(err, ErrContainerExec) {
		t.Fatalf("want ErrContainerExec, got %v", err)
	}
	if !strings.Contains(err.Error(), "exit code 3") || !strings.Contains(err.Error(), "oops") {
		t.Errorf("want exit code and stderr in error, got %v", err)
	}
}

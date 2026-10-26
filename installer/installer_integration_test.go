//go:build integration

package installer_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/installer"
)

func freeUDPPort(t *testing.T) uint16 {
	t.Helper()

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	defer func() { _ = conn.Close() }()

	return uint16(conn.LocalAddr().(*net.UDPAddr).Port)
}

// waitExec runs cmd in the container until it succeeds or the timeout ends.
// start.sh brings up awg0 some time after the container starts.
func waitExec(t *testing.T, ctr *container.Service, id string, cmd ...string) string {
	t.Helper()

	deadline := time.Now().Add(30 * time.Second)
	for {
		out, err := ctr.ContainerExec(t.Context(), id, cmd)
		if err == nil {
			return out
		}
		if time.Now().After(deadline) {
			t.Fatalf("exec %v: %v", cmd, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func TestInstallAwgBringsUpInterface(t *testing.T) {

	ctr, err := container.New()
	if err != nil {
		t.Fatalf("create container service: %v", err)
	}
	t.Cleanup(func() {
		if err := ctr.Close(); err != nil {
			t.Errorf("close container service: %v", err)
		}
	})

	workDir := t.TempDir()
	cfg := installer.DefaultAwgConfig()
	cfg.Name = "jao-awg-test"
	cfg.Port = freeUDPPort(t)

	// Remove the container also when the install fails after create.
	t.Cleanup(func() {
		if err := ctr.ContainerRemove(context.Background(), cfg.Name); err != nil {
			t.Logf("remove container: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()

	if err := installer.New(ctr, workDir).InstallAwg(ctx, cfg); err != nil {
		t.Fatalf("install: %v", err)
	}

	show := waitExec(t, ctr, cfg.Name, "awg", "show", "awg0")

	for _, want := range []string{
		"listening port: " + strconv.Itoa(int(cfg.Port)),
		"jc: " + strconv.Itoa(cfg.JunkPacketCount),
		"s4: 12",
		"header protection key: " + cfg.HeaderProtectionKey,
		"rekey after time: 100-120",
		"disable cookies: on",
	} {
		if !strings.Contains(show, want) {
			t.Errorf("awg show: want %q, got:\n%s", want, show)
		}
	}

	pubKey, err := os.ReadFile(filepath.Join(workDir, cfg.Name, "wireguard_server_public_key.key"))
	if err != nil {
		t.Fatalf("read public key on host: %v", err)
	}
	ifaceKey := waitExec(t, ctr, cfg.Name, "awg", "show", "awg0", "public-key")
	if strings.TrimSpace(string(pubKey)) != strings.TrimSpace(ifaceKey) {
		t.Errorf("public key: want host file %q to match awg0 %q", pubKey, ifaceKey)
	}
}

func newContainerService(t *testing.T) *container.Service {
	t.Helper()

	ctr, err := container.New()
	if err != nil {
		t.Fatalf("create container service: %v", err)
	}
	t.Cleanup(func() {
		if err := ctr.Close(); err != nil {
			t.Errorf("close container service: %v", err)
		}
	})
	return ctr
}

// installAwg installs cfg and removes the container when the test ends.
func installAwg(t *testing.T, ctr *container.Service, inst *installer.Service, cfg installer.AwgConfig) {
	t.Helper()

	t.Cleanup(func() {
		if err := ctr.ContainerRemove(context.Background(), cfg.Name); err != nil {
			t.Logf("remove container: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()

	if err := inst.InstallAwg(ctx, cfg); err != nil {
		t.Fatalf("install %s: %v", cfg.Name, err)
	}
	waitExec(t, ctr, cfg.Name, "awg", "show", "awg0")
}

func TestAddAwgPeerConnectsClient(t *testing.T) {

	ctr := newContainerService(t)
	inst := installer.New(ctr, t.TempDir())

	server := installer.DefaultAwgConfig()
	server.Name = "jao-awg-test-server"
	server.Port = freeUDPPort(t)
	installAwg(t, ctr, inst, server)

	// The server container is the endpoint on the docker bridge network.
	serverIP := strings.TrimSpace(waitExec(t, ctr, server.Name, "hostname", "-i"))

	first, err := inst.AddAwgPeer(t.Context(), server, serverIP)
	if err != nil {
		t.Fatalf("add first peer: %v", err)
	}
	second, err := inst.AddAwgPeer(t.Context(), server, serverIP)
	if err != nil {
		t.Fatalf("add second peer: %v", err)
	}

	show := waitExec(t, ctr, server.Name, "awg", "show", "awg0", "allowed-ips")
	for _, ip := range []string{"10.8.1.1/32", "10.8.1.2/32"} {
		if !strings.Contains(show, ip) {
			t.Errorf("awg show: want peer %s, got:\n%s", ip, show)
		}
	}
	if !strings.Contains(second, "Address = 10.8.1.2/32") {
		t.Errorf("second peer: want 10.8.1.2, got:\n%s", second)
	}

	// The client runs in a second AWG container with its own awg0 on another subnet.
	// The test drops DNS, because awg-quick needs resolvconf for it,
	// and routes only the server subnet, to keep the container network.
	client := installer.DefaultAwgConfig()
	client.Name = "jao-awg-test-client"
	client.Port = freeUDPPort(t)
	client.SubnetIP = "10.9.9.0"
	installAwg(t, ctr, inst, client)

	clientConf := regexp.MustCompile(`(?m)^DNS = .*\n`).ReplaceAllString(first, "")
	clientConf = strings.Replace(clientConf, "AllowedIPs = 0.0.0.0/0, ::/0", "AllowedIPs = 10.8.1.0/32", 1)

	waitExec(t, ctr, client.Name, "bash", "-c", `printf '%s' "$1" > /tmp/jao0.conf && awg-quick up /tmp/jao0.conf`, "_", clientConf)
	waitExec(t, ctr, client.Name, "ping", "-c", "1", "-W", "2", "10.8.1.0")
}

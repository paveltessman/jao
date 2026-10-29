//go:build integration

package installer_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/installer"
	"github.com/paveltessman/jao/protocol"
	"github.com/paveltessman/jao/protocol/awg"
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
	cfg := awg.DefaultConfig()
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

	// The key service finds the container by its label and reads cfg back.
	list, err := ctr.ContainerList(t.Context(), protocol.Label)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	i := slices.IndexFunc(list, func(c container.Container) bool { return c.Name == cfg.Name })
	if i < 0 {
		t.Fatalf("want %s in the labeled containers, got %v", cfg.Name, list)
	}
	if got := list[i].Labels[protocol.Label]; got != protocol.AWGLabel {
		t.Errorf("label %s: want %s, got %q", protocol.Label, protocol.AWGLabel, got)
	}
	read, err := awg.ConfigFromEnv(list[i].Name, list[i].Env)
	if err != nil {
		t.Fatalf("config from env: %v", err)
	}
	if read != cfg {
		t.Errorf("config from env: want\n%+v\ngot\n%+v", cfg, read)
	}
}

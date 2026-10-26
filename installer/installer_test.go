package installer

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type fakeContainerAPI struct {
	buildErr error
	execErr  error
	files    map[string]string // `cat <path>` returns files[path]

	builtTag  string
	created   []ContainerSpec
	startedID string
	execs     [][]string // all commands other than `cat`
}

func (f *fakeContainerAPI) ImageBuild(_ context.Context, _ fs.FS, imageTag string) error {
	f.builtTag = imageTag
	return f.buildErr
}

func (f *fakeContainerAPI) ContainerCreate(_ context.Context, spec ContainerSpec) (string, error) {
	f.created = append(f.created, spec)
	return "ctr-1", nil
}

func (f *fakeContainerAPI) ContainerStart(_ context.Context, id string) error {
	f.startedID = id
	return nil
}

func (f *fakeContainerAPI) ContainerExec(_ context.Context, _ string, cmd []string) (string, error) {
	if f.execErr != nil {
		return "", f.execErr
	}
	if cmd[0] == "cat" {
		content, ok := f.files[cmd[1]]
		if !ok {
			return "", fmt.Errorf("cat: %s: no such file", cmd[1])
		}
		return content, nil
	}
	f.execs = append(f.execs, cmd)
	return "", nil
}

func TestInstallAwgCreatesContainerFromConfig(t *testing.T) {

	ctr := &fakeContainerAPI{}
	workDir := t.TempDir()
	cfg := DefaultAwgConfig()
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

	wantPort := PortBinding{HostPort: 40001, ContainerPort: 40001, Protocol: UDP}
	if !slices.Equal(spec.Ports, []PortBinding{wantPort}) {
		t.Errorf("ports: want [%v], got %v", wantPort, spec.Ports)
	}

	wantMount := Mount{Source: filepath.Join(workDir, "awg-test"), Target: "/opt/amnezia/awg"}
	if !slices.Equal(spec.Mounts, []Mount{wantMount}) {
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
	if spec.Restart != RestartUnlessStopped {
		t.Errorf("restart: want unless-stopped, got %q", spec.Restart)
	}
}

func TestInstallAwgStopsOnBuildError(t *testing.T) {

	buildErr := errors.New("no network")
	ctr := &fakeContainerAPI{buildErr: buildErr}

	err := New(ctr, t.TempDir()).InstallAwg(t.Context(), DefaultAwgConfig())
	if !errors.Is(err, buildErr) {
		t.Errorf("want build error, got %v", err)
	}
	if len(ctr.created) != 0 {
		t.Errorf("want no container after a failed build, got %d", len(ctr.created))
	}
}

func awgServerFiles() map[string]string {
	return map[string]string{
		awgConfPath:            "[Interface]\nAddress = 10.8.1.0/24\n\n[Peer]\nPublicKey = YQ==\nAllowedIPs = 10.8.1.1/32\n",
		awgServerPublicKeyPath: "c2VydmVyLXB1Yg==\n",
		awgPskPath:             "cHNr\n",
	}
}

// confValue returns the value of the first `key = value` line in conf.
func confValue(t *testing.T, conf, key string) string {
	t.Helper()

	m := regexp.MustCompile(`(?m)^` + key + ` = (.*)$`).FindStringSubmatch(conf)
	if m == nil {
		t.Fatalf("want %s in:\n%s", key, conf)
	}
	return m[1]
}

func TestAddAwgPeerAddsPeerForClientConfig(t *testing.T) {

	ctr := &fakeContainerAPI{files: awgServerFiles()}
	cfg := DefaultAwgConfig()

	clientConf, err := New(ctr, t.TempDir()).AddAwgPeer(t.Context(), cfg, "203.0.113.7")
	if err != nil {
		t.Fatalf("add peer: %v", err)
	}

	if len(ctr.execs) != 1 {
		t.Fatalf("want 1 write exec, got %v", ctr.execs)
	}
	cmd := ctr.execs[0]
	if len(cmd) != 5 || cmd[0] != "bash" || cmd[2] != addAwgPeerScript {
		t.Fatalf("want the peer script with the peer as $1, got %q", cmd)
	}
	peer := cmd[4]

	// The server peer must have the public key of the client private key.
	raw, err := base64.StdEncoding.DecodeString(confValue(t, clientConf, "PrivateKey"))
	if err != nil {
		t.Fatalf("decode client private key: %v", err)
	}
	key, err := ecdh.X25519().NewPrivateKey(raw)
	if err != nil {
		t.Fatalf("parse client private key: %v", err)
	}
	wantPub := base64.StdEncoding.EncodeToString(key.PublicKey().Bytes())
	if got := confValue(t, peer, "PublicKey"); got != wantPub {
		t.Errorf("peer public key: want %s, got %s", wantPub, got)
	}

	if got := confValue(t, peer, "AllowedIPs"); got != "10.8.1.2/32" {
		t.Errorf("peer allowed IPs: want 10.8.1.2/32, got %s", got)
	}
	if got := confValue(t, peer, "PresharedKey"); got != "cHNr" {
		t.Errorf("peer PSK: want cHNr, got %q", got)
	}
	if !strings.HasPrefix(peer, "\n[Peer]\n") {
		t.Errorf("want the peer to start on a new line, got %q", peer)
	}

	if strings.Contains(peer, confValue(t, clientConf, "PrivateKey")) {
		t.Error("want no client private key on the server")
	}

	for key, want := range map[string]string{
		"Address":      "10.8.1.2/32",
		"PublicKey":    "c2VydmVyLXB1Yg==",
		"PresharedKey": "cHNr",
		"Endpoint":     "203.0.113.7:55424",
	} {
		if got := confValue(t, clientConf, key); got != want {
			t.Errorf("client %s: want %q, got %q", key, want, got)
		}
	}
}

func TestAddAwgPeerStopsOnReadError(t *testing.T) {

	execErr := errors.New("no such container")
	ctr := &fakeContainerAPI{execErr: execErr}

	_, err := New(ctr, t.TempDir()).AddAwgPeer(t.Context(), DefaultAwgConfig(), "203.0.113.7")
	if !errors.Is(err, execErr) {
		t.Errorf("want exec error, got %v", err)
	}
	if len(ctr.execs) != 0 {
		t.Errorf("want no write after a failed read, got %v", ctr.execs)
	}
}

func TestAddAwgPeerStopsOnFullSubnet(t *testing.T) {

	files := awgServerFiles()
	files[awgConfPath] = "AllowedIPs = 10.8.1.1/32\nAllowedIPs = 10.8.1.2/32\n"
	ctr := &fakeContainerAPI{files: files}
	cfg := DefaultAwgConfig()
	cfg.SubnetCIDR = 30

	_, err := New(ctr, t.TempDir()).AddAwgPeer(t.Context(), cfg, "203.0.113.7")
	if !errors.Is(err, ErrSubnetFull) {
		t.Errorf("want ErrSubnetFull, got %v", err)
	}
	if len(ctr.execs) != 0 {
		t.Errorf("want no write on a full subnet, got %v", ctr.execs)
	}
}

func TestAddAwgPeerNeedsEndpoint(t *testing.T) {

	ctr := &fakeContainerAPI{files: awgServerFiles()}

	_, err := New(ctr, t.TempDir()).AddAwgPeer(t.Context(), DefaultAwgConfig(), "")
	if !errors.Is(err, ErrNoEndpoint) {
		t.Errorf("want ErrNoEndpoint, got %v", err)
	}
	if len(ctr.execs) != 0 {
		t.Errorf("want no write without an endpoint, got %v", ctr.execs)
	}
}

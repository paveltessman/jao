package awg

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

type fakeExecer struct {
	execErr error
	files   map[string]string // `cat <path>` returns files[path]

	execs [][]string // all commands other than `cat`
}

func (f *fakeExecer) ContainerExec(_ context.Context, _ string, cmd []string) (string, error) {
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

func serverFiles() map[string]string {
	return map[string]string{
		confPath:            "[Interface]\nAddress = 10.8.1.0/24\n\n[Peer]\nPublicKey = YQ==\nAllowedIPs = 10.8.1.1/32\n",
		serverPublicKeyPath: "c2VydmVyLXB1Yg==\n",
		pskPath:             "cHNr\n",
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

func TestAddPeerAddsPeerForClientConfig(t *testing.T) {

	ctr := &fakeExecer{files: serverFiles()}
	cfg := DefaultConfig()

	clientConf, err := AddPeer(t.Context(), ctr, cfg, "203.0.113.7")
	if err != nil {
		t.Fatalf("add peer: %v", err)
	}

	if len(ctr.execs) != 1 {
		t.Fatalf("want 1 write exec, got %v", ctr.execs)
	}
	cmd := ctr.execs[0]
	if len(cmd) != 5 || cmd[0] != "bash" || cmd[2] != addPeerScript {
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

func TestAddPeerStopsOnReadError(t *testing.T) {

	execErr := errors.New("no such container")
	ctr := &fakeExecer{execErr: execErr}

	_, err := AddPeer(t.Context(), ctr, DefaultConfig(), "203.0.113.7")
	if !errors.Is(err, execErr) {
		t.Errorf("want exec error, got %v", err)
	}
	if len(ctr.execs) != 0 {
		t.Errorf("want no write after a failed read, got %v", ctr.execs)
	}
}

func TestAddPeerStopsOnFullSubnet(t *testing.T) {

	files := serverFiles()
	files[confPath] = "AllowedIPs = 10.8.1.1/32\nAllowedIPs = 10.8.1.2/32\n"
	ctr := &fakeExecer{files: files}
	cfg := DefaultConfig()
	cfg.SubnetCIDR = 30

	_, err := AddPeer(t.Context(), ctr, cfg, "203.0.113.7")
	if !errors.Is(err, ErrSubnetFull) {
		t.Errorf("want ErrSubnetFull, got %v", err)
	}
	if len(ctr.execs) != 0 {
		t.Errorf("want no write on a full subnet, got %v", ctr.execs)
	}
}

func TestAddPeerNeedsEndpoint(t *testing.T) {

	ctr := &fakeExecer{files: serverFiles()}

	_, err := AddPeer(t.Context(), ctr, DefaultConfig(), "")
	if !errors.Is(err, ErrNoEndpoint) {
		t.Errorf("want ErrNoEndpoint, got %v", err)
	}
	if len(ctr.execs) != 0 {
		t.Errorf("want no write without an endpoint, got %v", ctr.execs)
	}
}

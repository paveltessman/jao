package keys

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/protocol"
	"github.com/paveltessman/jao/protocol/awg"
)

type fakeContainerAPI struct {
	listErr    error
	containers []container.Container
	files      map[string]string // `cat <path>` returns files[path] for each container

	listLabel string
	execIDs   []string // the containers of all commands other than `cat`
}

func (f *fakeContainerAPI) ContainerList(_ context.Context, label string) ([]container.Container, error) {
	f.listLabel = label
	return f.containers, f.listErr
}

func (f *fakeContainerAPI) ContainerExec(_ context.Context, id string, cmd []string) (string, error) {
	if cmd[0] == "cat" {
		content, ok := f.files[cmd[1]]
		if !ok {
			return "", fmt.Errorf("cat: %s: no such file", cmd[1])
		}
		return content, nil
	}
	f.execIDs = append(f.execIDs, id)
	return "", nil
}

// serverFiles holds the files that the AWG peer logic reads in the container.
func serverFiles() map[string]string {
	return map[string]string{
		"/opt/amnezia/awg/awg0.conf":                       "[Interface]\nAddress = 10.8.1.0/24\n",
		"/opt/amnezia/awg/wireguard_server_public_key.key": "c2VydmVyLXB1Yg==\n",
		"/opt/amnezia/awg/wireguard_psk.key":               "cHNr\n",
	}
}

// awgContainer returns a labeled AWG container with the env of cfg.
func awgContainer(t *testing.T, cfg awg.Config) container.Container {
	t.Helper()

	spec := cfg.Spec(t.TempDir())
	env := maps.Clone(spec.Env)
	env["PATH"] = "/usr/bin"
	return container.Container{ID: "id-" + cfg.Name, Name: cfg.Name, Labels: spec.Labels, Env: env}
}

func awgConfig(name string) awg.Config {
	cfg := awg.DefaultConfig()
	cfg.Name = name
	return cfg
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

func TestCreateKeyUsesServerParamsFromContainer(t *testing.T) {

	// The server params differ from any new DefaultConfig,
	// so a client config from defaults fails this test.
	server := awgConfig("jao-awg")
	server.Port = 40001
	server.JunkPacketCount = 9
	server.HeaderProtectionKey = "aHBr"
	ctr := &fakeContainerAPI{
		containers: []container.Container{awgContainer(t, server)},
		files:      serverFiles(),
	}

	clientConf, err := New(ctr, "203.0.113.7").CreateKey(t.Context(), "")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	if ctr.listLabel != protocol.Label {
		t.Errorf("list label: want %s, got %q", protocol.Label, ctr.listLabel)
	}
	if !slices.Equal(ctr.execIDs, []string{"jao-awg"}) {
		t.Errorf("want 1 write in jao-awg, got %v", ctr.execIDs)
	}
	for key, want := range map[string]string{
		"Endpoint":            "203.0.113.7:40001",
		"Jc":                  strconv.Itoa(server.JunkPacketCount),
		"HeaderProtectionKey": server.HeaderProtectionKey,
		"I1":                  server.SpecialJunk1,
	} {
		if got := confValue(t, clientConf, key); got != want {
			t.Errorf("client %s: want %q, got %q", key, want, got)
		}
	}
}

func TestCreateKeyPicksContainerByName(t *testing.T) {

	ctr := &fakeContainerAPI{
		containers: []container.Container{awgContainer(t, awgConfig("awg-a")), awgContainer(t, awgConfig("awg-b"))},
		files:      serverFiles(),
	}

	if _, err := New(ctr, "203.0.113.7").CreateKey(t.Context(), "awg-b"); err != nil {
		t.Fatalf("create key: %v", err)
	}
	if !slices.Equal(ctr.execIDs, []string{"awg-b"}) {
		t.Errorf("want 1 write in awg-b, got %v", ctr.execIDs)
	}
}

func TestCreateKeyNeedsOneContainerWithoutName(t *testing.T) {

	ctr := &fakeContainerAPI{files: serverFiles()}
	_, err := New(ctr, "203.0.113.7").CreateKey(t.Context(), "")
	if !errors.Is(err, ErrNoProtocol) {
		t.Errorf("no containers: want ErrNoProtocol, got %v", err)
	}

	ctr.containers = []container.Container{awgContainer(t, awgConfig("awg-b")), awgContainer(t, awgConfig("awg-a"))}
	_, err = New(ctr, "203.0.113.7").CreateKey(t.Context(), "")
	if !errors.Is(err, ErrManyProtocols) {
		t.Fatalf("two containers: want ErrManyProtocols, got %v", err)
	}
	if !strings.Contains(err.Error(), "awg-a, awg-b") {
		t.Errorf("want the sorted names in the error, got %v", err)
	}
	if len(ctr.execIDs) != 0 {
		t.Errorf("want no write, got %v", ctr.execIDs)
	}
}

func TestCreateKeyFailsOnUnknownName(t *testing.T) {

	ctr := &fakeContainerAPI{
		containers: []container.Container{awgContainer(t, awgConfig("awg-a"))},
		files:      serverFiles(),
	}

	_, err := New(ctr, "203.0.113.7").CreateKey(t.Context(), "nope")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestCreateKeyNeedsEndpoint(t *testing.T) {

	ctr := &fakeContainerAPI{
		containers: []container.Container{awgContainer(t, awgConfig("awg-a"))},
		files:      serverFiles(),
	}

	_, err := New(ctr, "").CreateKey(t.Context(), "")
	if !errors.Is(err, ErrNoEndpoint) {
		t.Errorf("want ErrNoEndpoint, got %v", err)
	}
	if len(ctr.execIDs) != 0 {
		t.Errorf("want no write without an endpoint, got %v", ctr.execIDs)
	}
}

func TestCreateKeyStopsOnBadEnv(t *testing.T) {

	c := awgContainer(t, awgConfig("awg-a"))
	delete(c.Env, "AWG_SERVER_PORT")
	ctr := &fakeContainerAPI{containers: []container.Container{c}, files: serverFiles()}

	_, err := New(ctr, "203.0.113.7").CreateKey(t.Context(), "")
	if !errors.Is(err, awg.ErrBadEnv) {
		t.Errorf("want awg.ErrBadEnv, got %v", err)
	}
	if len(ctr.execIDs) != 0 {
		t.Errorf("want no write with a bad env, got %v", ctr.execIDs)
	}
}

func TestDiscoverSkipsUnknownProtocols(t *testing.T) {

	unknown := container.Container{Name: "jao-new", Labels: map[string]string{protocol.Label: "new"}}
	ctr := &fakeContainerAPI{
		containers: []container.Container{awgContainer(t, awgConfig("awg-b")), unknown, awgContainer(t, awgConfig("awg-a"))},
	}

	got, err := New(ctr, "203.0.113.7").Discover(t.Context())
	if err != nil {
		t.Fatalf("discover: %v", err)
	}

	want := []Installed{{Name: "awg-a", Protocol: protocol.AWGLabel}, {Name: "awg-b", Protocol: protocol.AWGLabel}}
	if !slices.Equal(got, want) {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestDiscoverStopsOnListError(t *testing.T) {

	listErr := errors.New("docker is down")
	ctr := &fakeContainerAPI{listErr: listErr}

	_, err := New(ctr, "203.0.113.7").Discover(t.Context())
	if !errors.Is(err, listErr) {
		t.Errorf("want list error, got %v", err)
	}
}

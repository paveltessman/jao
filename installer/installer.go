package installer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/installer/internal/workdir"
)

type ContainerAPI interface {
	ImageBuild(ctx context.Context, workdir fs.FS, imageTag string) error
	ContainerCreate(ctx context.Context, spec container.ContainerSpec) (string, error)
	ContainerStart(ctx context.Context, id string) error
	ContainerExec(ctx context.Context, id string, cmd []string) (string, error)
}

var _ ContainerAPI = (*container.Service)(nil)

type Service struct {
	ctr     ContainerAPI
	workDir string

	// peerMu makes sure that two peers do not get the same IP.
	peerMu sync.Mutex
}

func New(ctr ContainerAPI, workDir string) *Service {
	svc := Service{ctr: ctr, workDir: workDir}
	return &svc
}

// InstallAwg builds the AWG image and starts a container from cfg.
// The server keys stay in <workDir>/<cfg.Name> on the host.
func (s *Service) InstallAwg(ctx context.Context, cfg AwgConfig) error {
	workdir, err := workdir.FS(workdir.AWG)
	if err != nil {
		return err
	}

	dataDir, err := s.dataDir(cfg.Name)
	if err != nil {
		return err
	}

	err = s.ctr.ImageBuild(ctx, workdir, awgImageTag)
	if err != nil {
		return err
	}

	ctrID, err := s.ctr.ContainerCreate(ctx, cfg.spec(dataDir))
	if err != nil {
		return err
	}

	err = s.ctr.ContainerStart(ctx, ctrID)
	if err != nil {
		return err
	}

	return nil
}

var ErrNoEndpoint = errors.New("installer: endpoint host is empty")

// addAwgPeerScript appends the peer in $1 to awg0.conf and applies the file
// to awg0 without a restart. The peer comes as an argument, not as part of
// the script, so bash does not read it.
const addAwgPeerScript = `printf '%s' "$1" >> ` + awgConfPath + ` && awg syncconf awg0 <(awg-quick strip ` + awgConfPath + `)`

// AddAwgPeer adds a new peer to the AWG container from cfg and returns
// the client config for it. endpointHost is the public address of the node.
// The client private key is not kept on the server.
func (s *Service) AddAwgPeer(ctx context.Context, cfg AwgConfig, endpointHost string) (string, error) {
	if endpointHost == "" {
		return "", ErrNoEndpoint
	}

	s.peerMu.Lock()
	defer s.peerMu.Unlock()

	serverConf, err := s.readFile(ctx, cfg.Name, awgConfPath)
	if err != nil {
		return "", err
	}
	serverPublicKey, err := s.readFile(ctx, cfg.Name, awgServerPublicKeyPath)
	if err != nil {
		return "", err
	}
	psk, err := s.readFile(ctx, cfg.Name, awgPskPath)
	if err != nil {
		return "", err
	}

	ip, err := nextPeerIP(serverConf, cfg.SubnetIP, cfg.SubnetCIDR)
	if err != nil {
		return "", err
	}
	privateKey, publicKey, err := genKeyPair()
	if err != nil {
		return "", err
	}

	peer := fmt.Sprintf("\n[Peer]\nPublicKey = %s\nPresharedKey = %s\nAllowedIPs = %s/32\n", publicKey, psk, ip)
	_, err = s.ctr.ContainerExec(ctx, cfg.Name, []string{"bash", "-c", addAwgPeerScript, "_", peer})
	if err != nil {
		return "", fmt.Errorf("installer: can't add peer %s to %s: %w", ip, cfg.Name, err)
	}

	client := awgClient{
		PrivateKey:      privateKey,
		IP:              ip,
		ServerPublicKey: serverPublicKey,
		PresharedKey:    psk,
	}
	return cfg.clientConfig(client, endpointHost), nil
}

// readFile returns the content of a file in the container without
// the spaces at the ends.
func (s *Service) readFile(ctx context.Context, ctrName, path string) (string, error) {
	out, err := s.ctr.ContainerExec(ctx, ctrName, []string{"cat", path})
	if err != nil {
		return "", fmt.Errorf("installer: can't read %s in %s: %w", path, ctrName, err)
	}
	return strings.TrimSpace(out), nil
}

// dataDir creates the host dir for a container and returns its absolute path.
// Docker needs an absolute path for a bind mount.
func (s *Service) dataDir(name string) (string, error) {
	dir, err := filepath.Abs(filepath.Join(s.workDir, name))
	if err != nil {
		return "", fmt.Errorf("installer: can't resolve data dir for %s: %w", name, err)
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("installer: can't create data dir %s: %w", dir, err)
	}

	return dir, nil
}

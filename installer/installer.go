package installer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/paveltessman/jao/installer/internal/workdir"
)

type ContainerAPI interface {
	ImageBuild(ctx context.Context, workdir fs.FS, imageTag string) error
	ContainerCreate(ctx context.Context, spec ContainerSpec) (string, error)
	ContainerStart(ctx context.Context, id string) error
}

// ContainerSpec describes a container to create.
type ContainerSpec struct {
	Name    string
	Image   string
	Env     map[string]string
	CapAdd  []string
	Devices []string // host paths, mapped to the same path in the container
	Ports   []PortBinding
	Mounts  []Mount
	Sysctls map[string]string
	Restart RestartPolicy
}

type Protocol string

const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
)

// PortBinding publishes a container port on a host port.
type PortBinding struct {
	HostPort      uint16
	ContainerPort uint16
	Protocol      Protocol
}

// Mount binds a host directory to a path in the container.
type Mount struct {
	Source string
	Target string
}

type RestartPolicy string

const (
	RestartNo            RestartPolicy = "no"
	RestartUnlessStopped RestartPolicy = "unless-stopped"
)

type Service struct {
	ctr     ContainerAPI
	workDir string
}

func New(ctr ContainerAPI, workDir string) *Service {
	svc := Service{ctr, workDir}
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

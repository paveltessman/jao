package installer

import (
	"context"
	"io/fs"

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
	ctr ContainerAPI
}

func New(ctr ContainerAPI) *Service {
	svc := Service{ctr}
	return &svc
}

func (s *Service) InstallAwg(ctx context.Context) error {
	imageTag := "awg:test"
	workdir, err := workdir.FS(workdir.AWG)
	if err != nil {
		return err
	}

	err = s.ctr.ImageBuild(ctx, workdir, imageTag)
	if err != nil {
		return err
	}

	ctrID, err := s.ctr.ContainerCreate(ctx, ContainerSpec{Image: imageTag})
	if err != nil {
		return err
	}

	err = s.ctr.ContainerStart(ctx, ctrID)
	if err != nil {
		return err
	}

	return nil
}

package installer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/protocol/awg"
)

type ContainerAPI interface {
	ImageBuild(ctx context.Context, workdir fs.FS, imageTag string) error
	ContainerCreate(ctx context.Context, spec container.ContainerSpec) (string, error)
	ContainerStart(ctx context.Context, id string) error
}

var _ ContainerAPI = (*container.Service)(nil)

type Service struct {
	ctr     ContainerAPI
	workDir string
}

func New(ctr ContainerAPI, workDir string) *Service {
	svc := Service{ctr: ctr, workDir: workDir}
	return &svc
}

// InstallAwg builds the AWG image and starts a container from cfg.
func (s *Service) InstallAwg(ctx context.Context, cfg awg.Config) error {
	buildCtx, err := awg.BuildContext()
	if err != nil {
		return err
	}

	dataDir, err := s.createDataDir(cfg.Name)
	if err != nil {
		return err
	}

	err = s.ctr.ImageBuild(ctx, buildCtx, awg.ImageTag)
	if err != nil {
		return err
	}

	ctrID, err := s.ctr.ContainerCreate(ctx, cfg.Spec(dataDir))
	if err != nil {
		return err
	}

	err = s.ctr.ContainerStart(ctx, ctrID)
	if err != nil {
		return err
	}

	return nil
}

// createDataDir creates the host dir for a container and returns its absolute path.
func (s *Service) createDataDir(name string) (string, error) {
	dir, err := filepath.Abs(filepath.Join(s.workDir, name))
	if err != nil {
		return "", fmt.Errorf("installer: can't resolve data dir for %s: %w", name, err)
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("installer: can't create data dir %s: %w", dir, err)
	}

	return dir, nil
}

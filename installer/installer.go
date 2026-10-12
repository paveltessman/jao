package installer

import (
	"context"
	"io/fs"

	"github.com/paveltessman/jao/installer/internal/workdir"
)

type ContainerAPI interface {
	ImageBuild(ctx context.Context, workdir fs.FS, imageTag string) error
	ContainerCreate(ctx context.Context, imageTag string) (string, error)
	ContainerStart(ctx context.Context, id string) error
}

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

	ctrID, err := s.ctr.ContainerCreate(ctx, imageTag)
	if err != nil {
		return err
	}

	err = s.ctr.ContainerStart(ctx, ctrID)
	if err != nil {
		return err
	}

	return nil
}

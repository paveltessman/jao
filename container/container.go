package container

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/paveltessman/jao/installer"
)

var ErrBuildContext = errors.New("can't create build context")
var ErrImageBuild = errors.New("can't build image")
var ErrContainerCreate = errors.New("can't create container")
var ErrContainerStart = errors.New("can't start container")

var _ installer.ContainerAPI = installer.ContainerAPI(nil)

type Service struct {
	client *client.Client
}

func New() (*Service, error) {
	apiClient, err := client.New(
		client.FromEnv,
	)
	if err != nil {
		return nil, err
	}

	svc := Service{apiClient}
	return &svc, nil
}

func (s *Service) Close() error {
	return s.client.Close()
}

type PingData struct {
	APIVersion   string
	OSType       string
	Experimental bool
}

func (p PingData) String() string {
	s := fmt.Sprintf("version=%s, os=%s, exp=%v", p.APIVersion, p.OSType, p.Experimental)
	return s
}

func (s *Service) Ping(ctx context.Context) (PingData, error) {
	ping, err := s.client.Ping(ctx, client.PingOptions{NegotiateAPIVersion: true})
	return PingData{ping.APIVersion, ping.OSType, ping.Experimental}, err
}

func (s *Service) ImageBuild(ctx context.Context, workdir fs.FS, imageTag string) error {
	buildCtx, err := tarDir(workdir)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBuildContext, err)
	}

	res, err := s.client.ImageBuild(ctx, buildCtx, client.ImageBuildOptions{
		Tags:        []string{imageTag},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrImageBuild, err)
	}
	defer func() { _ = res.Body.Close() }()

	err = readErr(res.Body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrImageBuild, err)
	}

	return nil
}

func (s *Service) ContainerCreate(ctx context.Context, imageTag string) (string, error) {
	created, err := s.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image: imageTag,
		},
		HostConfig: &container.HostConfig{},
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContainerCreate, err)
	}
	for _, w := range created.Warnings {
		log.Println(w)
	}
	return created.ID, nil
}

func (s *Service) ContainerStart(ctx context.Context, id string) error {
	_, err := s.client.ContainerStart(ctx, id, client.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrContainerStart, err)
	}
	return nil
}

func tarDir(dir fs.FS) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	if err := tw.AddFS(dir); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

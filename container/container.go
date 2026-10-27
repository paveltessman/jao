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
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

var ErrBuildContext = errors.New("can't create build context")
var ErrImageBuild = errors.New("can't build image")
var ErrContainerCreate = errors.New("can't create container")
var ErrContainerStart = errors.New("can't start container")
var ErrContainerExec = errors.New("can't exec in container")
var ErrContainerRemove = errors.New("can't remove container")
var ErrContainerList = errors.New("can't list containers")

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

func (s *Service) ContainerCreate(ctx context.Context, spec ContainerSpec) (string, error) {
	cfg, hostCfg, err := toDocker(spec)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContainerCreate, err)
	}

	created, err := s.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Name:       spec.Name,
		Config:     cfg,
		HostConfig: hostCfg,
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

// ContainerExec runs cmd in the container and returns its stdout.
// A non-zero exit code is an error that holds the stderr.
func (s *Service) ContainerExec(ctx context.Context, id string, cmd []string) (string, error) {
	exec, err := s.client.ExecCreate(ctx, id, client.ExecCreateOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContainerExec, err)
	}

	attached, err := s.client.ExecAttach(ctx, exec.ID, client.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContainerExec, err)
	}
	defer attached.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, attached.Reader); err != nil {
		return "", fmt.Errorf("%w: %w", ErrContainerExec, err)
	}

	inspect, err := s.client.ExecInspect(ctx, exec.ID, client.ExecInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrContainerExec, err)
	}
	if inspect.ExitCode != 0 {
		return "", fmt.Errorf("%w: exit code %d: %s", ErrContainerExec, inspect.ExitCode, stderr.String())
	}

	return stdout.String(), nil
}

// ContainerRemove stops the container if it runs and removes it.
func (s *Service) ContainerRemove(ctx context.Context, id string) error {
	_, err := s.client.ContainerRemove(ctx, id, client.ContainerRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrContainerRemove, err)
	}
	return nil
}

// ContainerList returns all containers that have the label key,
// with any value. The list also holds the containers that do not run.
func (s *Service) ContainerList(ctx context.Context, label string) ([]Container, error) {
	list, err := s.client.ContainerList(ctx, client.ContainerListOptions{
		All:     true,
		Filters: client.Filters{}.Add("label", label),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrContainerList, err)
	}

	// The list does not hold the env, so each container needs an inspect.
	var containers []Container
	for _, item := range list.Items {
		inspect, err := s.client.ContainerInspect(ctx, item.ID, client.ContainerInspectOptions{})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrContainerList, err)
		}
		c := inspect.Container

		env := map[string]string{}
		if c.Config != nil {
			for _, kv := range c.Config.Env {
				key, value, _ := strings.Cut(kv, "=")
				env[key] = value
			}
		}

		containers = append(containers, Container{
			ID:     c.ID,
			Name:   strings.TrimPrefix(c.Name, "/"),
			Labels: item.Labels,
			Env:    env,
		})
	}
	return containers, nil
}

func toDocker(spec ContainerSpec) (*container.Config, *container.HostConfig, error) {
	cfg := container.Config{
		Image:        spec.Image,
		Labels:       spec.Labels,
		ExposedPorts: network.PortSet{},
	}
	for _, key := range slices.Sorted(maps.Keys(spec.Env)) {
		cfg.Env = append(cfg.Env, key+"="+spec.Env[key])
	}

	hostCfg := container.HostConfig{
		CapAdd:        spec.CapAdd,
		Sysctls:       spec.Sysctls,
		PortBindings:  network.PortMap{},
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyMode(spec.Restart)},
	}
	for _, dev := range spec.Devices {
		hostCfg.Devices = append(hostCfg.Devices, container.DeviceMapping{
			PathOnHost:        dev,
			PathInContainer:   dev,
			CgroupPermissions: "rwm",
		})
	}
	for _, m := range spec.Mounts {
		hostCfg.Mounts = append(hostCfg.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: m.Source,
			Target: m.Target,
		})
	}
	for _, pb := range spec.Ports {
		port, ok := network.PortFrom(pb.ContainerPort, network.IPProtocol(pb.Protocol))
		if !ok {
			return nil, nil, fmt.Errorf("bad port %d/%s", pb.ContainerPort, pb.Protocol)
		}
		cfg.ExposedPorts[port] = struct{}{}
		hostCfg.PortBindings[port] = append(hostCfg.PortBindings[port], network.PortBinding{
			HostPort: strconv.Itoa(int(pb.HostPort)),
		})
	}

	return &cfg, &hostCfg, nil
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

package container

import (
	"context"
	"fmt"

	"github.com/moby/moby/client"
)

type Service struct {
	client *client.Client
}

func New() (*Service, error) {
	apiClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
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
	ping, err := s.client.Ping(ctx)
	return PingData{ping.APIVersion, ping.OSType, ping.Experimental}, err
}

// Package keys finds the protocol containers on the node and creates keys in them.
package keys

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/protocol"
	"github.com/paveltessman/jao/protocol/awg"
)

type ContainerAPI interface {
	ContainerList(ctx context.Context, label string) ([]container.Container, error)
	ContainerExec(ctx context.Context, id string, cmd []string) (string, error)
}

var _ ContainerAPI = (*container.Service)(nil)

var ErrNoEndpoint = errors.New("keys: endpoint host is empty")
var ErrNoProtocol = errors.New("keys: no protocol is installed")
var ErrManyProtocols = errors.New("keys: more than one protocol is installed, give a name")
var ErrNotFound = errors.New("keys: protocol not found")

// Installed describes a protocol container on the node.
type Installed struct {
	Name     string
	Protocol string
}

type Service struct {
	ctr          ContainerAPI
	endpointHost string

	// peerMu makes sure that two peers do not get the same IP.
	peerMu sync.Mutex
}

// New returns a key service. endpointHost is the public address of the node.
func New(ctr ContainerAPI, endpointHost string) *Service {
	svc := Service{ctr: ctr, endpointHost: endpointHost}
	return &svc
}

// Discover returns the protocol containers on the node, sorted by name.
// It skips a container with a protocol that this build does not know.
func (s *Service) Discover(ctx context.Context) ([]Installed, error) {
	found, err := s.discover(ctx)
	if err != nil {
		return nil, err
	}

	installed := make([]Installed, 0, len(found))
	for _, c := range found {
		installed = append(installed, Installed{Name: c.Name, Protocol: c.Labels[protocol.Label]})
	}
	return installed, nil
}

// CreateKey creates a key in the protocol container with the given name
// and returns the client config. An empty name is correct only when
// the node has exactly one protocol container.
func (s *Service) CreateKey(ctx context.Context, name string) (string, error) {
	if s.endpointHost == "" {
		return "", ErrNoEndpoint
	}

	found, err := s.discover(ctx)
	if err != nil {
		return "", err
	}
	c, err := pick(found, name)
	if err != nil {
		return "", err
	}

	s.peerMu.Lock()
	defer s.peerMu.Unlock()

	switch c.Labels[protocol.Label] {
	case protocol.AWG:
		cfg, err := awg.ConfigFromEnv(c.Name, c.Env)
		if err != nil {
			return "", err
		}
		return awg.AddPeer(ctx, s.ctr, cfg, s.endpointHost)
	}
	// discover returns only the known protocols.
	return "", fmt.Errorf("keys: unknown protocol %q in %s", c.Labels[protocol.Label], c.Name)
}

// discover returns the containers with a known protocol, sorted by name.
func (s *Service) discover(ctx context.Context) ([]container.Container, error) {
	list, err := s.ctr.ContainerList(ctx, protocol.Label)
	if err != nil {
		return nil, err
	}

	var found []container.Container
	for _, c := range list {
		switch c.Labels[protocol.Label] {
		case protocol.AWG:
			found = append(found, c)
		}
	}
	slices.SortFunc(found, func(a, b container.Container) int { return strings.Compare(a.Name, b.Name) })
	return found, nil
}

// pick returns the container with the given name, or the only container
// when the name is empty.
func pick(found []container.Container, name string) (container.Container, error) {
	if name == "" {
		switch len(found) {
		case 0:
			return container.Container{}, ErrNoProtocol
		case 1:
			return found[0], nil
		}
		names := make([]string, 0, len(found))
		for _, c := range found {
			names = append(names, c.Name)
		}
		return container.Container{}, fmt.Errorf("%w: %s", ErrManyProtocols, strings.Join(names, ", "))
	}

	i := slices.IndexFunc(found, func(c container.Container) bool { return c.Name == name })
	if i < 0 {
		return container.Container{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return found[i], nil
}

package api

import (
	"fmt"
	"net/http"
)

func (s *Service) listContainers(r *http.Request) (Payload, error) {

	installed, err := s.deps.KeyService.Discover(r.Context())
	if err != nil {
		return nil, err
	}

	return installed, nil
}

func (s *Service) createKey(r *http.Request) (Payload, error) {

	containerName := r.PathValue("name")
	if len(containerName) == 0 {
		return nil, fmt.Errorf("%w: container name is not specified", ErrBadPathValue)
	}

	clientConf, err := s.deps.KeyService.CreateKey(r.Context(), containerName)
	if err != nil {
		return nil, err
	}

	return clientConf, err
}

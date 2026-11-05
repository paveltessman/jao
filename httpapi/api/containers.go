package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/paveltessman/jao/httpapi/shared"
)

var ErrBadPathValue = errors.New("bad path value")

func listContainers(deps shared.Deps) EndpointFunc {
	endpoint := func(r *http.Request) (Payload, error) {
		installed, err := deps.KeyService.Discover(r.Context())
		if err != nil {
			return nil, err
		}
		return installed, nil
	}
	return endpoint
}

func createKey(deps shared.Deps) EndpointFunc {
	endpoint := func(r *http.Request) (Payload, error) {
		containerName := r.PathValue("name")
		if len(containerName) == 0 {
			return nil, fmt.Errorf("%w: container name is not specified", ErrBadPathValue)
		}

		clientConf, err := deps.KeyService.CreateKey(r.Context(), containerName)
		if err != nil {
			return nil, err
		}

		return clientConf, err
	}
	return endpoint
}

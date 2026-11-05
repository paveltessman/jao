package api

import (
	"net/http"

	"github.com/paveltessman/jao/httpapi/shared"
)

func Route(deps shared.Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /containers/{name}/", Wrap(createKey(deps)))
	mux.Handle("GET /containers", Wrap(listContainers(deps)))

	return mux
}

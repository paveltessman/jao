package api

import (
	"net/http"

	"github.com/paveltessman/jao/httpapi/shared"
)

func Route(deps shared.Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /containers/{name}/", createKey(deps))
	mux.Handle("GET /containers", listContainers(deps))

	return mux
}

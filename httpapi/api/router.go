package api

import (
	"net/http"

	"github.com/paveltessman/jao/httpapi/shared"
)

type Service struct {
	deps shared.Deps
}

func Route(deps shared.Deps) http.Handler {
	mux := http.NewServeMux()

	svc := Service{deps}

	mux.Handle("POST /containers/{name}/", Wrap(svc.createKey))
	mux.Handle("GET /containers", Wrap(svc.listContainers))

	return mux
}

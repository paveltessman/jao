package api

import (
	"net/http"

	"github.com/paveltessman/jao/httpserver/shared"
)

type Service struct {
	deps shared.Deps
}

func Route(deps shared.Deps) http.Handler {
	mux := http.NewServeMux()

	svc := Service{deps}

	mux.Handle("POST /containers/{name}/", wrap(svc.createKey))
	mux.Handle("GET /containers", wrap(svc.listContainers))

	return mux
}

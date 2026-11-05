package httpapi

import (
	"net/http"

	"github.com/paveltessman/jao/httpapi/api"
	"github.com/paveltessman/jao/httpapi/shared"
)

func mainRoute(deps shared.Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api.Route(deps)))

	return mux
}

package httpapi

import (
	"net/http"

	"github.com/paveltessman/jao/httpapi/api"
	"github.com/paveltessman/jao/httpapi/middleware"
	"github.com/paveltessman/jao/httpapi/shared"
	"github.com/paveltessman/jao/httpapi/ui"
)

func mainRoute(deps shared.Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", middleware.Auth(deps.AuthService, api.Route(deps))))
	mux.Handle("/ui/", http.StripPrefix("/ui", ui.Route()))

	return mux
}

package httpserver

import (
	"net/http"

	"github.com/paveltessman/jao/httpserver/api"
	"github.com/paveltessman/jao/httpserver/auth"
	"github.com/paveltessman/jao/httpserver/middleware"
	"github.com/paveltessman/jao/httpserver/shared"
	"github.com/paveltessman/jao/httpserver/ui"
)

func mainRoute(deps shared.Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", middleware.Auth(deps.AuthService, api.Route(deps))))
	mux.Handle("/ui/", http.StripPrefix("/ui", ui.Route()))
	mux.Handle("/auth/", http.StripPrefix("/auth", auth.Route(deps)))

	return mux
}

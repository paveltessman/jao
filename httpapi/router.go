package httpapi

import (
	"fmt"
	"net/http"
)

func list(deps Deps) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, "\n")
		if err != nil {
			panic(err)
		}
	}
	return http.HandlerFunc(handler)
}

func apiRoute(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/list", list(deps))

	return mux
}

func mainRoute(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiRoute(deps)))

	return mux
}

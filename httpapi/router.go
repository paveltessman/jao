package httpapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func createKey(deps Deps) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		containerName := r.PathValue("containerName")
		if len(containerName) == 0 {
			http.Error(w, "container name is not specified", http.StatusBadRequest)
		}

		clientConf, err := deps.keyService.CreateKey(r.Context(), containerName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		_, err = fmt.Fprint(w, clientConf)
		if err != nil {
			log.Println(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
	return http.HandlerFunc(handler)
}

func list(deps Deps) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		installed, err := deps.keyService.Discover(r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return

		}

		err = json.NewEncoder(w).Encode(installed)
		if err != nil {
			log.Println(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	return http.HandlerFunc(handler)
}

func apiRoute(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /list", list(deps))
	mux.Handle("POST /key/{containerName}", createKey(deps))

	return mux
}

func mainRoute(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiRoute(deps)))

	return mux
}

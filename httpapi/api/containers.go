package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/paveltessman/jao/httpapi/shared"
)

func listContainers(deps shared.Deps) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		installed, err := deps.KeyService.Discover(r.Context())
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

func createKey(deps shared.Deps) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		containerName := r.PathValue("name")
		if len(containerName) == 0 {
			http.Error(w, "container name is not specified", http.StatusBadRequest)
		}

		clientConf, err := deps.KeyService.CreateKey(r.Context(), containerName)
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

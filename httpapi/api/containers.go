package api

import (
	"log"
	"net/http"

	"github.com/paveltessman/jao/httpapi/shared"
)

func listContainers(deps shared.Deps) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		installed, err := deps.KeyService.Discover(r.Context())
		if err != nil {
			log.Println(err)
			writeJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "internal server error",
			})
			return

		}

		writeJSON(w, http.StatusOK, Response{
			Success: true,
			Data:    installed,
		})
	}

	return http.HandlerFunc(handler)
}

func createKey(deps shared.Deps) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		containerName := r.PathValue("name")
		if len(containerName) == 0 {
			writeJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Error:   "container name is not specified",
			})
			return
		}

		clientConf, err := deps.KeyService.CreateKey(r.Context(), containerName)
		if err != nil {
			log.Println(err)
			writeJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "internal server error, see details in the server logs",
			})
			return
		}

		writeJSON(w, http.StatusOK, Response{
			Success: true,
			Data:    clientConf,
		})
	}
	return http.HandlerFunc(handler)
}

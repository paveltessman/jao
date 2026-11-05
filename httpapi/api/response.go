package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type Payload interface{}
type EndpointFunc func(r *http.Request) (Payload, error)

type Response struct {
	Success bool    `json:"success"`
	Data    Payload `json:"data,omitempty"`
	Error   string  `json:"error,omitempty"`
}

func Wrap(endpoint EndpointFunc) http.Handler {

	handler := func(w http.ResponseWriter, r *http.Request) {

		payload, err := endpoint(r)

		logError(err)

		writeJSON(w, statusCode(err), Response{
			Success: err == nil,
			Data:    payload,
			Error:   errorMessage(err),
		})
	}
	return http.HandlerFunc(handler)
}

func writeJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "error": "failed to encode response"}`, http.StatusInternalServerError)
	}
}

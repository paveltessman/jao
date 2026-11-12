package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Payload any
type EndpointFunc func(r *http.Request) (Payload, error)

type Response struct {
	Success bool    `json:"success"`
	Data    Payload `json:"data,omitempty"`
	Error   string  `json:"error,omitempty"`
}

func wrap(endpoint EndpointFunc) http.Handler {

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
	b, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = fmt.Fprintln(w, string(b))
	if err != nil {
		log.Println(err)
	}
}

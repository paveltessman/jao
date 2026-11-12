package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/paveltessman/jao/auth"
	"github.com/paveltessman/jao/httpapi/shared"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleLogin(svc *auth.Service) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req LoginRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Bad request: invalid json payload", http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" {
			http.Error(w, "Bad request: username and password are required", http.StatusBadRequest)
			return
		}

		_, tokens, err := svc.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		setRefreshCookie(w, tokens.Refresh)

		_, err = fmt.Fprintln(w, tokens.Access)
		if err != nil {
			log.Println(err)
		}
	}

	return http.HandlerFunc(handler)
}

func handleRefresh(svc *auth.Service) http.Handler {

	handler := func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized: missing refresh token", http.StatusUnauthorized)
			return
		}

		tokens, err := svc.Refresh(r.Context(), cookie.Value)
		if err != nil {
			log.Println(err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		setRefreshCookie(w, tokens.Refresh)

		_, err = fmt.Fprintln(w, tokens.Access)
		if err != nil {
			log.Println(err)
		}
	}

	return http.HandlerFunc(handler)
}

func setRefreshCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

}

func Route(deps shared.Deps) http.Handler {

	mux := http.NewServeMux()

	mux.Handle("POST /refresh", handleRefresh(deps.AuthService))
	mux.Handle("POST /login", handleLogin(deps.AuthService))

	return mux
}

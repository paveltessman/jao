package middleware

import (
	"net/http"
	"strings"

	"github.com/paveltessman/jao/auth"
)

func Auth(svc *auth.Service, next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")

		_, err := svc.Authenticate(r.Context(), accessToken)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(handler)
}

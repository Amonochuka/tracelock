package http

import (
	"net/http"
	"tracelock/internal/auth"
)

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			claims, ok := r.Context().Value(auth.UserContextKey).(*auth.UserClaims)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Role != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

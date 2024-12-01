package auth

import (
	"net/http"
)

type contextKey string

const UserContextKey = contextKey("user")

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// verify the OAuth token/session
		// For now, just a skeleton

		// If authentication fails:
		// w.WriteHeader(http.StatusUnauthorized)
		// return

		// If authentication succeeds:
		next.ServeHTTP(w, r)
	})
}

package server

import (
	"crypto/subtle"
	"net/http"
	"os"
)

func adminVerificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := r.Header.Get("X-Admin-Secret")
		expected := os.Getenv("ADMIN_SECRET")

		if subtle.ConstantTimeCompare([]byte(secret), []byte(expected)) != 1 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

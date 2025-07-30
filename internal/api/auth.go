package api

import (
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization")
			return
		}

		ok, err := service.VerifyAPIKey(token)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid api key")
			return
		}

		next(w, r)
	}
}

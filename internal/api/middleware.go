package api

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func withAuth(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			writeError(w, http.StatusUnauthorized, "Missing Authorization")
			return
		}

		ok, err := service.VerifyAPIKey(token)
		if err != nil {
			if errors.Is(err, service.ErrNoKeyStore) {
				log.Printf("missing key store")
				writeError(w, http.StatusInternalServerError, "Internal Server Error")
			} else {
				writeError(w, http.StatusUnauthorized, "Unauthorized")
			}
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next(w, r)
	}
}

func withCORS(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		var err error
		if origin != "" {
			allowed, err = service.IsAllowedOrigin(origin)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Internal Server Error")
				return
			}
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			if allowed {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}

		next(w, r)
	}
}

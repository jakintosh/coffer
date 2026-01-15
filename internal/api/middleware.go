package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func (a *API) withCORS(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		var err error
		if origin != "" {
			allowed, err = a.svc.IsAllowedOrigin(origin)
			if err != nil {
				wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
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

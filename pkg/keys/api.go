package keys

import (
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

type APIResponse = wire.Response

// Router registers key management routes and returns an auth middleware.
// Routes added: POST {prefix}/keys, DELETE {prefix}/keys/{id}
// The returned middleware can be used to protect other routes with key authentication.
func (s *Service) Router(
	mux *http.ServeMux,
	prefix string,
) {
	mux.HandleFunc("POST "+prefix+"/keys", s.WithAuth(s.handleCreate))
	mux.HandleFunc("DELETE "+prefix+"/keys/{id}", s.WithAuth(s.handleDelete))
}

func (s *Service) handleCreate(
	w http.ResponseWriter,
	r *http.Request,
) {
	token, err := s.Create()
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	wire.WriteData(w, http.StatusCreated, token)
}

func (s *Service) handleDelete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		wire.WriteError(w, http.StatusBadRequest, "Missing Key ID")
		return
	}
	if err := s.Delete(id); err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) WithAuth(
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			wire.WriteError(w, http.StatusUnauthorized, "Missing Authorization")
			return
		}

		ok, err := s.Verify(token)
		if err != nil || !ok {
			wire.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next(w, r)
	}
}

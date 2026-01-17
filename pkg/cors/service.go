// Package cors provides CORS origin management with SQL storage and HTTP handlers.
package cors

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

var ErrInvalidOrigin = errors.New("invalid origin URL")

// AllowedOrigin represents a permitted CORS origin.
type AllowedOrigin struct {
	URL string `json:"url"`
}

// Store defines the persistence interface for CORS origins.
type Store interface {
	Count() (int, error)
	Get() ([]AllowedOrigin, error)
	Set([]AllowedOrigin) error
}

// Options configures a cors Service.
type Options struct {
	Store          Store
	InitialOrigins []string
}

// Service provides CORS origin management.
type Service struct {
	store Store
}

// New creates a Service with the provided options.
// If InitialOrigins is provided and the store is empty, it seeds the store.
func New(
	opts Options,
) (
	*Service,
	error,
) {
	if opts.Store == nil {
		return nil, errors.New("cors: store required")
	}

	service := &Service{
		store: opts.Store,
	}

	if len(opts.InitialOrigins) == 0 {
		return service, nil
	}

	count, err := service.store.Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		var list []AllowedOrigin
		for _, o := range opts.InitialOrigins {
			if t := strings.TrimSpace(o); t != "" {
				list = append(list, AllowedOrigin{URL: t})
			}
		}
		if len(list) > 0 {
			if err := service.SetOrigins(list); err != nil {
				return nil, err
			}
		}
	}

	return service, nil
}

// GetOrigins returns all allowed origins.
func (s *Service) GetOrigins() (
	[]AllowedOrigin,
	error,
) {
	return s.store.Get()
}

// SetOrigins replaces the allowed origins list.
// Returns ErrInvalidOrigin if any origin doesn't start with http:// or https://.
func (s *Service) SetOrigins(
	origins []AllowedOrigin,
) error {
	for _, o := range origins {
		if !strings.HasPrefix(o.URL, "http://") && !strings.HasPrefix(o.URL, "https://") {
			return ErrInvalidOrigin
		}
	}
	return s.store.Set(origins)
}

// IsAllowed checks if an origin is in the allowed list.
func (s *Service) IsAllowed(
	origin string,
) (
	bool,
	error,
) {
	origins, err := s.store.Get()
	if err != nil {
		return false, err
	}
	for _, o := range origins {
		if o.URL == origin {
			return true, nil
		}
	}
	return false, nil
}

// WithCORS returns HTTP middleware that handles CORS headers and preflight requests.
// It checks the Origin header against the allowed origins list and sets appropriate
// CORS headers for allowed origins.
func (s *Service) WithCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		var err error
		if origin != "" {
			allowed, err = s.IsAllowed(origin)
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

// Router registers CORS management routes.
// Routes added: GET {prefix}/cors, PUT {prefix}/cors
// The auth middleware wraps handlers for authentication.
func (s *Service) Router(
	mux *http.ServeMux,
	prefix string,
	auth func(http.HandlerFunc) http.HandlerFunc,
) {
	mux.HandleFunc("GET "+prefix+"/cors", auth(s.handleGet))
	mux.HandleFunc("PUT "+prefix+"/cors", auth(s.handlePut))
}

func (s *Service) handleGet(w http.ResponseWriter, r *http.Request) {
	origins, err := s.GetOrigins()
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	// Ensure we return an empty array, not null
	if origins == nil {
		origins = []AllowedOrigin{}
	}
	wire.WriteData(w, http.StatusOK, origins)
}

func (s *Service) handlePut(w http.ResponseWriter, r *http.Request) {
	var origins []AllowedOrigin
	if err := json.NewDecoder(r.Body).Decode(&origins); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	if err := s.SetOrigins(origins); err != nil {
		if errors.Is(err, ErrInvalidOrigin) {
			wire.WriteError(w, http.StatusBadRequest, "Invalid Origin URL")
		} else {
			wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

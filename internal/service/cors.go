package service

import (
	"strings"
)

type AllowedOrigin struct {
	URL string `json:"url"`
}

type CORSStore interface {
	CountOrigins() (int, error)
	GetOrigins() ([]AllowedOrigin, error)
	SetOrigins([]AllowedOrigin) error
}

func (s *Service) InitCORS(
	origins []string,
) error {
	if s == nil || s.CORS == nil {
		return ErrNoCORSStore
	}

	count, err := s.CORS.CountOrigins()
	if err != nil {
		return DatabaseError{err}
	}
	if count > 0 {
		return nil
	}

	var list []AllowedOrigin
	for _, o := range origins {
		if t := strings.TrimSpace(o); t != "" {
			list = append(list, AllowedOrigin{URL: t})
		}
	}
	if len(list) == 0 {
		return nil
	}

	if err := s.CORS.SetOrigins(list); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func (s *Service) GetAllowedOrigins() (
	[]AllowedOrigin,
	error,
) {
	if s == nil || s.CORS == nil {
		return nil, ErrNoCORSStore
	}

	origins, err := s.CORS.GetOrigins()
	if err != nil {
		return nil, DatabaseError{err}
	}

	return origins, nil
}

func (s *Service) SetAllowedOrigins(
	origins []AllowedOrigin,
) error {
	if s == nil || s.CORS == nil {
		return ErrNoCORSStore
	}

	for _, o := range origins {
		if strings.HasPrefix(o.URL, "http://") {
			continue
		}
		if strings.HasPrefix(o.URL, "https://") {
			continue
		}
		return ErrInvalidOrigin
	}

	if err := s.CORS.SetOrigins(origins); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func (s *Service) IsAllowedOrigin(
	origin string,
) (
	bool,
	error,
) {
	if s == nil || s.CORS == nil {
		return false, ErrNoCORSStore
	}

	origins, err := s.CORS.GetOrigins()
	if err != nil {
		return false, DatabaseError{err}
	}

	for _, o := range origins {
		if o.URL == origin {
			return true, nil
		}
	}
	return false, nil
}

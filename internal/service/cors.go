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

func (s *Service) initCORS(
	origins []string,
) error {
	count, err := s.cors.CountOrigins()
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

	if err := s.cors.SetOrigins(list); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func (s *Service) GetAllowedOrigins() (
	[]AllowedOrigin,
	error,
) {
	origins, err := s.cors.GetOrigins()
	if err != nil {
		return nil, DatabaseError{err}
	}

	return origins, nil
}

func (s *Service) SetAllowedOrigins(
	origins []AllowedOrigin,
) error {
	for _, o := range origins {
		if strings.HasPrefix(o.URL, "http://") {
			continue
		}
		if strings.HasPrefix(o.URL, "https://") {
			continue
		}
		return ErrInvalidOrigin
	}

	if err := s.cors.SetOrigins(origins); err != nil {
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
	origins, err := s.cors.GetOrigins()
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

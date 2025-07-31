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

var corsStore CORSStore

func SetCORSStore(s CORSStore) {
	corsStore = s
}

func InitCORS(
	origins []string,
) error {
	if corsStore == nil {
		return ErrNoCORSStore
	}

	count, err := corsStore.CountOrigins()
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

	if err := corsStore.SetOrigins(list); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func GetAllowedOrigins() (
	[]AllowedOrigin,
	error,
) {
	if corsStore == nil {
		return nil, ErrNoCORSStore
	}

	origins, err := corsStore.GetOrigins()
	if err != nil {
		return nil, DatabaseError{err}
	}

	return origins, nil
}

func SetAllowedOrigins(
	origins []AllowedOrigin,
) error {
	if corsStore == nil {
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

	if err := corsStore.SetOrigins(origins); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func IsAllowedOrigin(
	origin string,
) (
	bool,
	error,
) {
	if corsStore == nil {
		return false, ErrNoCORSStore
	}

	origins, err := corsStore.GetOrigins()
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

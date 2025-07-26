package service

import (
	"errors"
	"fmt"
)

// ErrInvalidDate is returned when the provided date string cannot be parsed.
var ErrInvalidDate = errors.New("invalid date format")

// DatabaseError wraps database level errors so that callers can
// differentiate them from validation errors.
type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }

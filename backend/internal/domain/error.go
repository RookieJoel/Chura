package domain

import (
	"errors"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidID       = errors.New("invalid id")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("conflict")
)

// Violation describes one invalid field.
type Violation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// InvalidInputError carries every offending field; it matches ErrInvalidInput.
type InvalidInputError struct {
	Violations []Violation
}

func (e *InvalidInputError) Error() string {
	return "invalid input"
}

func (e *InvalidInputError) Is(target error) bool {
	return target == ErrInvalidInput
}

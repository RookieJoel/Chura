package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidID       = errors.New("invalid id")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("conflict")
	ErrUnavailable     = errors.New("unavailable")

	// ErrNotProjectMember means the Project exists but the Team Member is not
	// one of its members. It wraps ErrForbidden, so errors.Is matches both;
	// transports that know it map it to a more specific message first.
	ErrNotProjectMember = fmt.Errorf("not a project member: %w", ErrForbidden)

	// ErrAlreadyMember means the user being added already belongs to the
	// Project. It wraps ErrConflict, so errors.Is matches both.
	ErrAlreadyMember = fmt.Errorf("already a project member: %w", ErrConflict)
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

package domain

import (
	"errors"
)

var (
	ErrNotFound       = errors.New("work item not found")
	ErrSprintNotFound = errors.New("sprint not found")
)

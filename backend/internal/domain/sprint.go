package domain

import (
	"errors"
	"strings"
	"time"
)

type SprintStatus string

const (
	Planned   SprintStatus = "planned"
	Active    SprintStatus = "active"
	Completed SprintStatus = "completed"
)

type Sprint struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Team      string       `json:"team"`
	StartDate *time.Time   `json:"start_date,omitempty"`
	EndDate   *time.Time   `json:"end_date,omitempty"`
	Status    SprintStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
	UpdatedAt time.Time    `json:"updated_at,omitempty"`
}

func (s *Sprint) Validate() error {
	if s == nil {
		return errors.New("sprint is required")
	}

	if strings.TrimSpace(s.Name) == "" {
		return errors.New("name is required")
	}

	if len(s.Name) > 120 {
		return errors.New("name must not exceed 120 characters")
	}

	if strings.TrimSpace(s.Team) == "" {
		return errors.New("team is required")
	}

	if len(s.Team) > 120 {
		return errors.New("team must not exceed 120 characters")
	}

	if s.Status == "" {
		s.Status = Planned
	}

	switch s.Status {
	case Planned, Active, Completed:
		// valid
	default:
		return errors.New("invalid sprint status")
	}

	if s.StartDate != nil && s.EndDate != nil {
		if s.EndDate.Before(*s.StartDate) {
			return errors.New("end_date must be greater than or equal to start_date")
		}
	}

	return nil
}

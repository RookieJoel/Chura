package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type SprintStatus string

const (
	Planned   SprintStatus = "planned"
	Active    SprintStatus = "active"
	Completed SprintStatus = "completed"
)

const (
	MaxSprintNameRunes = 120
	MaxSprintTeamRunes = 120
)

type Sprint struct {
	ID        string       `json:"id"`
	ProjectID string       `json:"project_id"`
	Name      string       `json:"name"`
	Team      string       `json:"team"`
	StartDate *time.Time   `json:"start_date,omitempty"`
	EndDate   *time.Time   `json:"end_date,omitempty"`
	Status    SprintStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
	UpdatedAt time.Time    `json:"updated_at,omitempty"`
}

// Validate defaults an empty Status to Planned and reports every invalid
// field as an *InvalidInputError.
func (s *Sprint) Validate() error {
	if s == nil {
		return &InvalidInputError{Violations: []Violation{{Field: "sprint", Message: "sprint is required"}}}
	}
	if s.Status == "" {
		s.Status = Planned
	}
	var violations []Violation
	violations = appendTextViolation(violations, "name", s.Name, MaxSprintNameRunes)
	violations = appendTextViolation(violations, "team", s.Team, MaxSprintTeamRunes)
	switch s.Status {
	case Planned, Active, Completed:
	default:
		violations = append(violations, Violation{Field: "status", Message: fmt.Sprintf("status must be one of %s, %s, %s", Planned, Active, Completed)})
	}
	if s.StartDate != nil && s.EndDate != nil && s.EndDate.Before(*s.StartDate) {
		violations = append(violations, Violation{Field: "end_date", Message: "end_date must be greater than or equal to start_date"})
	}
	if len(violations) > 0 {
		return &InvalidInputError{Violations: violations}
	}
	return nil
}

// appendTextViolation checks a required, length-limited text field.
func appendTextViolation(violations []Violation, field, value string, maxRunes int) []Violation {
	switch {
	case strings.TrimSpace(value) == "":
		return append(violations, Violation{Field: field, Message: field + " is required"})
	case utf8.RuneCountInString(value) > maxRunes:
		return append(violations, Violation{Field: field, Message: fmt.Sprintf("%s must be at most %d characters", field, maxRunes)})
	}
	return violations
}

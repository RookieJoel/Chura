package domain

import "time"

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
}

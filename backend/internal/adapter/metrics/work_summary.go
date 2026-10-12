package metrics

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// SprintWorkSummaryCalculator supplies the quantitative part of the sprint
// review summary (planned, completed, carry-over and blocked work).
//
// Work items are not linked to sprints yet, so it returns fixed placeholder
// numbers. Once they are, count the sprint's work items by status here.
type SprintWorkSummaryCalculator struct{}

func NewSprintWorkSummaryCalculator() *SprintWorkSummaryCalculator {
	return &SprintWorkSummaryCalculator{}
}

func (c *SprintWorkSummaryCalculator) GetSprintWorkSummary(
	_ context.Context,
	_ string,
) (domain.SprintWorkSummary, error) {

	return domain.SprintWorkSummary{
		Planned:   10,
		Completed: 7,
		CarryOver: 2,
		Blocked:   1,
	}, nil
}

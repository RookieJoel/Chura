package metrics

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// MockWorkSummaryProvider returns fixed numbers until work items can be
// linked to sprints. Replace it with a real provider that implements
// out.SprintWorkSummaryProvider.
type MockWorkSummaryProvider struct{}

func NewMockWorkSummaryProvider() *MockWorkSummaryProvider {
	return &MockWorkSummaryProvider{}
}

func (p *MockWorkSummaryProvider) GetSprintWorkSummary(
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

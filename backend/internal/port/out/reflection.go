package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type ReflectionRepository interface {
	ListBySprintID(ctx context.Context, sprintID string) ([]domain.SprintReflection, error)
}

type SprintWorkSummaryProvider interface {
	GetSprintWorkSummary(ctx context.Context, sprintID string) (domain.SprintWorkSummary, error)
}

type ReportExporter interface {
	Export(summary *domain.SprintReviewSummary) ([]byte, error)
}

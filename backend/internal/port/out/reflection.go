package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type ReflectionRepository interface {
	ListBySprintID(ctx context.Context, sprintID string) ([]domain.SprintReflection, error)
}

type ReportExporter interface {
	Export(summary *domain.SprintReviewSummary) ([]byte, error)
}

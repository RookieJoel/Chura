package in

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type AnalyticsService interface {
	RetrieveSprintReflection(ctx context.Context, sprintID string) ([]domain.SprintReflection, error)
	GenerateSprintReviewSummary(ctx context.Context, sprintID string) (*domain.SprintReviewSummary, error)
	ExportSummaryReport(ctx context.Context, sprintID string) ([]byte, error)
}

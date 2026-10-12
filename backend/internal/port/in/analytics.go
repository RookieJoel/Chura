package in

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// AnalyticsReportingService turns a Sprint's work items and the team's
// reflections into metrics and Sprint Review Summaries. Reads need Project
// membership (or the Auditor role); CreateSprintReflection additionally needs
// the Team Member role. Sprints of invisible Projects are ErrNotFound or
// ErrNotProjectMember, as in SprintService.
type AnalyticsReportingService interface {
	CalculateSprintWorkSummary(ctx context.Context, actor domain.Actor, sprintID string) (*domain.SprintWorkSummary, error)
	RetrieveSprintMetrics(ctx context.Context, actor domain.Actor, sprintID string) (*domain.SprintMetrics, error)
	// CreateSprintReflection validates answers against the Project Template's
	// Reflection topics and stores them as the actor's reflection, replacing any
	// earlier one.
	CreateSprintReflection(ctx context.Context, actor domain.Actor, sprintID string, answers map[string]string) (*domain.SprintReflection, error)
	RetrieveSprintReflection(ctx context.Context, actor domain.Actor, sprintID string) ([]domain.SprintReflection, error)
	// GenerateSprintReviewSummary shows the requested metrics, or the Template
	// defaults when metrics is empty.
	GenerateSprintReviewSummary(ctx context.Context, actor domain.Actor, sprintID string, metrics []domain.MetricKey) (*domain.SprintReviewSummary, error)
	ExportSummaryReport(ctx context.Context, actor domain.Actor, sprintID string, format domain.ReportFormat, metrics []domain.MetricKey) (*domain.ExportedReport, error)
}

package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// SprintWorkItemReader lists the work items committed to a Sprint.
type SprintWorkItemReader interface {
	ListBySprint(ctx context.Context, projectID, sprintID string) ([]domain.WorkItem, error)
}

// ReflectionRepository stores at most one reflection per (sprint, author).
type ReflectionRepository interface {
	// Upsert inserts the reflection or replaces the author's earlier answers
	// for the same Sprint, and returns the stored row.
	Upsert(ctx context.Context, reflection *domain.SprintReflection) (*domain.SprintReflection, error)
	// ListBySprint returns the Sprint's reflections, oldest first.
	ListBySprint(ctx context.Context, sprintID string) ([]domain.SprintReflection, error)
}

// TemplateGateway lets other services read a Template definition.
type TemplateGateway interface {
	RetrieveTemplateDefinition(ctx context.Context, actor domain.Actor, templateID string) (domain.Template, error)
}

// ReportExporter renders a Sprint Review Summary in one file format.
type ReportExporter interface {
	Format() domain.ReportFormat
	ContentType() string
	FileExtension() string
	Export(summary *domain.SprintReviewSummary) ([]byte, error)
}

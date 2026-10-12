package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type AnalyticsReportingService struct {
	sprints     out.SprintRepository
	projects    out.ProjectGateway
	templates   out.TemplateGateway
	workItems   out.SprintWorkItemReader
	reflections out.ReflectionRepository
	exporters   map[domain.ReportFormat]out.ReportExporter
	now         func() time.Time
}

var (
	_ in.AnalyticsReportingService = (*AnalyticsReportingService)(nil)
	_ out.TemplateGateway          = (*ProjectConfigurationService)(nil)
)

// AnalyticsOption customises an AnalyticsReportingService (used by tests).
type AnalyticsOption func(*AnalyticsReportingService)

func WithAnalyticsClock(now func() time.Time) AnalyticsOption {
	return func(s *AnalyticsReportingService) { s.now = now }
}

func NewAnalyticsReportingService(
	sprints out.SprintRepository,
	projects out.ProjectGateway,
	templates out.TemplateGateway,
	workItems out.SprintWorkItemReader,
	reflections out.ReflectionRepository,
	exporters []out.ReportExporter,
	opts ...AnalyticsOption,
) *AnalyticsReportingService {
	s := &AnalyticsReportingService{
		sprints:     sprints,
		projects:    projects,
		templates:   templates,
		workItems:   workItems,
		reflections: reflections,
		exporters:   make(map[domain.ReportFormat]out.ReportExporter, len(exporters)),
		now:         func() time.Time { return time.Now().UTC() },
	}
	for _, e := range exporters {
		s.exporters[e.Format()] = e
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *AnalyticsReportingService) CalculateSprintWorkSummary(ctx context.Context, actor domain.Actor, sprintID string) (*domain.SprintWorkSummary, error) {
	scope, err := s.loadVisible(ctx, actor, sprintID)
	if err != nil {
		return nil, fmt.Errorf("calculate sprint work summary: %w", err)
	}
	items, err := s.workItems.ListBySprint(ctx, scope.project.ID, scope.sprint.ID)
	if err != nil {
		return nil, fmt.Errorf("calculate sprint work summary: %w", err)
	}
	summary := domain.CalculateSprintWorkSummary(items)
	return &summary, nil
}

func (s *AnalyticsReportingService) RetrieveSprintMetrics(ctx context.Context, actor domain.Actor, sprintID string) (*domain.SprintMetrics, error) {
	scope, err := s.loadVisible(ctx, actor, sprintID)
	if err != nil {
		return nil, fmt.Errorf("retrieve sprint metrics: %w", err)
	}
	items, err := s.workItems.ListBySprint(ctx, scope.project.ID, scope.sprint.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieve sprint metrics: %w", err)
	}
	metrics := domain.CalculateSprintMetrics(*scope.sprint, items, s.now())
	return &metrics, nil
}

func (s *AnalyticsReportingService) CreateSprintReflection(ctx context.Context, actor domain.Actor, sprintID string, answers map[string]string) (*domain.SprintReflection, error) {
	scope, err := s.loadVisible(ctx, actor, sprintID)
	if err != nil {
		return nil, fmt.Errorf("create sprint reflection: %w", err)
	}
	if err := actor.RequireTeamMember(); err != nil {
		return nil, fmt.Errorf("create sprint reflection: %w", err)
	}
	template, err := s.templateOf(ctx, actor, scope.project)
	if err != nil {
		return nil, fmt.Errorf("create sprint reflection: %w", err)
	}
	clean, err := template.Reflection.ValidateAnswers(answers)
	if err != nil {
		return nil, fmt.Errorf("create sprint reflection: %w", err)
	}
	now := s.now()
	stored, err := s.reflections.Upsert(ctx, &domain.SprintReflection{
		SprintID:    scope.sprint.ID,
		ProjectID:   scope.project.ID,
		AuthorID:    actor.UserID,
		AuthorEmail: actor.Email,
		TemplateID:  template.ID,
		Answers:     clean,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return nil, fmt.Errorf("create sprint reflection: %w", err)
	}
	return stored, nil
}

func (s *AnalyticsReportingService) RetrieveSprintReflection(ctx context.Context, actor domain.Actor, sprintID string) ([]domain.SprintReflection, error) {
	scope, err := s.loadVisible(ctx, actor, sprintID)
	if err != nil {
		return nil, fmt.Errorf("retrieve sprint reflection: %w", err)
	}
	reflections, err := s.reflections.ListBySprint(ctx, scope.sprint.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieve sprint reflection: %w", err)
	}
	return reflections, nil
}

func (s *AnalyticsReportingService) GenerateSprintReviewSummary(ctx context.Context, actor domain.Actor, sprintID string, metrics []domain.MetricKey) (*domain.SprintReviewSummary, error) {
	summary, err := s.buildSummary(ctx, actor, sprintID, metrics)
	if err != nil {
		return nil, fmt.Errorf("generate sprint review summary: %w", err)
	}
	return summary, nil
}

func (s *AnalyticsReportingService) ExportSummaryReport(ctx context.Context, actor domain.Actor, sprintID string, format domain.ReportFormat, metrics []domain.MetricKey) (*domain.ExportedReport, error) {
	exporter, ok := s.exporters[format]
	if !ok {
		return nil, fmt.Errorf("export summary report: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "format", Message: fmt.Sprintf("format %q is not supported", format)},
		}})
	}
	summary, err := s.buildSummary(ctx, actor, sprintID, metrics)
	if err != nil {
		return nil, fmt.Errorf("export summary report: %w", err)
	}
	data, err := exporter.Export(summary)
	if err != nil {
		return nil, fmt.Errorf("export summary report: %w", err)
	}
	return &domain.ExportedReport{
		Filename:    fmt.Sprintf("sprint-%s-review.%s", summary.Sprint.ID, exporter.FileExtension()),
		ContentType: exporter.ContentType(),
		Data:        data,
	}, nil
}

// buildSummary gathers the Sprint, its Template, work items and reflections.
// Metric selection is checked before any work item or reflection is read.
func (s *AnalyticsReportingService) buildSummary(ctx context.Context, actor domain.Actor, sprintID string, metrics []domain.MetricKey) (*domain.SprintReviewSummary, error) {
	scope, err := s.loadVisible(ctx, actor, sprintID)
	if err != nil {
		return nil, err
	}
	template, err := s.templateOf(ctx, actor, scope.project)
	if err != nil {
		return nil, err
	}
	selected, err := template.SprintReview.SelectMetrics(metrics)
	if err != nil {
		return nil, err
	}
	items, err := s.workItems.ListBySprint(ctx, scope.project.ID, scope.sprint.ID)
	if err != nil {
		return nil, err
	}
	reflections, err := s.reflections.ListBySprint(ctx, scope.sprint.ID)
	if err != nil {
		return nil, err
	}
	summary := domain.BuildSprintReviewSummary(domain.ReviewInput{
		Project:     *scope.project,
		Sprint:      *scope.sprint,
		Template:    template,
		Metrics:     selected,
		WorkItems:   items,
		Reflections: reflections,
		Now:         s.now(),
	})
	return &summary, nil
}

type sprintScope struct {
	sprint  *domain.Sprint
	project *domain.Project
}

// loadVisible returns the Sprint and its Project when the actor may see the
// Project, with the same rules as SprintService: a missing Sprint is
// ErrNotFound and a non-member gets ErrNotProjectMember from the gateway.
func (s *AnalyticsReportingService) loadVisible(ctx context.Context, actor domain.Actor, sprintID string) (sprintScope, error) {
	if err := actor.Validate(); err != nil {
		return sprintScope{}, err
	}
	sprint, err := s.sprints.GetByID(ctx, sprintID)
	if err != nil {
		return sprintScope{}, err
	}
	project, err := s.projects.RetrieveProjectData(ctx, actor, sprint.ProjectID)
	if err != nil {
		return sprintScope{}, err
	}
	return sprintScope{sprint: sprint, project: project}, nil
}

// templateOf asks the Project Configuration Service for the Project's
// Template. A stored Project whose Template is unknown is a data-integrity
// fault, reported as an internal error rather than a client 404.
func (s *AnalyticsReportingService) templateOf(ctx context.Context, actor domain.Actor, project *domain.Project) (domain.Template, error) {
	template, err := s.templates.RetrieveTemplateDefinition(ctx, actor, project.TemplateID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Template{}, fmt.Errorf("project %s references unknown template %q", project.ID, project.TemplateID)
	}
	if err != nil {
		return domain.Template{}, err
	}
	return template, nil
}

package service

import (
	"context"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type AnalyticsService struct {
	sprints     out.SprintRepository
	reflections out.ReflectionRepository
	workSummary out.SprintWorkSummaryProvider
	exporter    out.ReportExporter
}

func NewAnalyticsService(
	sprints out.SprintRepository,
	reflections out.ReflectionRepository,
	workSummary out.SprintWorkSummaryProvider,
	exporter out.ReportExporter,
) *AnalyticsService {
	return &AnalyticsService{
		sprints:     sprints,
		reflections: reflections,
		workSummary: workSummary,
		exporter:    exporter,
	}
}

func (s *AnalyticsService) RetrieveSprintReflection(
	ctx context.Context,
	sprintID string,
) ([]domain.SprintReflection, error) {

	if _, err := s.findSprint(ctx, sprintID); err != nil {
		return nil, err
	}

	return s.reflections.ListBySprintID(ctx, sprintID)
}

func (s *AnalyticsService) GenerateSprintReviewSummary(
	ctx context.Context,
	sprintID string,
) (*domain.SprintReviewSummary, error) {

	sprint, err := s.findSprint(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	reflections, err := s.reflections.ListBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	work, err := s.workSummary.GetSprintWorkSummary(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	return &domain.SprintReviewSummary{
		Sprint:       *sprint,
		Quantitative: domain.QuantitativeSummary{Work: work},
		Qualitative:  domain.QualitativeSummary{Reflections: reflections},
		GeneratedAt:  time.Now(),
	}, nil
}

func (s *AnalyticsService) ExportSummaryReport(
	ctx context.Context,
	sprintID string,
) ([]byte, error) {

	summary, err := s.GenerateSprintReviewSummary(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	return s.exporter.Export(summary)
}

func (s *AnalyticsService) findSprint(
	ctx context.Context,
	sprintID string,
) (*domain.Sprint, error) {

	sprint, err := s.sprints.GetByID(ctx, sprintID)
	if err != nil {
		return nil, err
	}
	if sprint == nil {
		return nil, domain.ErrSprintNotFound
	}

	return sprint, nil
}

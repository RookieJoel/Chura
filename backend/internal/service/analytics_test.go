package service

import (
	"context"
	"errors"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type fakeSprintRepo struct{ sprint *domain.Sprint }

func (f fakeSprintRepo) Create(context.Context, *domain.Sprint) (*domain.Sprint, error) {
	return nil, nil
}

func (f fakeSprintRepo) GetByID(context.Context, string) (*domain.Sprint, error) {
	return f.sprint, nil
}

func (f fakeSprintRepo) List(context.Context) ([]domain.Sprint, error) { return nil, nil }

func (f fakeSprintRepo) Update(context.Context, string, *domain.Sprint) (*domain.Sprint, error) {
	return nil, nil
}

func (f fakeSprintRepo) Delete(context.Context, string) error { return nil }

type fakeReflectionRepo struct{ items []domain.SprintReflection }

func (f fakeReflectionRepo) ListBySprintID(context.Context, string) ([]domain.SprintReflection, error) {
	return f.items, nil
}

type fakeExporter struct{}

func (fakeExporter) Export(s *domain.SprintReviewSummary) ([]byte, error) {
	return []byte(s.Sprint.Name), nil
}

func newTestAnalytics(sprint *domain.Sprint, items []domain.SprintReflection) *AnalyticsService {
	return NewAnalyticsService(
		fakeSprintRepo{sprint: sprint},
		fakeReflectionRepo{items: items},
		fakeExporter{},
	)
}

func TestRetrieveSprintReflection(t *testing.T) {
	items := []domain.SprintReflection{{ID: "1", SprintID: "1", Author: "a", Content: "good"}}
	svc := newTestAnalytics(&domain.Sprint{ID: "1", Name: "S1"}, items)

	got, err := svc.RetrieveSprintReflection(context.Background(), "1")
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v, err %v", got, err)
	}
}

func TestRetrieveSprintReflectionSprintNotFound(t *testing.T) {
	svc := newTestAnalytics(nil, nil)

	_, err := svc.RetrieveSprintReflection(context.Background(), "9")
	if !errors.Is(err, domain.ErrSprintNotFound) {
		t.Fatalf("want ErrSprintNotFound, got %v", err)
	}
}

func TestGenerateSprintReviewSummary(t *testing.T) {
	items := []domain.SprintReflection{{ID: "1"}, {ID: "2"}}
	svc := newTestAnalytics(&domain.Sprint{ID: "1", Name: "S1"}, items)

	got, err := svc.GenerateSprintReviewSummary(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ReflectionCount != 2 || got.Sprint.Name != "S1" {
		t.Fatalf("unexpected summary %+v", got)
	}
}

func TestExportSummaryReport(t *testing.T) {
	svc := newTestAnalytics(&domain.Sprint{ID: "1", Name: "S1"}, nil)

	got, err := svc.ExportSummaryReport(context.Background(), "1")
	if err != nil || string(got) != "S1" {
		t.Fatalf("got %q, err %v", got, err)
	}
}

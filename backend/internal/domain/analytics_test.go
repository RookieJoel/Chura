package domain_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func item(id string, status domain.WorkItemStatus, points int32) domain.WorkItem {
	return domain.WorkItem{ID: id, Title: "item " + id, Type: domain.WorkItemTypeTask, Status: status, StoryPoints: points}
}

var sprintItems = []domain.WorkItem{
	item("1", domain.WorkItemStatusDone, 5),
	item("2", domain.WorkItemStatusDone, 3),
	item("3", domain.WorkItemStatusInProgress, 2),
	item("4", domain.WorkItemStatusToDo, 1),
	item("5", domain.WorkItemStatusReview, 2),
	item("6", domain.WorkItemStatusBlocked, 8),
}

func TestCalculateSprintWorkSummary_CategorizesEveryItem(t *testing.T) {
	got := domain.CalculateSprintWorkSummary(sprintItems)

	want := domain.SprintWorkSummary{
		Planned:   domain.WorkBucket{Items: 6, Points: 21},
		Completed: domain.WorkBucket{Items: 2, Points: 8},
		CarryOver: domain.WorkBucket{Items: 3, Points: 5},
		Blocked:   domain.WorkBucket{Items: 1, Points: 8},
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if got.CompletionRate() != 33.3 || got.PointsCompletionRate() != 38.1 {
		t.Fatalf("rates = %v / %v, want 33.3 / 38.1", got.CompletionRate(), got.PointsCompletionRate())
	}
}

func TestCalculateSprintWorkSummary_EmptySprintHasZeroRates(t *testing.T) {
	got := domain.CalculateSprintWorkSummary(nil)
	if got != (domain.SprintWorkSummary{}) || got.CompletionRate() != 0 || got.PointsCompletionRate() != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestMetricValue_CoversEveryTemplateMetric(t *testing.T) {
	summary := domain.CalculateSprintWorkSummary(sprintItems)
	for _, m := range domain.SESprintReviewTemplate().Metrics {
		if _, ok := summary.MetricValue(m.Key); !ok {
			t.Errorf("metric %q has no value", m.Key)
		}
	}
	if v, _ := summary.MetricValue(domain.MetricCompletedPoints); v != 8 {
		t.Errorf("completed points = %v, want 8", v)
	}
	if _, ok := summary.MetricValue("bogus"); ok {
		t.Error("unknown metric must report false")
	}
}

func date(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &t
}

func TestCalculateSprintProgress(t *testing.T) {
	sprint := domain.Sprint{StartDate: date(2026, 10, 5), EndDate: date(2026, 10, 16)}
	cases := []struct {
		name string
		now  time.Time
		want domain.SprintProgress
	}{
		{"before start", time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC), domain.SprintProgress{TotalDays: 12, ElapsedDays: 0, RemainingDays: 12, TimeElapsedRate: 0}},
		{"first day", time.Date(2026, 10, 5, 23, 0, 0, 0, time.UTC), domain.SprintProgress{TotalDays: 12, ElapsedDays: 1, RemainingDays: 11, TimeElapsedRate: 8.3}},
		{"middle", time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC), domain.SprintProgress{TotalDays: 12, ElapsedDays: 6, RemainingDays: 6, TimeElapsedRate: 50}},
		{"after end", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), domain.SprintProgress{TotalDays: 12, ElapsedDays: 12, RemainingDays: 0, TimeElapsedRate: 100}},
	}
	for _, tc := range cases {
		got := domain.CalculateSprintProgress(sprint, tc.now)
		if got == nil || *got != tc.want {
			t.Errorf("%s: got %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

func TestCalculateSprintProgress_NilWithoutDates(t *testing.T) {
	if got := domain.CalculateSprintProgress(domain.Sprint{StartDate: date(2026, 10, 5)}, time.Now()); got != nil {
		t.Fatalf("got %+v, want nil", got)
	}
}

func TestCalculateSprintMetrics_BreaksDownItems(t *testing.T) {
	now := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	got := domain.CalculateSprintMetrics(domain.Sprint{ID: "7"}, sprintItems, now)

	if got.Sprint.ID != "7" || got.CompletionRate != 33.3 || got.Progress != nil || !got.CalculatedAt.Equal(now) {
		t.Fatalf("unexpected metrics header: %+v", got)
	}
	wantStatus := map[domain.WorkItemStatus]int{"done": 2, "in_progress": 1, "to_do": 1, "review": 1, "blocked": 1}
	if !reflect.DeepEqual(got.StatusBreakdown, wantStatus) {
		t.Fatalf("status breakdown = %v", got.StatusBreakdown)
	}
	if got.TypeBreakdown[domain.WorkItemTypeTask] != 6 || len(got.WorkItems) != 6 {
		t.Fatalf("type breakdown = %v, %d items", got.TypeBreakdown, len(got.WorkItems))
	}
	if got.WorkItems[5].Category != domain.WorkCategoryBlocked || got.WorkItems[2].Category != domain.WorkCategoryCarryOver {
		t.Fatalf("categories = %+v", got.WorkItems)
	}
}

func TestBuildSprintReviewSummary_ConsolidatesMetricsAndFeedback(t *testing.T) {
	tmpl, _ := domain.TemplateByID(domain.TemplateIDSE)
	selected, _ := tmpl.SprintReview.SelectMetrics([]domain.MetricKey{domain.MetricCompletedItems, domain.MetricPointsCompletionRate})
	now := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)

	got := domain.BuildSprintReviewSummary(domain.ReviewInput{
		Project:   domain.Project{ID: "p1", Name: "Chura", Members: []domain.Member{{UserID: "u1"}, {UserID: "u2"}, {UserID: "u3"}, {UserID: "u4"}}},
		Sprint:    domain.Sprint{ID: "7", Name: "Sprint 1"},
		Template:  tmpl,
		Metrics:   selected,
		WorkItems: sprintItems,
		Reflections: []domain.SprintReflection{
			{AuthorID: "u1", AuthorEmail: "u1@example.com", Answers: map[string]string{"went_well": "demo", "stop": "late merges", "start": "tests"}},
			{AuthorID: "u2", AuthorEmail: "u2@example.com", Answers: map[string]string{"went_well": "pairing"}},
		},
		Now: now,
	})

	wantMetrics := []domain.MetricValue{
		{Key: domain.MetricCompletedItems, Label: "Completed work items", Unit: domain.MetricUnitItems, Value: 2},
		{Key: domain.MetricPointsCompletionRate, Label: "Story point completion rate", Unit: domain.MetricUnitPercent, Value: 38.1},
	}
	if !reflect.DeepEqual(got.Quantitative.Metrics, wantMetrics) {
		t.Fatalf("metrics = %+v", got.Quantitative.Metrics)
	}
	if got.Title != "Sprint Review Summary" || got.Project.Name != "Chura" || got.TemplateID != "se" || !got.GeneratedAt.Equal(now) {
		t.Fatalf("unexpected header: %+v", got)
	}
	q := got.Qualitative
	if q.Respondents != 2 || q.TeamSize != 4 || q.ResponseRate != 50 {
		t.Fatalf("participation = %d/%d %v%%", q.Respondents, q.TeamSize, q.ResponseRate)
	}
	if len(q.Topics) != 3 || q.Topics[0].Key != "went_well" || q.Topics[0].Question != "What went well?" {
		t.Fatalf("topics = %+v", q.Topics)
	}
	if len(q.Topics[0].Responses) != 2 || len(q.Topics[1].Responses) != 1 || q.Topics[1].Responses[0].Answer != "late merges" {
		t.Fatalf("responses grouped wrongly: %+v", q.Topics)
	}
}

func TestBuildSprintReviewSummary_TopicWithoutAnswersIsEmptyNotNil(t *testing.T) {
	tmpl, _ := domain.TemplateByID(domain.TemplateIDGeneral)
	got := domain.BuildSprintReviewSummary(domain.ReviewInput{Template: tmpl})
	for _, topic := range got.Qualitative.Topics {
		if topic.Responses == nil {
			t.Fatalf("topic %s responses are nil (would be JSON null)", topic.Key)
		}
	}
	if got.Qualitative.ResponseRate != 0 {
		t.Fatalf("response rate = %v", got.Qualitative.ResponseRate)
	}
}

func TestParseReportFormat(t *testing.T) {
	for raw, want := range map[string]domain.ReportFormat{"": "csv", "CSV": "csv", "markdown": "markdown", " md ": "markdown"} {
		if got, err := domain.ParseReportFormat(raw); err != nil || got != want {
			t.Errorf("ParseReportFormat(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
	if _, err := domain.ParseReportFormat("pdf"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("pdf: want ErrInvalidInput, got %v", err)
	}
}

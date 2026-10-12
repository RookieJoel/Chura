package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

var (
	analyticsNow   = time.Date(2026, 10, 12, 8, 0, 0, 0, time.UTC)
	auditorActor   = domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}
	validAnswers   = map[string]string{"went_well": " demo went well ", "stop": "late merges", "start": "tests first"}
	analyticsItems = []domain.WorkItem{
		{ID: "w1", Status: domain.WorkItemStatusDone, StoryPoints: 5},
		{ID: "w2", Status: domain.WorkItemStatusInProgress, StoryPoints: 3},
		{ID: "w3", Status: domain.WorkItemStatusBlocked, StoryPoints: 2},
	}
)

type analyticsFixture struct {
	svc         *service.AnalyticsReportingService
	sprints     *fakeSprintRepo
	projects    *fakeGateway
	templates   *fakeTemplates
	workItems   *fakeSprintWorkItems
	reflections *fakeReflections
	exporter    *fakeExporter
}

func newAnalyticsFixture(templateID string) *analyticsFixture {
	f := &analyticsFixture{
		sprints: &fakeSprintRepo{sprints: []domain.Sprint{{ID: "7", ProjectID: sprintProjectID, Name: "Sprint 1", Team: "Alpha", Status: domain.Active}}},
		projects: &fakeGateway{project: &domain.Project{
			ID: sprintProjectID, Name: "Chura", TemplateID: templateID,
			Members: []domain.Member{{UserID: "u1"}, {UserID: "u2"}},
		}},
		templates:   &fakeTemplates{},
		workItems:   &fakeSprintWorkItems{items: analyticsItems},
		reflections: &fakeReflections{},
		exporter:    &fakeExporter{},
	}
	f.svc = service.NewAnalyticsReportingService(f.sprints, f.projects, f.templates, f.workItems, f.reflections,
		[]out.ReportExporter{f.exporter}, service.WithAnalyticsClock(func() time.Time { return analyticsNow }))
	return f
}

func TestCalculateSprintWorkSummary_ReadsTheSprintsItems(t *testing.T) {
	f := newAnalyticsFixture("se")

	got, err := f.svc.CalculateSprintWorkSummary(context.Background(), memberActor, "7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Planned.Items != 3 || got.Completed.Points != 5 || got.CarryOver.Items != 1 || got.Blocked.Items != 1 {
		t.Fatalf("summary = %+v", got)
	}
	if f.workItems.gotProject != sprintProjectID || f.workItems.gotSprint != "7" {
		t.Fatalf("work items read for project %q sprint %q", f.workItems.gotProject, f.workItems.gotSprint)
	}
	if f.projects.gotActr != memberActor || f.projects.gotID != sprintProjectID {
		t.Fatalf("visibility not checked with the caller: %+v", f.projects)
	}
}

func TestRetrieveSprintMetrics_UsesClock(t *testing.T) {
	f := newAnalyticsFixture("se")

	got, err := f.svc.RetrieveSprintMetrics(context.Background(), auditorActor, "7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Sprint.ID != "7" || !got.CalculatedAt.Equal(analyticsNow) || len(got.WorkItems) != 3 || got.CompletionRate != 33.3 {
		t.Fatalf("metrics = %+v", got)
	}
}

func TestAnalyticsReads_FailBeforeReadingData(t *testing.T) {
	cases := []struct {
		name  string
		actor domain.Actor
		id    string
		gwErr error
		want  error
	}{
		{"unauthenticated", domain.Actor{}, "7", nil, domain.ErrUnauthenticated},
		{"missing sprint", memberActor, "404", nil, domain.ErrNotFound},
		{"not a project member", memberActor, "7", domain.ErrNotProjectMember, domain.ErrNotProjectMember},
	}
	for _, tc := range cases {
		f := newAnalyticsFixture("se")
		f.projects.err = tc.gwErr
		ctx := context.Background()

		_, errSummary := f.svc.CalculateSprintWorkSummary(ctx, tc.actor, tc.id)
		_, errMetrics := f.svc.RetrieveSprintMetrics(ctx, tc.actor, tc.id)
		_, errList := f.svc.RetrieveSprintReflection(ctx, tc.actor, tc.id)
		_, errReview := f.svc.GenerateSprintReviewSummary(ctx, tc.actor, tc.id, nil)
		_, errExport := f.svc.ExportSummaryReport(ctx, tc.actor, tc.id, domain.ReportFormatCSV, nil)
		_, errCreate := f.svc.CreateSprintReflection(ctx, tc.actor, tc.id, validAnswers)

		for op, err := range map[string]error{"summary": errSummary, "metrics": errMetrics, "list": errList, "review": errReview, "export": errExport, "create": errCreate} {
			if !errors.Is(err, tc.want) {
				t.Errorf("%s/%s: want %v, got %v", tc.name, op, tc.want, err)
			}
		}
		if f.workItems.calls != 0 || f.reflections.calls != 0 {
			t.Errorf("%s: data read before access check (work items %d, reflections %d)", tc.name, f.workItems.calls, f.reflections.calls)
		}
	}
}

func TestCreateSprintReflection_StoresTrimmedAnswersForTheCaller(t *testing.T) {
	f := newAnalyticsFixture("se")

	got, err := f.svc.CreateSprintReflection(context.Background(), memberActor, "7", validAnswers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "1" || got.SprintID != "7" || got.ProjectID != sprintProjectID || got.TemplateID != "se" {
		t.Fatalf("reflection = %+v", got)
	}
	if got.AuthorID != memberActor.UserID || got.AuthorEmail != memberActor.Email {
		t.Fatalf("author = %s %s", got.AuthorID, got.AuthorEmail)
	}
	if got.Answers["went_well"] != "demo went well" || !got.CreatedAt.Equal(analyticsNow) {
		t.Fatalf("answers = %v created = %v", got.Answers, got.CreatedAt)
	}
	if f.templates.calls != 1 {
		t.Fatalf("template definition read %d times, want 1", f.templates.calls)
	}
}

func TestCreateSprintReflection_SecondSubmitReplacesTheFirst(t *testing.T) {
	f := newAnalyticsFixture("general")
	ctx := context.Background()

	if _, err := f.svc.CreateSprintReflection(ctx, memberActor, "7", validAnswers); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	again := map[string]string{"went_well": "retro", "stop": "x", "start": "y"}
	if _, err := f.svc.CreateSprintReflection(ctx, memberActor, "7", again); err != nil {
		t.Fatalf("second submit: %v", err)
	}

	list, _ := f.svc.RetrieveSprintReflection(ctx, memberActor, "7")
	if len(list) != 1 || list[0].Answers["went_well"] != "retro" {
		t.Fatalf("reflections = %+v", list)
	}
}

func TestCreateSprintReflection_InvalidAnswersAreNotStored(t *testing.T) {
	f := newAnalyticsFixture("se")

	_, err := f.svc.CreateSprintReflection(context.Background(), memberActor, "7", map[string]string{"went_well": "only this"})

	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) || len(invalid.Violations) != 2 {
		t.Fatalf("want 2 violations, got %v", err)
	}
	if f.reflections.calls != 0 {
		t.Fatal("invalid reflection reached the repository")
	}
}

func TestCreateSprintReflection_AuditorIsForbidden(t *testing.T) {
	f := newAnalyticsFixture("se")

	_, err := f.svc.CreateSprintReflection(context.Background(), auditorActor, "7", validAnswers)

	if !errors.Is(err, domain.ErrForbidden) || f.reflections.calls != 0 {
		t.Fatalf("want ErrForbidden and no write, got %v (writes %d)", err, f.reflections.calls)
	}
}

func TestCreateSprintReflection_UnknownProjectTemplateIsInternal(t *testing.T) {
	f := newAnalyticsFixture("deleted-template")

	_, err := f.svc.CreateSprintReflection(context.Background(), memberActor, "7", validAnswers)

	if err == nil || errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want an internal error, got %v", err)
	}
}

func TestGenerateSprintReviewSummary_CombinesMetricsAndReflections(t *testing.T) {
	f := newAnalyticsFixture("se")
	ctx := context.Background()
	if _, err := f.svc.CreateSprintReflection(ctx, memberActor, "7", validAnswers); err != nil {
		t.Fatalf("submit reflection: %v", err)
	}

	got, err := f.svc.GenerateSprintReviewSummary(ctx, auditorActor, "7", []domain.MetricKey{domain.MetricCompletedPoints})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Quantitative.Metrics) != 1 || got.Quantitative.Metrics[0].Value != 5 {
		t.Fatalf("metrics = %+v", got.Quantitative.Metrics)
	}
	if got.Quantitative.WorkSummary.Planned.Items != 3 || got.Project.Name != "Chura" || !got.GeneratedAt.Equal(analyticsNow) {
		t.Fatalf("summary = %+v", got)
	}
	q := got.Qualitative
	if q.Respondents != 1 || q.TeamSize != 2 || q.ResponseRate != 50 || q.Topics[2].Responses[0].Answer != "tests first" {
		t.Fatalf("qualitative = %+v", q)
	}
}

func TestGenerateSprintReviewSummary_DefaultsToTemplateMetrics(t *testing.T) {
	f := newAnalyticsFixture("general")

	got, err := f.svc.GenerateSprintReviewSummary(context.Background(), memberActor, "7", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Quantitative.Metrics) != len(domain.GeneralSprintReviewTemplate().DefaultMetrics) {
		t.Fatalf("got %d metrics", len(got.Quantitative.Metrics))
	}
}

func TestGenerateSprintReviewSummary_RejectsMetricOutsideTemplateBeforeReadingData(t *testing.T) {
	f := newAnalyticsFixture("general")

	_, err := f.svc.GenerateSprintReviewSummary(context.Background(), memberActor, "7", []domain.MetricKey{domain.MetricPlannedPoints})

	if !errors.Is(err, domain.ErrInvalidInput) || f.workItems.calls != 0 {
		t.Fatalf("want ErrInvalidInput without reading work items, got %v (reads %d)", err, f.workItems.calls)
	}
}

func TestGenerateSprintReviewSummary_PropagatesRepositoryFailure(t *testing.T) {
	f := newAnalyticsFixture("se")
	boom := errors.New("mongo down")
	f.workItems.err = boom

	if _, err := f.svc.GenerateSprintReviewSummary(context.Background(), memberActor, "7", nil); !errors.Is(err, boom) {
		t.Fatalf("want wrapped repository error, got %v", err)
	}
}

func TestExportSummaryReport_UsesExporterForFormat(t *testing.T) {
	f := newAnalyticsFixture("se")

	got, err := f.svc.ExportSummaryReport(context.Background(), memberActor, "7", domain.ReportFormatCSV, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Filename != "sprint-7-review.csv" || got.ContentType != "text/csv" || string(got.Data) != "report:Sprint 1" {
		t.Fatalf("report = %+v", got)
	}
	if f.exporter.got == nil || f.exporter.got.Sprint.ID != "7" {
		t.Fatal("exporter did not receive the summary")
	}
}

func TestExportSummaryReport_UnsupportedFormatIsInvalidInput(t *testing.T) {
	f := newAnalyticsFixture("se")

	_, err := f.svc.ExportSummaryReport(context.Background(), memberActor, "7", domain.ReportFormatMarkdown, nil)

	if !errors.Is(err, domain.ErrInvalidInput) || f.workItems.calls != 0 {
		t.Fatalf("want ErrInvalidInput before reading data, got %v", err)
	}
}

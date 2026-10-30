package http_test

import (
	"bytes"
	"context"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"

	handler "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

// stubAnalyticsService embeds the interface so only the methods under test
// are implemented; any other call panics on the nil embedded value.
type stubAnalyticsService struct {
	in.AnalyticsReportingService
	summaryFn func(a domain.Actor, id string) (*domain.SprintWorkSummary, error)
	metricsFn func(a domain.Actor, id string) (*domain.SprintMetrics, error)
	createFn  func(a domain.Actor, id string, answers map[string]string) (*domain.SprintReflection, error)
	listFn    func(a domain.Actor, id string) ([]domain.SprintReflection, error)
	reviewFn  func(a domain.Actor, id string, metrics []domain.MetricKey) (*domain.SprintReviewSummary, error)
	exportFn  func(a domain.Actor, id string, f domain.ReportFormat, metrics []domain.MetricKey) (*domain.ExportedReport, error)
}

func (s stubAnalyticsService) CalculateSprintWorkSummary(_ context.Context, a domain.Actor, id string) (*domain.SprintWorkSummary, error) {
	return s.summaryFn(a, id)
}

func (s stubAnalyticsService) RetrieveSprintMetrics(_ context.Context, a domain.Actor, id string) (*domain.SprintMetrics, error) {
	return s.metricsFn(a, id)
}

func (s stubAnalyticsService) CreateSprintReflection(_ context.Context, a domain.Actor, id string, answers map[string]string) (*domain.SprintReflection, error) {
	return s.createFn(a, id, answers)
}

func (s stubAnalyticsService) RetrieveSprintReflection(_ context.Context, a domain.Actor, id string) ([]domain.SprintReflection, error) {
	return s.listFn(a, id)
}

func (s stubAnalyticsService) GenerateSprintReviewSummary(_ context.Context, a domain.Actor, id string, metrics []domain.MetricKey) (*domain.SprintReviewSummary, error) {
	return s.reviewFn(a, id, metrics)
}

func (s stubAnalyticsService) ExportSummaryReport(_ context.Context, a domain.Actor, id string, f domain.ReportFormat, metrics []domain.MetricKey) (*domain.ExportedReport, error) {
	return s.exportFn(a, id, f, metrics)
}

func analyticsApp(svc in.AnalyticsReportingService) *fiber.App {
	return newTestRouterWithAnalytics(nil, handler.NewTemplateHandler(nil), handler.NewProjectHandler(nil), handler.NewAnalyticsHandler(svc))
}

var auditorHeadersForAnalytics = map[string]string{"X-Test-Actor-ID": "a1", "X-Test-Actor-Email": "a1@example.com", "X-Test-Actor-Role": "auditor"}

func TestWorkSummary_Returns200(t *testing.T) {
	var gotID string
	app := analyticsApp(stubAnalyticsService{summaryFn: func(_ domain.Actor, id string) (*domain.SprintWorkSummary, error) {
		gotID = id
		return &domain.SprintWorkSummary{Planned: domain.WorkBucket{Items: 4, Points: 9}}, nil
	}})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/sprints/7/work-summary", "", memberHeaders)

	planned, _ := body["planned"].(map[string]any)
	if status != 200 || gotID != "7" || planned["items"] != float64(4) || planned["points"] != float64(9) {
		t.Fatalf("status=%d body=%v id=%q", status, body, gotID)
	}
}

func TestMetrics_MapsNotProjectMemberTo403(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{metricsFn: func(domain.Actor, string) (*domain.SprintMetrics, error) {
		return nil, domain.ErrNotProjectMember
	}})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/sprints/7/metrics", "", memberHeaders)

	if status != 403 || body["error"] != "you are not a member of this project" {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestSubmitReflection_PassesAnswersAndActor(t *testing.T) {
	var gotActor domain.Actor
	var gotAnswers map[string]string
	app := analyticsApp(stubAnalyticsService{createFn: func(a domain.Actor, id string, answers map[string]string) (*domain.SprintReflection, error) {
		gotActor, gotAnswers = a, answers
		return &domain.SprintReflection{ID: "1", SprintID: id, AuthorID: a.UserID, Answers: answers}, nil
	}})

	status, body := send(t, app, nethttp.MethodPut, "/api/v1/sprints/7/reflections/me",
		`{"answers":{"went_well":"demo","stop":"late merges","start":"tests"}}`, memberHeaders)

	if status != 200 || body["id"] != "1" || body["sprint_id"] != "7" || body["author_id"] != "u1" {
		t.Fatalf("status=%d body=%v", status, body)
	}
	if gotActor.UserID != "u1" || !reflect.DeepEqual(gotAnswers, map[string]string{"went_well": "demo", "stop": "late merges", "start": "tests"}) {
		t.Fatalf("actor=%+v answers=%v", gotActor, gotAnswers)
	}
}

func TestSubmitReflection_BadJSONIs400WithoutCallingService(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{})

	status, body := send(t, app, nethttp.MethodPut, "/api/v1/sprints/7/reflections/me", `{"answers":`, memberHeaders)

	if status != 400 || body["error"] != "invalid JSON body" {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestSubmitReflection_ViolationsAreReturned(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{createFn: func(domain.Actor, string, map[string]string) (*domain.SprintReflection, error) {
		return nil, &domain.InvalidInputError{Violations: []domain.Violation{{Field: "answers.stop", Message: "required"}}}
	}})

	status, body := send(t, app, nethttp.MethodPut, "/api/v1/sprints/7/reflections/me", `{"answers":{}}`, memberHeaders)

	violations, _ := body["violations"].([]any)
	if status != 400 || len(violations) != 1 {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestSubmitReflection_AuditorForbiddenIs403(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{createFn: func(domain.Actor, string, map[string]string) (*domain.SprintReflection, error) {
		return nil, domain.ErrForbidden
	}})

	status, _ := send(t, app, nethttp.MethodPut, "/api/v1/sprints/7/reflections/me", `{"answers":{}}`, auditorHeadersForAnalytics)

	if status != 403 {
		t.Fatalf("status=%d", status)
	}
}

func TestListReflections_ReturnsArray(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{listFn: func(domain.Actor, string) ([]domain.SprintReflection, error) {
		return []domain.SprintReflection{{ID: "1"}, {ID: "2"}}, nil
	}})

	status, _, arr := get(t, app, "/api/v1/sprints/7/reflections", memberHeaders)

	if status != 200 || len(arr) != 2 {
		t.Fatalf("status=%d arr=%v", status, arr)
	}
}

func TestReviewSummary_ParsesMetricsQuery(t *testing.T) {
	var gotMetrics []domain.MetricKey
	app := analyticsApp(stubAnalyticsService{reviewFn: func(_ domain.Actor, _ string, metrics []domain.MetricKey) (*domain.SprintReviewSummary, error) {
		gotMetrics = metrics
		return &domain.SprintReviewSummary{Title: "Sprint Review Summary"}, nil
	}})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/sprints/7/review-summary?metrics=planned_items,%20completion_rate,,", "", memberHeaders)

	if status != 200 || body["title"] != "Sprint Review Summary" {
		t.Fatalf("status=%d body=%v", status, body)
	}
	if !reflect.DeepEqual(gotMetrics, []domain.MetricKey{"planned_items", "completion_rate"}) {
		t.Fatalf("metrics = %v", gotMetrics)
	}
}

func TestReviewSummary_NoMetricsQueryMeansDefaults(t *testing.T) {
	var gotMetrics []domain.MetricKey
	app := analyticsApp(stubAnalyticsService{reviewFn: func(_ domain.Actor, _ string, metrics []domain.MetricKey) (*domain.SprintReviewSummary, error) {
		gotMetrics = metrics
		return &domain.SprintReviewSummary{}, nil
	}})

	if status, _ := send(t, app, nethttp.MethodGet, "/api/v1/sprints/7/review-summary", "", memberHeaders); status != 200 || gotMetrics != nil {
		t.Fatalf("status=%d metrics=%v", status, gotMetrics)
	}
}

func TestExportReviewSummary_SendsFileWithHeaders(t *testing.T) {
	var gotFormat domain.ReportFormat
	app := analyticsApp(stubAnalyticsService{exportFn: func(_ domain.Actor, _ string, f domain.ReportFormat, _ []domain.MetricKey) (*domain.ExportedReport, error) {
		gotFormat = f
		return &domain.ExportedReport{Filename: "sprint-7-review.md", ContentType: "text/markdown; charset=utf-8", Data: []byte("# Review")}, nil
	}})

	req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/sprints/7/review-summary/export?format=md", bytes.NewReader(nil))
	for k, v := range memberHeaders {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 || string(body) != "# Review" || gotFormat != domain.ReportFormatMarkdown {
		t.Fatalf("status=%d body=%q format=%q", resp.StatusCode, body, gotFormat)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/markdown; charset=utf-8" {
		t.Fatalf("content type = %q", ct)
	}
	if cd := resp.Header.Get("Content-Disposition"); cd != `attachment; filename="sprint-7-review.md"` {
		t.Fatalf("content disposition = %q", cd)
	}
}

func TestExportReviewSummary_SanitizesFilenameAndDisablesSniffing(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{exportFn: func(_ domain.Actor, _ string, _ domain.ReportFormat, _ []domain.MetricKey) (*domain.ExportedReport, error) {
		return &domain.ExportedReport{Filename: "../../etc/pass\"wd\r\nX-Evil: 1.md", ContentType: "text/markdown; charset=utf-8", Data: []byte("# Review")}, nil
	}})

	req := httptest.NewRequest(nethttp.MethodGet, "/api/v1/sprints/7/review-summary/export?format=md", bytes.NewReader(nil))
	for k, v := range memberHeaders {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if cd := resp.Header.Get("Content-Disposition"); cd != `attachment; filename="pass-wd-X-Evil-1.md"` {
		t.Fatalf("content disposition = %q", cd)
	}
	if resp.Header.Get("X-Evil") != "" {
		t.Fatalf("header injection via filename")
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}
}

func TestExportReviewSummary_UnknownFormatIs400WithoutCallingService(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/sprints/7/review-summary/export?format=pdf", "", memberHeaders)

	if status != 400 || body["error"] != "invalid input" {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestAnalyticsRoutes_RequireAuthentication(t *testing.T) {
	app := analyticsApp(stubAnalyticsService{})
	for _, path := range []string{"/work-summary", "/metrics", "/reflections", "/review-summary", "/review-summary/export"} {
		if status, _ := send(t, app, nethttp.MethodGet, "/api/v1/sprints/7"+path, "", nil); status != 401 {
			t.Errorf("%s: status=%d, want 401", path, status)
		}
	}
}

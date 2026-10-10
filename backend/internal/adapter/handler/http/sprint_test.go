package http_test

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"

	handler "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

const sprintProjectID = "11111111-1111-4111-8111-111111111111"

// stubSprintService embeds the interface so only the methods under test are
// implemented; any other call panics on the nil embedded value.
type stubSprintService struct {
	in.SprintService
	createFn func(domain.Actor, *domain.Sprint) (*domain.Sprint, error)
	listFn   func(a domain.Actor, projectID string) ([]domain.Sprint, error)
	getFn    func(a domain.Actor, id string) (*domain.Sprint, error)
	updateFn func(a domain.Actor, id string, s *domain.Sprint) (*domain.Sprint, error)
	deleteFn func(a domain.Actor, id string) error
}

func (s stubSprintService) CreateSprint(_ context.Context, a domain.Actor, sp *domain.Sprint) (*domain.Sprint, error) {
	return s.createFn(a, sp)
}

func (s stubSprintService) ListSprints(_ context.Context, a domain.Actor, projectID string) ([]domain.Sprint, error) {
	return s.listFn(a, projectID)
}

func (s stubSprintService) GetSprint(_ context.Context, a domain.Actor, id string) (*domain.Sprint, error) {
	return s.getFn(a, id)
}

func (s stubSprintService) UpdateSprint(_ context.Context, a domain.Actor, id string, sp *domain.Sprint) (*domain.Sprint, error) {
	return s.updateFn(a, id, sp)
}

func (s stubSprintService) DeleteSprint(_ context.Context, a domain.Actor, id string) error {
	return s.deleteFn(a, id)
}

func sprintApp(svc in.SprintService) *fiber.App {
	return newTestRouter(handler.NewSprintHandler(svc), handler.NewTemplateHandler(nil), handler.NewProjectHandler(nil))
}

func TestCreateSprint_Returns201WithProjectID(t *testing.T) {
	var gotActor domain.Actor
	var gotSprint *domain.Sprint
	app := sprintApp(stubSprintService{createFn: func(a domain.Actor, s *domain.Sprint) (*domain.Sprint, error) {
		gotActor, gotSprint = a, s
		return &domain.Sprint{ID: "7", ProjectID: s.ProjectID, Name: s.Name, Team: s.Team, Status: domain.Planned}, nil
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/sprints",
		fmt.Sprintf(`{"project_id":%q,"name":"Sprint 1","team":"Alpha"}`, sprintProjectID), memberHeaders)

	if status != 201 || body["id"] != "7" || body["project_id"] != sprintProjectID {
		t.Fatalf("status=%d body=%v", status, body)
	}
	if gotActor != (domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}) {
		t.Fatalf("actor = %+v", gotActor)
	}
	if gotSprint.ProjectID != sprintProjectID || gotSprint.Name != "Sprint 1" || gotSprint.Team != "Alpha" {
		t.Fatalf("sprint = %+v", gotSprint)
	}
}

func TestCreateSprint_WithoutActorIs401(t *testing.T) {
	called := false
	app := sprintApp(stubSprintService{createFn: func(domain.Actor, *domain.Sprint) (*domain.Sprint, error) {
		called = true
		return nil, nil
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/sprints",
		fmt.Sprintf(`{"project_id":%q,"name":"Sprint 1","team":"Alpha"}`, sprintProjectID), nil)

	if status != 401 || body["error"] == nil || called {
		t.Fatalf("status=%d body=%v serviceCalled=%v", status, body, called)
	}
}

func TestCreateSprint_MapsServiceErrorsToStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"auditor forbidden", fmt.Errorf("create sprint: %w", domain.ErrForbidden), 403},
		{"project not found", fmt.Errorf("project x: %w", domain.ErrNotFound), 404},
		{"not a project member", fmt.Errorf("project x: %w", domain.ErrNotProjectMember), 403},
		{"bad project id", fmt.Errorf("project id: %w", domain.ErrInvalidID), 400},
		{"invalid sprint", fmt.Errorf("%w: name is required", domain.ErrInvalidInput), 400},
		{"missing project_id", &domain.InvalidInputError{Violations: []domain.Violation{{Field: "project_id", Message: "project_id is required"}}}, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := sprintApp(stubSprintService{createFn: func(domain.Actor, *domain.Sprint) (*domain.Sprint, error) {
				return nil, tc.err
			}})

			status, body := send(t, app, nethttp.MethodPost, "/api/v1/sprints", `{"name":"S","team":"T"}`, memberHeaders)

			if status != tc.status || body["error"] == nil {
				t.Fatalf("status=%d body=%v, want %d", status, body, tc.status)
			}
		})
	}
}

func TestListSprints_PassesActorAndRawProjectIDToService(t *testing.T) {
	var got []string
	var gotActor domain.Actor
	app := sprintApp(stubSprintService{listFn: func(a domain.Actor, projectID string) ([]domain.Sprint, error) {
		gotActor = a
		got = append(got, projectID)
		return []domain.Sprint{}, nil
	}})

	for _, path := range []string{"/api/v1/sprints?project_id=" + sprintProjectID, "/api/v1/sprints?project_id=nope", "/api/v1/sprints"} {
		if status, _ := send(t, app, nethttp.MethodGet, path, "", memberHeaders); status != 200 {
			t.Fatalf("%s: status=%d", path, status)
		}
	}

	if len(got) != 3 || got[0] != sprintProjectID || got[1] != "nope" || got[2] != "" || gotActor.UserID != "u1" {
		t.Fatalf("filters = %q actor = %+v", got, gotActor)
	}
}

func TestListSprints_MissingProjectIDViolationIs400WithViolations(t *testing.T) {
	app := sprintApp(stubSprintService{listFn: func(domain.Actor, string) ([]domain.Sprint, error) {
		return nil, &domain.InvalidInputError{Violations: []domain.Violation{{Field: "project_id", Message: "project_id is required"}}}
	}})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/sprints", "", memberHeaders)

	want := []any{map[string]any{"field": "project_id", "message": "project_id is required"}}
	if status != 400 || body["error"] != "invalid input" || !reflect.DeepEqual(body["violations"], want) {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestSprintRoutes_WithoutActorAre401(t *testing.T) {
	app := sprintApp(stubSprintService{})
	routes := []struct{ method, path, body string }{
		{nethttp.MethodGet, "/api/v1/sprints?project_id=" + sprintProjectID, ""},
		{nethttp.MethodGet, "/api/v1/sprints/5", ""},
		{nethttp.MethodPut, "/api/v1/sprints/5", `{"name":"S","team":"T"}`},
		{nethttp.MethodDelete, "/api/v1/sprints/5", ""},
	}
	for _, r := range routes {
		if status, body := send(t, app, r.method, r.path, r.body, nil); status != 401 || body["error"] != "unauthenticated" {
			t.Fatalf("%s %s: status=%d body=%v", r.method, r.path, status, body)
		}
	}
}

func TestSprintGetUpdateDelete_PassActorAndIDToService(t *testing.T) {
	var actors []domain.Actor
	var updated *domain.Sprint
	var deletedID string
	app := sprintApp(stubSprintService{
		getFn: func(a domain.Actor, id string) (*domain.Sprint, error) {
			actors = append(actors, a)
			return &domain.Sprint{ID: id, ProjectID: sprintProjectID, Name: "S"}, nil
		},
		updateFn: func(a domain.Actor, id string, s *domain.Sprint) (*domain.Sprint, error) {
			actors, updated = append(actors, a), s
			return &domain.Sprint{ID: id, ProjectID: sprintProjectID, Name: s.Name}, nil
		},
		deleteFn: func(a domain.Actor, id string) error { actors, deletedID = append(actors, a), id; return nil },
	})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/sprints/5", "", memberHeaders)
	if status != 200 || body["project_id"] != sprintProjectID {
		t.Fatalf("get status=%d body=%v", status, body)
	}

	status, body = send(t, app, nethttp.MethodPut, "/api/v1/sprints/5",
		`{"name":"Renamed","team":"Alpha","start_date":"2026-01-01","end_date":"2026-01-14"}`, memberHeaders)
	if status != 200 || body["project_id"] != sprintProjectID || updated == nil || updated.Name != "Renamed" ||
		updated.ProjectID != "" || updated.StartDate == nil || updated.EndDate == nil {
		t.Fatalf("update status=%d body=%v updated=%+v", status, body, updated)
	}

	if status, _ := send(t, app, nethttp.MethodDelete, "/api/v1/sprints/5", "", memberHeaders); status != 204 || deletedID != "5" {
		t.Fatalf("delete status=%d id=%q", status, deletedID)
	}
	if len(actors) != 3 || actors[0].UserID != "u1" || actors[2].Role != domain.SystemRoleTeamMember {
		t.Fatalf("actors = %+v", actors)
	}
}

func TestSprintGetUpdateDelete_MapServiceErrorsWithFixedMessages(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"not found", fmt.Errorf("sprint 5: %w", domain.ErrNotFound), 404, "not found"},
		{"forbidden", fmt.Errorf("sprint 5: only team members: %w", domain.ErrForbidden), 403, "forbidden"},
		{"internal", errors.New("pq: connection refused at 10.0.0.1"), 500, "internal server error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := sprintApp(stubSprintService{
				getFn:    func(domain.Actor, string) (*domain.Sprint, error) { return nil, tc.err },
				updateFn: func(domain.Actor, string, *domain.Sprint) (*domain.Sprint, error) { return nil, tc.err },
				deleteFn: func(domain.Actor, string) error { return tc.err },
			})
			for _, r := range []struct{ method, body string }{
				{nethttp.MethodGet, ""}, {nethttp.MethodPut, `{"name":"S","team":"T"}`}, {nethttp.MethodDelete, ""},
			} {
				status, body := send(t, app, r.method, "/api/v1/sprints/5", r.body, memberHeaders)
				if status != tc.status || body["error"] != tc.message {
					t.Fatalf("%s: status=%d body=%v, want %d %q", r.method, status, body, tc.status, tc.message)
				}
			}
		})
	}
}

func TestSprintCreateUpdate_BadBodyOrDateIs400(t *testing.T) {
	app := sprintApp(stubSprintService{})
	cases := []struct {
		name, method, path, body, field string
	}{
		{"create malformed", nethttp.MethodPost, "/api/v1/sprints", `{"name":`, ""},
		{"update malformed", nethttp.MethodPut, "/api/v1/sprints/5", `{"name":`, ""},
		{"create bad start", nethttp.MethodPost, "/api/v1/sprints", `{"start_date":"01/02/2026"}`, "start_date"},
		{"create bad end", nethttp.MethodPost, "/api/v1/sprints", `{"end_date":"x"}`, "end_date"},
		{"update bad start", nethttp.MethodPut, "/api/v1/sprints/5", `{"start_date":"x"}`, "start_date"},
		{"update bad end", nethttp.MethodPut, "/api/v1/sprints/5", `{"end_date":"x"}`, "end_date"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := send(t, app, tc.method, tc.path, tc.body, memberHeaders)
			if status != 400 {
				t.Fatalf("status=%d body=%v", status, body)
			}
			if tc.field == "" {
				if body["error"] != "invalid JSON body" {
					t.Fatalf("body=%v", body)
				}
				return
			}
			violations, _ := body["violations"].([]any)
			if len(violations) != 1 || violations[0].(map[string]any)["field"] != tc.field {
				t.Fatalf("body=%v, want violation on %s", body, tc.field)
			}
		})
	}
}

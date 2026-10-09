package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	handler "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

// stubProjectService embeds the interface so only the methods under test are
// implemented; any other call panics on the nil embedded value.
type stubProjectService struct {
	in.ProjectConfigurationService
	createFn   func(domain.Actor, in.CreateProjectInput) (*domain.Project, error)
	getFn      func(domain.Actor, string) (*domain.Project, error)
	addFn      func(a domain.Actor, projectID, email string) (*domain.Project, error)
	assignFn   func(a domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error)
	validateFn func(a domain.Actor, projectID string, attrs domain.WorkItemAttributes) (domain.ValidationResult, error)
}

func (s stubProjectService) AddProjectMember(_ context.Context, a domain.Actor, projectID, email string) (*domain.Project, error) {
	return s.addFn(a, projectID, email)
}

func (s stubProjectService) AssignProjectRole(_ context.Context, a domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error) {
	return s.assignFn(a, projectID, userID, role)
}

func (s stubProjectService) CreateProjectBoard(_ context.Context, a domain.Actor, i in.CreateProjectInput) (*domain.Project, error) {
	return s.createFn(a, i)
}

func (s stubProjectService) RetrieveProjectData(_ context.Context, a domain.Actor, id string) (*domain.Project, error) {
	return s.getFn(a, id)
}

func (s stubProjectService) ValidateWorkItemAttributes(_ context.Context, a domain.Actor, projectID string, attrs domain.WorkItemAttributes) (domain.ValidationResult, error) {
	return s.validateFn(a, projectID, attrs)
}

func projectApp(svc in.ProjectConfigurationService) *fiber.App {
	return newTestRouter(nil, handler.NewTemplateHandler(svc), handler.NewProjectHandler(svc))
}

func send(t *testing.T, app *fiber.App, method, path, body string, headers map[string]string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	return resp.StatusCode, obj
}

var sampleProject = &domain.Project{
	ID: "11111111-1111-4111-8111-111111111111", Name: "Chura", Description: "d",
	TemplateID: "se", Mode: domain.ModeSE, CreatedBy: "u1",
	GroupID:   "group-secret",
	Members:   []domain.Member{{UserID: "u1", Email: "u1@example.com", Name: "Jojo Test", Role: "product_owner"}},
	CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), UpdatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
}

func TestCreateProject_Returns201WithProjectJSON(t *testing.T) {
	var gotActor domain.Actor
	var gotInput in.CreateProjectInput
	app := projectApp(stubProjectService{createFn: func(a domain.Actor, i in.CreateProjectInput) (*domain.Project, error) {
		gotActor, gotInput = a, i
		return sampleProject, nil
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects",
		`{"name":"Chura","description":"d","template_id":"se"}`, memberHeaders)

	if status != 201 {
		t.Fatalf("status = %d, want 201 (%v)", status, body)
	}
	if wantActor := (domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}); gotActor != wantActor {
		t.Fatalf("actor = %+v, want %+v", gotActor, wantActor)
	}
	if wantInput := (in.CreateProjectInput{Name: "Chura", Description: "d", TemplateID: "se"}); gotInput != wantInput {
		t.Fatalf("input = %+v, want %+v", gotInput, wantInput)
	}
	if body["id"] != "11111111-1111-4111-8111-111111111111" || body["template_id"] != "se" || body["mode"] != "se" || body["created_by"] != "u1" {
		t.Fatalf("unexpected body: %v", body)
	}
	members := body["members"].([]any)
	wantMember := map[string]any{"user_id": "u1", "email": "u1@example.com", "name": "Jojo Test", "role": "product_owner"}
	if len(members) != 1 || !reflect.DeepEqual(members[0], wantMember) {
		t.Fatalf("members = %v, want [%v]", members, wantMember)
	}
}

func TestCreateProject_MalformedJSON_Returns400(t *testing.T) {
	app := projectApp(stubProjectService{})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects", `{"name":`, memberHeaders)

	if status != 400 {
		t.Fatalf("status = %d, want 400 (%v)", status, body)
	}
	if body["error"] != "invalid JSON body" {
		t.Fatalf("error = %v, want \"invalid JSON body\"", body["error"])
	}
}

func TestCreateProject_ErrorsMapToStatus(t *testing.T) {
	violations := []domain.Violation{{Field: "name", Message: "name is required"}}
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid input", &domain.InvalidInputError{Violations: violations}, 400},
		{"auditor forbidden", fmt.Errorf("create project: %w", domain.ErrForbidden), 403},
		{"keycloak unavailable", fmt.Errorf("create project: %w", domain.ErrUnavailable), 503},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := projectApp(stubProjectService{createFn: func(domain.Actor, in.CreateProjectInput) (*domain.Project, error) {
				return nil, tc.err
			}})

			status, _ := send(t, app, nethttp.MethodPost, "/api/v1/projects", `{}`, memberHeaders)

			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
		})
	}
}

func TestCreateProject_InvalidInput_Returns400WithViolationsArray(t *testing.T) {
	app := projectApp(stubProjectService{createFn: func(domain.Actor, in.CreateProjectInput) (*domain.Project, error) {
		return nil, &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "name", Message: "name is required"},
			{Field: "template_id", Message: "unknown template"},
		}}
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects", `{}`, memberHeaders)

	if status != 400 {
		t.Fatalf("status = %d, want 400", status)
	}
	want := []any{
		map[string]any{"field": "name", "message": "name is required"},
		map[string]any{"field": "template_id", "message": "unknown template"},
	}
	if !reflect.DeepEqual(body["violations"], want) {
		t.Fatalf("violations = %#v, want %#v", body["violations"], want)
	}
}

func TestProjectRoutes_MissingActor_Return401(t *testing.T) {
	app := projectApp(stubProjectService{})

	postStatus, _ := send(t, app, nethttp.MethodPost, "/api/v1/projects", `{}`, nil)
	getStatus, _ := send(t, app, nethttp.MethodGet, "/api/v1/projects/"+sampleProject.ID, "", nil)

	if postStatus != 401 || getStatus != 401 {
		t.Fatalf("statuses = %d, %d, want 401, 401", postStatus, getStatus)
	}
}

func TestGetProject_Returns200WithProjectJSON(t *testing.T) {
	var gotID string
	app := projectApp(stubProjectService{getFn: func(_ domain.Actor, id string) (*domain.Project, error) {
		gotID = id
		return sampleProject, nil
	}})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/projects/"+sampleProject.ID, "", memberHeaders)

	if status != 200 {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	if gotID != "11111111-1111-4111-8111-111111111111" || body["name"] != "Chura" || body["created_at"] != "2026-01-02T03:04:05Z" {
		t.Fatalf("unexpected id %q / body %v", gotID, body)
	}
}

func TestGetProject_MemberJSONShape_HasNoAddedAtAndHidesGroupID(t *testing.T) {
	app := projectApp(stubProjectService{getFn: func(domain.Actor, string) (*domain.Project, error) {
		return sampleProject, nil
	}})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/projects/"+sampleProject.ID, "", memberHeaders)

	if status != 200 {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	wantMember := map[string]any{"user_id": "u1", "email": "u1@example.com", "name": "Jojo Test", "role": "product_owner"}
	members := body["members"].([]any)
	if len(members) != 1 || !reflect.DeepEqual(members[0], wantMember) {
		t.Fatalf("members = %v, want [%v]", members, wantMember)
	}
	if _, leaked := body["GroupID"]; leaked || body["group_id"] != nil {
		t.Fatalf("group id must not be exposed: %v", body)
	}
}

func TestGetProject_ErrorsMapToStatus(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"not found", fmt.Errorf("project x: %w", domain.ErrNotFound), 404},
		{"bad id", fmt.Errorf("project id: %w", domain.ErrInvalidID), 400},
		{"keycloak unavailable", fmt.Errorf("list members: %w", domain.ErrUnavailable), 503},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := projectApp(stubProjectService{getFn: func(domain.Actor, string) (*domain.Project, error) {
				return nil, tc.err
			}})

			status, _ := send(t, app, nethttp.MethodGet, "/api/v1/projects/x", "", memberHeaders)

			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
		})
	}
}

func TestProjectErrors_UseFixedPublicMessages(t *testing.T) {
	cases := []struct {
		err     error
		status  int
		message string
	}{
		{fmt.Errorf("project 1234 (secret): %w", domain.ErrNotFound), 404, "not found"},
		{fmt.Errorf("add member: duplicate key value violates unique constraint: %w", domain.ErrConflict), 409, "conflict"},
		{fmt.Errorf("project x: only team members may modify it: %w", domain.ErrForbidden), 403, "forbidden"},
		{fmt.Errorf("unknown role %q: %w", "root", domain.ErrUnauthenticated), 401, "unauthenticated"},
		{fmt.Errorf("project id %q: %w", "<script>", domain.ErrInvalidID), 400, "invalid id"},
		{fmt.Errorf("something: %w", domain.ErrInvalidInput), 400, "invalid input"},
		{fmt.Errorf("list members: http://keycloak:8080/admin token=abc: %w", domain.ErrUnavailable), 503, "service unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.message, func(t *testing.T) {
			app := projectApp(stubProjectService{getFn: func(domain.Actor, string) (*domain.Project, error) {
				return nil, tc.err
			}})

			status, body := send(t, app, nethttp.MethodGet, "/api/v1/projects/"+sampleProject.ID, "", memberHeaders)

			if status != tc.status || body["error"] != tc.message {
				t.Fatalf("status=%d body=%v, want %d %q", status, body, tc.status, tc.message)
			}
		})
	}
}

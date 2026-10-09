package http_test

import (
	"fmt"
	nethttp "net/http"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func TestAddMember_Returns201WithProjectJSON(t *testing.T) {
	var gotActor domain.Actor
	var gotProjectID, gotUserID string
	app := projectApp(stubProjectService{addFn: func(a domain.Actor, projectID, userID string) (*domain.Project, error) {
		gotActor, gotProjectID, gotUserID = a, projectID, userID
		return sampleProject, nil
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects/"+sampleProject.ID+"/members", `{"user_id":"u2"}`, memberHeaders)

	if status != 201 {
		t.Fatalf("status = %d, want 201 (%v)", status, body)
	}
	if wantActor := (domain.Actor{UserID: "u1", Role: domain.SystemRoleTeamMember}); gotActor != wantActor {
		t.Fatalf("actor = %+v, want %+v", gotActor, wantActor)
	}
	if gotProjectID != sampleProject.ID || gotUserID != "u2" || body["id"] != sampleProject.ID {
		t.Fatalf("project %q user %q body %v", gotProjectID, gotUserID, body)
	}
}

func TestAddMember_MalformedJSON_Returns400(t *testing.T) {
	app := projectApp(stubProjectService{})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects/"+sampleProject.ID+"/members", `{"user_id":`, memberHeaders)

	if status != 400 || body["error"] != "invalid JSON body" {
		t.Fatalf("status = %d, body = %v", status, body)
	}
}

func TestAddMember_ErrorsMapToStatus(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid input", &domain.InvalidInputError{Violations: []domain.Violation{{Field: "user_id", Message: "required"}}}, 400},
		{"bad project id", fmt.Errorf("project id: %w", domain.ErrInvalidID), 400},
		{"unauthenticated", fmt.Errorf("actor: %w", domain.ErrUnauthenticated), 401},
		{"auditor forbidden", fmt.Errorf("add: %w", domain.ErrForbidden), 403},
		{"not found", fmt.Errorf("project: %w", domain.ErrNotFound), 404},
		{"duplicate", fmt.Errorf("member: %w", domain.ErrConflict), 409},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := projectApp(stubProjectService{addFn: func(domain.Actor, string, string) (*domain.Project, error) {
				return nil, tc.err
			}})

			status, _ := send(t, app, nethttp.MethodPost, "/api/v1/projects/x/members", `{"user_id":"u2"}`, memberHeaders)

			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
		})
	}
}

func TestMemberRoutes_MissingActorHeaders_Return401(t *testing.T) {
	app := projectApp(stubProjectService{})

	postStatus, _ := send(t, app, nethttp.MethodPost, "/api/v1/projects/"+sampleProject.ID+"/members", `{"user_id":"u2"}`, nil)
	putStatus, _ := send(t, app, nethttp.MethodPut, "/api/v1/projects/"+sampleProject.ID+"/members/u2/role", `{"role":"developer"}`, nil)

	if postStatus != 401 || putStatus != 401 {
		t.Fatalf("statuses = %d, %d, want 401, 401", postStatus, putStatus)
	}
}

func TestAssignRole_Returns200WithProjectJSON(t *testing.T) {
	var gotActor domain.Actor
	var gotProjectID, gotUserID string
	var gotRole domain.ProjectRole
	app := projectApp(stubProjectService{assignFn: func(a domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error) {
		gotActor, gotProjectID, gotUserID, gotRole = a, projectID, userID, role
		return sampleProject, nil
	}})

	status, body := send(t, app, nethttp.MethodPut, "/api/v1/projects/"+sampleProject.ID+"/members/u2/role", `{"role":"scrum_master"}`, memberHeaders)

	if status != 200 {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	if wantActor := (domain.Actor{UserID: "u1", Role: domain.SystemRoleTeamMember}); gotActor != wantActor {
		t.Fatalf("actor = %+v, want %+v", gotActor, wantActor)
	}
	if gotProjectID != sampleProject.ID || gotUserID != "u2" || gotRole != "scrum_master" || body["id"] != sampleProject.ID {
		t.Fatalf("project %q user %q role %q body %v", gotProjectID, gotUserID, gotRole, body)
	}
}

func TestAssignRole_MalformedJSON_Returns400(t *testing.T) {
	app := projectApp(stubProjectService{})

	status, body := send(t, app, nethttp.MethodPut, "/api/v1/projects/"+sampleProject.ID+"/members/u2/role", `{"role":`, memberHeaders)

	if status != 400 || body["error"] != "invalid JSON body" {
		t.Fatalf("status = %d, body = %v", status, body)
	}
}

func TestAssignRole_ErrorsMapToStatus(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"role not in template", &domain.InvalidInputError{Violations: []domain.Violation{{Field: "role", Message: "bad"}}}, 400},
		{"unauthenticated", fmt.Errorf("actor: %w", domain.ErrUnauthenticated), 401},
		{"auditor forbidden", fmt.Errorf("assign: %w", domain.ErrForbidden), 403},
		{"member not found", fmt.Errorf("member: %w", domain.ErrNotFound), 404},
		{"last creator", fmt.Errorf("assign: %w", domain.ErrConflict), 409},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := projectApp(stubProjectService{assignFn: func(domain.Actor, string, string, domain.ProjectRole) (*domain.Project, error) {
				return nil, tc.err
			}})

			status, _ := send(t, app, nethttp.MethodPut, "/api/v1/projects/x/members/u2/role", `{"role":"developer"}`, memberHeaders)

			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
		})
	}
}

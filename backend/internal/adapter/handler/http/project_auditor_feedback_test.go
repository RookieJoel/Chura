package http_test

import (
	"fmt"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

var auditorHeaders = map[string]string{
	"X-Test-Actor-ID": "a1", "X-Test-Actor-Email": "a1@example.com", "X-Test-Actor-Role": "auditor",
}

func TestSelfAssignAuditor_Returns200(t *testing.T) {
	want := domain.ProjectAuditorAssignment{
		ProjectID: sampleProject.ID, AuditorUserID: "a1", AuditorName: "Auditor One",
		AuditorEmail: "a1@example.com", AssignedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	var gotActor domain.Actor
	var gotProjectID string
	app := projectApp(stubProjectService{selfAssignFn: func(a domain.Actor, projectID string) (domain.ProjectAuditorAssignment, error) {
		gotActor, gotProjectID = a, projectID
		return want, nil
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects/"+sampleProject.ID+"/auditors/me", "", auditorHeaders)
	if status != 200 {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	if gotActor.Role != domain.SystemRoleAuditor || gotProjectID != sampleProject.ID {
		t.Fatalf("actor=%+v project=%q", gotActor, gotProjectID)
	}
	if body["auditor_user_id"] != "a1" || body["project_id"] != sampleProject.ID {
		t.Fatalf("body = %v", body)
	}
}

func TestListAuditorsAndFeedback_ReturnArrays(t *testing.T) {
	app := projectApp(stubProjectService{
		listAuditorsFn: func(domain.Actor, string) ([]domain.ProjectAuditorAssignment, error) {
			return []domain.ProjectAuditorAssignment{{ProjectID: sampleProject.ID, AuditorUserID: "a1"}}, nil
		},
		listFeedbackFn: func(domain.Actor, string) ([]domain.ProjectFeedback, error) {
			return []domain.ProjectFeedback{{ID: "f1", ProjectID: sampleProject.ID, Body: "note"}}, nil
		},
	})

	status, body := send(t, app, nethttp.MethodGet, "/api/v1/projects/"+sampleProject.ID+"/auditors", "", memberHeaders)
	if status != 200 || body == nil {
		// send unmarshals objects; arrays need get helper — check status only via raw path
		_ = body
	}
	status, _, arr := get(t, app, "/api/v1/projects/"+sampleProject.ID+"/auditors", memberHeaders)
	if status != 200 || len(arr) != 1 {
		t.Fatalf("auditors status=%d arr=%v", status, arr)
	}
	status, _, arr = get(t, app, "/api/v1/projects/"+sampleProject.ID+"/feedback", memberHeaders)
	if status != 200 || len(arr) != 1 {
		t.Fatalf("feedback status=%d arr=%v", status, arr)
	}
}

func TestSubmitFeedback_Returns201(t *testing.T) {
	app := projectApp(stubProjectService{submitFeedbackFn: func(_ domain.Actor, projectID, body string) (domain.ProjectFeedback, error) {
		return domain.ProjectFeedback{ID: "f1", ProjectID: projectID, Body: body}, nil
	}})
	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects/"+sampleProject.ID+"/feedback",
		`{"body":"Looks solid"}`, auditorHeaders)
	if status != 201 || body["body"] != "Looks solid" {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestSubmitFeedback_ErrorsMapToStatus(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid", &domain.InvalidInputError{Violations: []domain.Violation{{Field: "body", Message: "required"}}}, 400},
		{"forbidden", fmt.Errorf("submit: %w", domain.ErrForbidden), 403},
		{"not found", fmt.Errorf("project: %w", domain.ErrNotFound), 404},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := projectApp(stubProjectService{submitFeedbackFn: func(domain.Actor, string, string) (domain.ProjectFeedback, error) {
				return domain.ProjectFeedback{}, tc.err
			}})
			status, _ := send(t, app, nethttp.MethodPost, "/api/v1/projects/x/feedback", `{"body":"x"}`, auditorHeaders)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
		})
	}
}

func TestUpdateAndDeleteProject_Handlers(t *testing.T) {
	var updated in.UpdateProjectInput
	app := projectApp(stubProjectService{
		updateFn: func(_ domain.Actor, _ string, input in.UpdateProjectInput) (*domain.Project, error) {
			updated = input
			return sampleProject, nil
		},
		deleteFn: func(domain.Actor, string) error { return nil },
	})
	status, body := send(t, app, nethttp.MethodPatch, "/api/v1/projects/"+sampleProject.ID,
		`{"name":"New","description":"d"}`, memberHeaders)
	if status != 200 || body["id"] != sampleProject.ID || updated.Name == nil || *updated.Name != "New" {
		t.Fatalf("patch status=%d body=%v updated=%+v", status, body, updated)
	}
	status, _ = send(t, app, nethttp.MethodDelete, "/api/v1/projects/"+sampleProject.ID, "", memberHeaders)
	if status != 204 {
		t.Fatalf("delete status = %d, want 204", status)
	}
}

func TestAssignAuditor_TeamMemberAssignsByUserID(t *testing.T) {
	var gotActor domain.Actor
	var gotProjectID, gotAuditorID string
	app := projectApp(stubProjectService{assignAuditorFn: func(a domain.Actor, projectID, auditorUserID string) (domain.ProjectAuditorAssignment, error) {
		gotActor, gotProjectID, gotAuditorID = a, projectID, auditorUserID
		return domain.ProjectAuditorAssignment{ProjectID: projectID, AuditorUserID: auditorUserID}, nil
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/projects/"+sampleProject.ID+"/auditors",
		`{"auditor_user_id":"a1"}`, memberHeaders)
	if status != 200 || body["auditor_user_id"] != "a1" {
		t.Fatalf("status=%d body=%v", status, body)
	}
	if gotActor.Role != domain.SystemRoleTeamMember || gotProjectID != sampleProject.ID || gotAuditorID != "a1" {
		t.Fatalf("actor=%+v project=%q auditor=%q", gotActor, gotProjectID, gotAuditorID)
	}
}

func TestListAuditorDirectory_ReturnsIDNameEmail(t *testing.T) {
	app := projectApp(stubProjectService{auditorDirFn: func(domain.Actor) ([]domain.DirectoryUser, error) {
		return []domain.DirectoryUser{{ID: "a1", Name: "Auditor One", Email: "a1@example.com", Role: domain.SystemRoleAuditor}}, nil
	}})

	status, _, arr := get(t, app, "/api/v1/auditors", memberHeaders)
	if status != 200 || len(arr) != 1 {
		t.Fatalf("status=%d arr=%v", status, arr)
	}
	if fmt.Sprint(arr[0]) != fmt.Sprint(map[string]any{"id": "a1", "name": "Auditor One", "email": "a1@example.com"}) {
		t.Fatalf("auditor = %v", arr[0])
	}
}

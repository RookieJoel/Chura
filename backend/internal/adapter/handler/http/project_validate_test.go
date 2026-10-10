package http_test

import (
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

const validatePath = "/api/v1/projects/11111111-1111-4111-8111-111111111111/work-items/validate"

var teamHeaders = map[string]string{"X-Test-Actor-ID": "u1", "X-Test-Actor-Email": "u1@example.com", "X-Test-Actor-Role": "team_member"}

func TestValidateWorkItem_Valid_Returns200WithEmptyViolations(t *testing.T) {
	var gotActor domain.Actor
	var gotProjectID string
	var gotAttrs domain.WorkItemAttributes
	app := projectApp(stubProjectService{validateFn: func(a domain.Actor, id string, attrs domain.WorkItemAttributes) (domain.ValidationResult, error) {
		gotActor, gotProjectID, gotAttrs = a, id, attrs
		return domain.ValidationResult{Valid: true, Violations: []domain.Violation{}}, nil
	}})

	status, body := send(t, app, "POST", validatePath,
		`{"type":"user_story","title":"Board","description":"As a dev I want x so that y","story_points":5,"acceptance_criteria":["a","b"]}`, teamHeaders)

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if !reflect.DeepEqual(body, map[string]any{"valid": true, "violations": []any{}}) {
		t.Fatalf("body = %v", body)
	}
	five := int32(5)
	wantAttrs := domain.WorkItemAttributes{
		Type: domain.WorkItemTypeUserStory, Title: "Board", Description: "As a dev I want x so that y",
		StoryPoints: &five, AcceptanceCriteria: []string{"a", "b"},
	}
	if !reflect.DeepEqual(gotAttrs, wantAttrs) {
		t.Fatalf("attrs = %+v, want %+v", gotAttrs, wantAttrs)
	}
	if gotActor != (domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}) || gotProjectID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("actor/project = %+v / %s", gotActor, gotProjectID)
	}
}

func TestValidateWorkItem_InvalidAttributes_Returns200WithViolations(t *testing.T) {
	app := projectApp(stubProjectService{validateFn: func(domain.Actor, string, domain.WorkItemAttributes) (domain.ValidationResult, error) {
		return domain.ValidationResult{Violations: []domain.Violation{{Field: "title", Message: "title is required"}}}, nil
	}})

	status, body := send(t, app, "POST", validatePath, `{"type":"task","title":""}`, teamHeaders)

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	want := map[string]any{"valid": false, "violations": []any{map[string]any{"field": "title", "message": "title is required"}}}
	if !reflect.DeepEqual(body, want) {
		t.Fatalf("body = %v, want %v", body, want)
	}
}

func TestValidateWorkItem_MalformedJSON_Returns400(t *testing.T) {
	app := projectApp(stubProjectService{})

	status, body := send(t, app, "POST", validatePath, `{"type":`, teamHeaders)

	if status != 400 || body["error"] != "invalid JSON body" {
		t.Fatalf("got %d %v, want 400 invalid JSON body", status, body)
	}
}

func TestValidateWorkItem_WithoutActor_Returns401(t *testing.T) {
	app := projectApp(stubProjectService{})

	status, _ := send(t, app, "POST", validatePath, `{"type":"task","title":"x"}`, nil)

	if status != 401 {
		t.Fatalf("status = %d, want 401", status)
	}
}

func TestValidateWorkItem_ServiceErrors_MapToStatus(t *testing.T) {
	for name, tc := range map[string]struct {
		err    error
		status int
	}{
		"not found":  {fmt.Errorf("project: %w", domain.ErrNotFound), 404},
		"bad id":     {fmt.Errorf("project id: %w", domain.ErrInvalidID), 400},
		"unexpected": {fmt.Errorf("boom"), 500},
	} {
		t.Run(name, func(t *testing.T) {
			app := projectApp(stubProjectService{validateFn: func(domain.Actor, string, domain.WorkItemAttributes) (domain.ValidationResult, error) {
				return domain.ValidationResult{}, tc.err
			}})

			status, _ := send(t, app, "POST", validatePath, `{"type":"task","title":"x"}`, teamHeaders)

			if status != tc.status {
				t.Fatalf("status = %d, want %d", status, tc.status)
			}
		})
	}
}

// fasthttp rejects oversized bodies before routing (413 on a live server);
// fiber's app.Test surfaces that rejection as an error instead of a response.
func TestRequestBodyOverOneMegabyte_IsRejectedBeforeHandler(t *testing.T) {
	called := false
	app := projectApp(stubProjectService{validateFn: func(domain.Actor, string, domain.WorkItemAttributes) (domain.ValidationResult, error) {
		called = true
		return domain.ValidationResult{Valid: true, Violations: []domain.Violation{}}, nil
	}})
	body := `{"type":"task","title":"x","description":"` + strings.Repeat("a", 1024*1024) + `"}`

	req := httptest.NewRequest("POST", validatePath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range teamHeaders {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	if resp != nil {
		_ = resp.Body.Close()
	}

	if err == nil || !strings.Contains(err.Error(), "body size exceeds") || called {
		t.Fatalf("err = %v, called = %v; want body-limit rejection without reaching the service", err, called)
	}
}

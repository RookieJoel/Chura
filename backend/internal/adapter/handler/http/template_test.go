package http_test

import (
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"

	handler "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func newTestApp() *fiber.App {
	templates := handler.NewTemplateHandler(service.NewProjectConfigurationService(nil))
	return newTestRouter(nil, templates, handler.NewProjectHandler(nil))
}

func get(t *testing.T, app *fiber.App, path string, headers map[string]string) (int, map[string]any, []any) {
	t.Helper()
	req := httptest.NewRequest(nethttp.MethodGet, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	var obj map[string]any
	var arr []any
	if json.Unmarshal(body, &obj) != nil {
		_ = json.Unmarshal(body, &arr)
	}
	return resp.StatusCode, obj, arr
}

var memberHeaders = map[string]string{"X-User-ID": "u1", "X-User-Role": "team_member"}

func TestListTemplates_Returns200WithBothTemplates(t *testing.T) {
	status, _, arr := get(t, newTestApp(), "/api/v1/templates", memberHeaders)

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	if len(arr) != 2 {
		t.Fatalf("got %d templates, want 2: %v", len(arr), arr)
	}
	general := arr[0].(map[string]any)
	if general["id"] != "general" || general["mode"] != "general" || general["default_role"] != "member" || general["creator_role"] != "owner" {
		t.Fatalf("unexpected general template: %v", general)
	}
	if !reflect.DeepEqual(general["capabilities"], []any{}) {
		t.Fatalf("general capabilities = %#v, want empty array", general["capabilities"])
	}
	se := arr[1].(map[string]any)
	wantRoles := []any{"product_owner", "scrum_master", "developer"}
	if se["id"] != "se" || !reflect.DeepEqual(se["roles"], wantRoles) {
		t.Fatalf("unexpected se template: %v", se)
	}
}

func TestListTemplates_WithoutHeaders_Returns401(t *testing.T) {
	status, obj, _ := get(t, newTestApp(), "/api/v1/templates", nil)

	if status != 401 {
		t.Fatalf("status = %d, want 401", status)
	}
	if msg, _ := obj["error"].(string); msg == "" {
		t.Fatalf("want non-empty error message, got %v", obj)
	}
}

func TestGetTemplate_SE_Returns200WithSnakeCaseShape(t *testing.T) {
	status, obj, _ := get(t, newTestApp(), "/api/v1/templates/se", memberHeaders)

	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	wantTypes := []any{"user_story", "task", "bug"}
	wantCaps := []any{"user_story_guidance", "acceptance_criteria", "story_points", "sprints", "se_templates"}
	if obj["id"] != "se" || obj["mode"] != "se" || obj["default_role"] != "developer" || obj["creator_role"] != "product_owner" {
		t.Fatalf("unexpected body: %v", obj)
	}
	if !reflect.DeepEqual(obj["work_item_types"], wantTypes) || !reflect.DeepEqual(obj["capabilities"], wantCaps) {
		t.Fatalf("unexpected types/capabilities: %v", obj)
	}
}

func TestGetTemplate_UnknownID_Returns404(t *testing.T) {
	status, obj, _ := get(t, newTestApp(), "/api/v1/templates/nope", memberHeaders)

	if status != 404 {
		t.Fatalf("status = %d, want 404", status)
	}
	if msg, _ := obj["error"].(string); msg == "" {
		t.Fatalf("want error message, got %v", obj)
	}
}

func TestTemplates_InvalidHeaders_Return401(t *testing.T) {
	cases := map[string]map[string]string{
		"unknown role": {"X-User-ID": "u1", "X-User-Role": "admin"},
		"blank user":   {"X-User-ID": "  ", "X-User-Role": "auditor"},
		"missing role": {"X-User-ID": "u1"},
	}
	for name, headers := range cases {
		for _, path := range []string{"/api/v1/templates", "/api/v1/templates/se"} {
			if status, _, _ := get(t, newTestApp(), path, headers); status != 401 {
				t.Errorf("%s %s: status = %d, want 401", name, path, status)
			}
		}
	}
}

func TestAuditor_CanReadTemplates(t *testing.T) {
	headers := map[string]string{"X-User-ID": "a1", "X-User-Role": "auditor"}
	if status, _, _ := get(t, newTestApp(), "/api/v1/templates", headers); status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
}

func TestCORS_AllowsActorHeaders(t *testing.T) {
	req := httptest.NewRequest(nethttp.MethodOptions, "/api/v1/templates", nil)
	resp, err := newTestApp().Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	want := "Content-Type, Authorization, X-User-ID, X-User-Role"
	if got := resp.Header.Get("Access-Control-Allow-Headers"); got != want {
		t.Fatalf("Allow-Headers = %q, want %q", got, want)
	}
}

func TestGetTemplate_SE_ExposesValidationRules(t *testing.T) {
	_, obj, _ := get(t, newTestApp(), "/api/v1/templates/se", memberHeaders)

	want := map[string]any{
		"required_fields": []any{"title"},
		"story_points":    map[string]any{"greater_than": float64(0)},
		"type_rules": []any{map[string]any{
			"type": "user_story", "description_format": "user_story", "require_acceptance_criteria": true,
		}},
	}
	if !reflect.DeepEqual(obj["validation_rules"], want) {
		t.Fatalf("validation_rules = %#v\nwant %#v", obj["validation_rules"], want)
	}
}

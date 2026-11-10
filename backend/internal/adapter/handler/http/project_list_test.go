package http_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

type projectListService struct {
	in.ProjectConfigurationService
	listFn func(domain.Actor, in.ProjectListQuery) (in.ProjectListResult, error)
}

func (s projectListService) ListAccessibleProjects(_ context.Context, actor domain.Actor, query in.ProjectListQuery) (in.ProjectListResult, error) {
	return s.listFn(actor, query)
}

func TestListProjectsDefaultsPaginationAndReturnsPage(t *testing.T) {
	var gotActor domain.Actor
	var gotQuery in.ProjectListQuery
	app := projectApp(projectListService{listFn: func(actor domain.Actor, query in.ProjectListQuery) (in.ProjectListResult, error) {
		gotActor, gotQuery = actor, query
		return in.ProjectListResult{
			Items: []domain.ProjectSummary{{ID: "p1", Name: "Alpha", TemplateID: "se", Mode: domain.ModeSE}},
			Page:  1, PageSize: 15, Total: 1,
		}, nil
	}})

	status, body := send(t, app, http.MethodGet, "/api/v1/projects", "", memberHeaders)

	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	if gotActor != (domain.Actor{UserID: "u1", Email: "u1@example.com", Role: domain.SystemRoleTeamMember}) {
		t.Fatalf("actor = %+v, want Team Member actor", gotActor)
	}
	if gotQuery != (in.ProjectListQuery{Page: 1, PageSize: 15}) {
		t.Fatalf("query = %+v, want default page 1 and page_size 15", gotQuery)
	}
	if body["page"] != float64(1) || body["page_size"] != float64(15) || body["total"] != float64(1) {
		t.Fatalf("pagination body = %v", body)
	}
	items, ok := body["items"].([]any)
	if !ok || len(items) != 1 || items[0].(map[string]any)["name"] != "Alpha" {
		t.Fatalf("items = %v, want Alpha summary", body["items"])
	}
}

func TestListProjectsUsesRequestedPagination(t *testing.T) {
	var gotQuery in.ProjectListQuery
	app := projectApp(projectListService{listFn: func(_ domain.Actor, query in.ProjectListQuery) (in.ProjectListResult, error) {
		gotQuery = query
		return in.ProjectListResult{Items: []domain.ProjectSummary{}, Page: query.Page, PageSize: query.PageSize}, nil
	}})

	status, _ := send(t, app, http.MethodGet, "/api/v1/projects?page=2&page_size=30", "", memberHeaders)

	if status != http.StatusOK || gotQuery != (in.ProjectListQuery{Page: 2, PageSize: 30}) {
		t.Fatalf("status/query = %d / %+v, want 200 / page 2 size 30", status, gotQuery)
	}
}

func TestListProjectsRejectsInvalidPagination(t *testing.T) {
	app := projectApp(projectListService{listFn: func(domain.Actor, in.ProjectListQuery) (in.ProjectListResult, error) {
		t.Fatal("service must not be called for invalid pagination")
		return in.ProjectListResult{}, nil
	}})

	for _, path := range []string{
		"/api/v1/projects?page=0",
		"/api/v1/projects?page_size=31",
		"/api/v1/projects?page=abc",
		"/api/v1/projects?page_size=",
	} {
		t.Run(path, func(t *testing.T) {
			status, body := send(t, app, http.MethodGet, path, "", memberHeaders)
			if status != http.StatusBadRequest || body["error"] != "invalid input" {
				t.Fatalf("status/body = %d / %v, want 400 invalid input", status, body)
			}
		})
	}
}

func TestListProjectsRequiresAuthentication(t *testing.T) {
	app := projectApp(projectListService{listFn: func(domain.Actor, in.ProjectListQuery) (in.ProjectListResult, error) {
		t.Fatal("service must not be called without an actor")
		return in.ProjectListResult{}, nil
	}})

	status, body := send(t, app, http.MethodGet, "/api/v1/projects", "", nil)
	if status != http.StatusUnauthorized || body["error"] != "unauthenticated" {
		t.Fatalf("status/body = %d / %v, want 401 unauthenticated", status, body)
	}
}

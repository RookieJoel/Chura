package http_test

import (
	"context"
	"fmt"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	handler "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

const commentWorkItemID = "wi-comment-1"

type stubCommentService struct {
	in.WorkItemCommentService
	listFn   func(domain.Actor, string) ([]domain.WorkItemComment, error)
	createFn func(domain.Actor, string, string) (domain.WorkItemComment, error)
}

func (s stubCommentService) ListWorkItemComments(_ context.Context, a domain.Actor, workItemID string) ([]domain.WorkItemComment, error) {
	return s.listFn(a, workItemID)
}

func (s stubCommentService) CreateWorkItemComment(_ context.Context, a domain.Actor, workItemID, body string) (domain.WorkItemComment, error) {
	return s.createFn(a, workItemID, body)
}

func commentApp(svc in.WorkItemCommentService) *fiber.App {
	app := handler.NewBaseApp("http://localhost:3000", testInstanceName)
	api := app.Group("/api/v1", testActorMiddleware)
	handler.RegisterWorkItemCommentRoutes(api, handler.NewWorkItemCommentHandler(svc))
	return app
}

func TestListWorkItemComments_Returns200(t *testing.T) {
	want := []domain.WorkItemComment{{
		ID: "c1", WorkItemID: commentWorkItemID, AuthorUserID: "u1",
		AuthorName: "Member One", AuthorEmail: "u1@example.com",
		Body: "hello", CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}}
	var gotActor domain.Actor
	var gotID string
	app := commentApp(stubCommentService{listFn: func(a domain.Actor, workItemID string) ([]domain.WorkItemComment, error) {
		gotActor, gotID = a, workItemID
		return want, nil
	}})

	status, _, arr := get(t, app, "/api/v1/work-items/"+commentWorkItemID+"/comments", memberHeaders)
	if status != 200 || len(arr) != 1 {
		t.Fatalf("status=%d arr=%v", status, arr)
	}
	if gotActor.UserID != "u1" || gotID != commentWorkItemID {
		t.Fatalf("actor=%+v id=%q", gotActor, gotID)
	}
	item, ok := arr[0].(map[string]any)
	if !ok || item["body"] != "hello" || item["work_item_id"] != commentWorkItemID {
		t.Fatalf("item = %v", arr[0])
	}
}

func TestCreateWorkItemComment_Returns201(t *testing.T) {
	app := commentApp(stubCommentService{createFn: func(_ domain.Actor, workItemID, body string) (domain.WorkItemComment, error) {
		return domain.WorkItemComment{ID: "c1", WorkItemID: workItemID, Body: body}, nil
	}})

	status, body := send(t, app, nethttp.MethodPost, "/api/v1/work-items/"+commentWorkItemID+"/comments",
		`{"body":"Ship it"}`, memberHeaders)
	if status != 201 || body["body"] != "Ship it" || body["work_item_id"] != commentWorkItemID {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestCreateWorkItemComment_ErrorsMapToStatus(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"invalid", &domain.InvalidInputError{Violations: []domain.Violation{{Field: "body", Message: "required"}}}, 400},
		{"forbidden", fmt.Errorf("create: %w", domain.ErrForbidden), 403},
		{"not member", fmt.Errorf("project: %w", domain.ErrNotProjectMember), 403},
		{"not found", fmt.Errorf("work item: %w", domain.ErrNotFound), 404},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := commentApp(stubCommentService{createFn: func(domain.Actor, string, string) (domain.WorkItemComment, error) {
				return domain.WorkItemComment{}, tc.err
			}})
			status, _ := send(t, app, nethttp.MethodPost, "/api/v1/work-items/x/comments", `{"body":"x"}`, memberHeaders)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d", status, tc.wantStatus)
			}
		})
	}
}

func TestCreateWorkItemComment_AuditorHeadersAccepted(t *testing.T) {
	var gotActor domain.Actor
	app := commentApp(stubCommentService{createFn: func(a domain.Actor, _, body string) (domain.WorkItemComment, error) {
		gotActor = a
		return domain.WorkItemComment{ID: "c1", Body: body}, nil
	}})

	status, _ := send(t, app, nethttp.MethodPost, "/api/v1/work-items/"+commentWorkItemID+"/comments",
		`{"body":"Auditor note"}`, auditorHeaders)
	if status != 201 {
		t.Fatalf("status = %d, want 201", status)
	}
	if gotActor.Role != domain.SystemRoleAuditor || gotActor.UserID != "a1" {
		t.Fatalf("actor = %+v", gotActor)
	}
}

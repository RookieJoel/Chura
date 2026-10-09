package http_test

import (
	"context"
	nethttp "net/http"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

// panickingSprintService panics on ListSprints and answers GetSprint normally.
type panickingSprintService struct{ in.SprintService }

func (panickingSprintService) ListSprints(context.Context, domain.Actor, string) ([]domain.Sprint, error) {
	panic("boom")
}

func (panickingSprintService) GetSprint(context.Context, domain.Actor, string) (*domain.Sprint, error) {
	return &domain.Sprint{ID: "5", Status: domain.Planned}, nil
}

func TestRouter_HandlerPanicReturns500AndServerKeepsServing(t *testing.T) {
	app := sprintApp(panickingSprintService{})

	panicStatus, _ := send(t, app, nethttp.MethodGet, "/api/v1/sprints", "", memberHeaders)
	nextStatus, body := send(t, app, nethttp.MethodGet, "/api/v1/sprints/5", "", memberHeaders)

	if panicStatus != 500 {
		t.Fatalf("panicking request status = %d, want 500", panicStatus)
	}
	if nextStatus != 200 || body["id"] != "5" {
		t.Fatalf("next request status = %d body = %v, want 200 with id 5", nextStatus, body)
	}
}

package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

const sprintProjectID = "11111111-1111-4111-8111-111111111111"

var memberActor = domain.Actor{UserID: "u1", Role: domain.SystemRoleTeamMember}

func TestCreateSprint_PersistsSprintWithProjectID(t *testing.T) {
	repo := &fakeSprintRepo{}
	gw := &fakeGateway{project: &domain.Project{ID: sprintProjectID, Mode: domain.ModeSE}}
	svc := service.NewSprintService(repo, gw)

	got, err := svc.CreateSprint(context.Background(), memberActor,
		&domain.Sprint{ProjectID: sprintProjectID, Name: "Sprint 1", Team: "Alpha"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "1" || got.ProjectID != sprintProjectID || got.Name != "Sprint 1" || got.Team != "Alpha" || got.Status != domain.Planned {
		t.Fatalf("unexpected sprint: %+v", got)
	}
	if repo.createCalls != 1 || repo.sprints[0].ProjectID != sprintProjectID {
		t.Fatalf("repo did not receive project id: %+v", repo.sprints)
	}
	if gw.calls != 1 || gw.gotID != sprintProjectID || gw.gotActr != memberActor {
		t.Fatalf("gateway not called as expected: %+v", gw)
	}
}

func TestCreateSprint_AuditorIsForbiddenAndGatewayNotCalled(t *testing.T) {
	repo := &fakeSprintRepo{}
	gw := &fakeGateway{project: &domain.Project{ID: sprintProjectID}}
	svc := service.NewSprintService(repo, gw)
	auditor := domain.Actor{UserID: "a1", Role: domain.SystemRoleAuditor}

	_, err := svc.CreateSprint(context.Background(), auditor,
		&domain.Sprint{ProjectID: sprintProjectID, Name: "Sprint 1", Team: "Alpha"})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
	if gw.calls != 0 || repo.createCalls != 0 {
		t.Fatalf("gateway calls=%d repo calls=%d, want 0/0", gw.calls, repo.createCalls)
	}
}

func TestCreateSprint_RejectsUnauthenticatedActor(t *testing.T) {
	repo := &fakeSprintRepo{}
	gw := &fakeGateway{}
	svc := service.NewSprintService(repo, gw)

	_, err := svc.CreateSprint(context.Background(), domain.Actor{},
		&domain.Sprint{ProjectID: sprintProjectID, Name: "Sprint 1", Team: "Alpha"})

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
	if gw.calls != 0 || repo.createCalls != 0 {
		t.Fatalf("gateway calls=%d repo calls=%d, want 0/0", gw.calls, repo.createCalls)
	}
}

func TestCreateSprint_BlankProjectIDIsInvalidInputWithViolation(t *testing.T) {
	repo := &fakeSprintRepo{}
	gw := &fakeGateway{}
	svc := service.NewSprintService(repo, gw)

	_, err := svc.CreateSprint(context.Background(), memberActor,
		&domain.Sprint{ProjectID: "  ", Name: "Sprint 1", Team: "Alpha"})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) || len(invalid.Violations) != 1 || invalid.Violations[0].Field != "project_id" {
		t.Fatalf("want one project_id violation, got %#v", err)
	}
	if gw.calls != 0 || repo.createCalls != 0 {
		t.Fatalf("gateway calls=%d repo calls=%d, want 0/0", gw.calls, repo.createCalls)
	}
}

func TestCreateSprint_NonUUIDProjectIDIsInvalidID(t *testing.T) {
	repo := &fakeSprintRepo{}
	gw := &fakeGateway{}
	svc := service.NewSprintService(repo, gw)

	_, err := svc.CreateSprint(context.Background(), memberActor,
		&domain.Sprint{ProjectID: "not-a-uuid", Name: "Sprint 1", Team: "Alpha"})

	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("want ErrInvalidID, got %v", err)
	}
	if gw.calls != 0 || repo.createCalls != 0 {
		t.Fatalf("gateway calls=%d repo calls=%d, want 0/0", gw.calls, repo.createCalls)
	}
}

func TestCreateSprint_UnknownOrOutsiderProjectIsNotFound(t *testing.T) {
	repo := &fakeSprintRepo{}
	gw := &fakeGateway{err: fmt.Errorf("project %s: %w", sprintProjectID, domain.ErrNotFound)}
	svc := service.NewSprintService(repo, gw)

	_, err := svc.CreateSprint(context.Background(), memberActor,
		&domain.Sprint{ProjectID: sprintProjectID, Name: "Sprint 1", Team: "Alpha"})

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("repo.Create called %d times, want 0", repo.createCalls)
	}
}

func TestCreateSprint_InvalidSprintIsInvalidInputAndNotPersisted(t *testing.T) {
	jan2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	jan1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		sprint domain.Sprint
		want   []domain.Violation
	}{
		{"blank name", domain.Sprint{ProjectID: sprintProjectID, Name: " ", Team: "Alpha"},
			[]domain.Violation{{Field: "name", Message: "name is required"}}},
		{"end before start", domain.Sprint{ProjectID: sprintProjectID, Name: "S", Team: "Alpha", StartDate: &jan2, EndDate: &jan1},
			[]domain.Violation{{Field: "end_date", Message: "end_date must be greater than or equal to start_date"}}},
		{"several fields", domain.Sprint{ProjectID: sprintProjectID, Name: strings.Repeat("ก", 121), Team: " ", Status: "done"},
			[]domain.Violation{
				{Field: "name", Message: "name must be at most 120 characters"},
				{Field: "team", Message: "team is required"},
				{Field: "status", Message: "status must be one of planned, active, completed"},
			}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeSprintRepo{}
			gw := &fakeGateway{project: &domain.Project{ID: sprintProjectID}}
			svc := service.NewSprintService(repo, gw)

			_, err := svc.CreateSprint(context.Background(), memberActor, &tc.sprint)

			var invalid *domain.InvalidInputError
			if !errors.As(err, &invalid) || !reflect.DeepEqual(invalid.Violations, tc.want) {
				t.Fatalf("want violations %+v, got %v", tc.want, err)
			}
			if repo.createCalls != 0 {
				t.Fatalf("repo.Create called %d times, want 0", repo.createCalls)
			}
		})
	}
}

func TestCreateSprint_GeneralModeProjectIsAllowed(t *testing.T) {
	repo := &fakeSprintRepo{}
	gw := &fakeGateway{project: &domain.Project{ID: sprintProjectID, Mode: domain.ModeGeneral}}
	svc := service.NewSprintService(repo, gw)

	got, err := svc.CreateSprint(context.Background(), memberActor,
		&domain.Sprint{ProjectID: sprintProjectID, Name: "Sprint 1", Team: "Alpha"})

	if err != nil || got.ProjectID != sprintProjectID {
		t.Fatalf("got %+v, %v", got, err)
	}
}

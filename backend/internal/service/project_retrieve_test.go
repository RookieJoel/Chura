package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

const seededProjectID = "11111111-1111-4111-8111-111111111111"

func seededService(t *testing.T) (*fakeProjectRepo, *domain.Project, func(actor domain.Actor, id string) (*domain.Project, error)) {
	t.Helper()
	repo := newFakeProjectRepo()
	svc := newProjectService(repo)
	created, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: "se"})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return repo, created, func(actor domain.Actor, id string) (*domain.Project, error) {
		return svc.RetrieveProjectData(context.Background(), actor, id)
	}
}

func TestRetrieveProjectData_AsMember_ReturnsProject(t *testing.T) {
	_, created, retrieve := seededService(t)

	got, err := retrieve(teamMember, seededProjectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, created) {
		t.Fatalf("got %+v, want %+v", got, created)
	}
}

func TestRetrieveProjectData_AsAuditor_ReturnsProject(t *testing.T) {
	_, created, retrieve := seededService(t)

	got, err := retrieve(domain.Actor{UserID: "a1", Role: domain.SystemRoleAuditor}, seededProjectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, created) {
		t.Fatalf("got %+v, want %+v", got, created)
	}
}

func TestRetrieveProjectData_AsNonMemberTeamMember_IsNotFound(t *testing.T) {
	_, _, retrieve := seededService(t)

	_, err := retrieve(domain.Actor{UserID: "u2", Role: domain.SystemRoleTeamMember}, seededProjectID)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRetrieveProjectData_MalformedID_IsInvalidID(t *testing.T) {
	_, _, retrieve := seededService(t)

	_, err := retrieve(teamMember, "not-a-uuid")

	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("want ErrInvalidID, got %v", err)
	}
}

func TestRetrieveProjectData_UnknownProject_IsNotFound(t *testing.T) {
	_, _, retrieve := seededService(t)

	_, err := retrieve(teamMember, "22222222-2222-4222-8222-222222222222")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRetrieveProjectData_UnauthenticatedActor_IsRejected(t *testing.T) {
	_, _, retrieve := seededService(t)

	_, err := retrieve(domain.Actor{UserID: "u1", Role: "admin"}, seededProjectID)

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
}

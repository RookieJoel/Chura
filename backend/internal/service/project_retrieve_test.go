package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

const seededProjectID = "11111111-1111-4111-8111-111111111111"

type seeded struct {
	repo    *fakeProjectRepo
	dir     *fakeDirectory
	svc     *service.ProjectConfigurationService
	created *domain.Project
}

func seedProject(t *testing.T) seeded {
	t.Helper()
	repo, dir, svc := newProjectEnv()
	created, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: "se"})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return seeded{repo: repo, dir: dir, svc: svc, created: created}
}

func seededService(t *testing.T) (*fakeProjectRepo, *domain.Project, func(actor domain.Actor, id string) (*domain.Project, error)) {
	t.Helper()
	s := seedProject(t)
	return s.repo, s.created, func(actor domain.Actor, id string) (*domain.Project, error) {
		return s.svc.RetrieveProjectData(context.Background(), actor, id)
	}
}

func TestRetrieveProjectData_AsMember_ReturnsProjectWithDirectoryMembers(t *testing.T) {
	_, _, retrieve := seededService(t)

	got, err := retrieve(teamMember, seededProjectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []domain.Member{{UserID: "u1", Email: "u1@example.com", Name: "User u1", Role: "product_owner"}}
	if !reflect.DeepEqual(got.Members, want) || got.Name != "Chura" || got.GroupID != "group-1" {
		t.Fatalf("got %+v, want members %+v", got, want)
	}
}

func TestRetrieveProjectData_ReadsMembersFromDirectoryOnEveryRequest(t *testing.T) {
	s := seedProject(t)
	ctx := context.Background()
	if _, err := s.svc.RetrieveProjectData(ctx, teamMember, seededProjectID); err != nil {
		t.Fatalf("first read: %v", err)
	}

	s.dir.members["group-1"] = nil // removed in Keycloak between requests
	_, err := s.svc.RetrieveProjectData(ctx, teamMember, seededProjectID)

	if !errors.Is(err, domain.ErrNotProjectMember) {
		t.Fatalf("removed member must lose access on the next request, got %v", err)
	}
}

func TestRetrieveProjectData_AsAuditor_SeesAnyProject(t *testing.T) {
	_, _, retrieve := seededService(t)

	got, err := retrieve(domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}, seededProjectID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Members) != 1 || got.Members[0].UserID != "u1" {
		t.Fatalf("members = %+v", got.Members)
	}
}

func TestRetrieveProjectData_AsNonMemberTeamMember_IsNotProjectMember(t *testing.T) {
	_, _, retrieve := seededService(t)

	_, err := retrieve(domain.Actor{UserID: "u2", Email: "u2@example.com", Role: domain.SystemRoleTeamMember}, seededProjectID)

	if !errors.Is(err, domain.ErrNotProjectMember) || !errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotProjectMember (and ErrForbidden), got %v", err)
	}
}

func TestRetrieveProjectData_DirectoryDown_IsUnavailable(t *testing.T) {
	for _, actor := range []domain.Actor{teamMember, auditor} {
		t.Run(string(actor.Role), func(t *testing.T) {
			s := seedProject(t)
			s.dir.errs["ListMembers"] = fmt.Errorf("keycloak: %w", domain.ErrUnavailable)

			_, err := s.svc.RetrieveProjectData(context.Background(), actor, seededProjectID)

			if !errors.Is(err, domain.ErrUnavailable) {
				t.Fatalf("want ErrUnavailable, got %v", err)
			}
		})
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

	_, err := retrieve(domain.Actor{UserID: "u1", Email: "u1@example.com", Role: "admin"}, seededProjectID)

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
}

package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

var (
	sprintAuditor   = domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}
	hiddenProjectGW = &fakeGateway{err: fmt.Errorf("project %s: %w", sprintProjectID, domain.ErrNotProjectMember)}
)

func seededSprints() *fakeSprintRepo {
	return &fakeSprintRepo{sprints: []domain.Sprint{{ID: "1", ProjectID: sprintProjectID, Name: "Old", Team: "Alpha"}}}
}

func visibleGateway() *fakeGateway {
	return &fakeGateway{project: &domain.Project{ID: sprintProjectID}}
}

func TestListSprints_ChecksProjectAccessAndFiltersByProject(t *testing.T) {
	for _, actor := range []domain.Actor{memberActor, sprintAuditor} {
		repo, gw := seededSprints(), visibleGateway()
		svc := service.NewSprintService(repo, gw)

		got, err := svc.ListSprints(context.Background(), actor, " "+sprintProjectID+" ")

		if err != nil || len(got) != 1 {
			t.Fatalf("got %+v, %v", got, err)
		}
		if repo.listFilter == nil || *repo.listFilter != sprintProjectID {
			t.Fatalf("repo filter = %v, want %q", repo.listFilter, sprintProjectID)
		}
		if gw.calls != 1 || gw.gotID != sprintProjectID || gw.gotActr != actor {
			t.Fatalf("gateway not consulted as expected: %+v", gw)
		}
	}
}

func TestListSprints_MissingProjectIDIsInvalidInputOnProjectID(t *testing.T) {
	repo, gw := seededSprints(), visibleGateway()
	svc := service.NewSprintService(repo, gw)

	_, err := svc.ListSprints(context.Background(), memberActor, "  ")

	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) || len(invalid.Violations) != 1 || invalid.Violations[0].Field != "project_id" {
		t.Fatalf("want one project_id violation, got %v", err)
	}
	if gw.calls != 0 || repo.listFilter != nil {
		t.Fatal("must not reach gateway or repository")
	}
}

func TestListSprints_RejectsBadActorBadIDAndHiddenProject(t *testing.T) {
	cases := []struct {
		name      string
		actor     domain.Actor
		projectID string
		gw        *fakeGateway
		want      error
	}{
		{"unauthenticated", domain.Actor{}, sprintProjectID, visibleGateway(), domain.ErrUnauthenticated},
		{"non-uuid", memberActor, "nope", visibleGateway(), domain.ErrInvalidID},
		{"non-member", memberActor, sprintProjectID, hiddenProjectGW, domain.ErrNotProjectMember},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := seededSprints()
			_, err := service.NewSprintService(repo, tc.gw).ListSprints(context.Background(), tc.actor, tc.projectID)

			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
			if repo.listFilter != nil {
				t.Fatal("repository must not be queried")
			}
		})
	}
}

func TestGetSprint_MemberAndAuditorCanRead(t *testing.T) {
	for _, actor := range []domain.Actor{memberActor, sprintAuditor} {
		gw := visibleGateway()
		got, err := service.NewSprintService(seededSprints(), gw).GetSprint(context.Background(), actor, "1")

		if err != nil || got.Name != "Old" {
			t.Fatalf("get: %+v, %v", got, err)
		}
		if gw.gotID != sprintProjectID || gw.gotActr != actor {
			t.Fatalf("gateway got %q / %+v", gw.gotID, gw.gotActr)
		}
	}
}

func TestGetSprint_ErrorCases(t *testing.T) {
	cases := []struct {
		name  string
		actor domain.Actor
		id    string
		gw    *fakeGateway
		want  error
	}{
		{"unauthenticated", domain.Actor{UserID: "u1", Email: "u1@example.com", Role: "admin"}, "1", visibleGateway(), domain.ErrUnauthenticated},
		{"unknown sprint", memberActor, "99", visibleGateway(), domain.ErrNotFound},
		{"non-member", memberActor, "1", hiddenProjectGW, domain.ErrNotProjectMember},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := service.NewSprintService(seededSprints(), tc.gw).GetSprint(context.Background(), tc.actor, tc.id)

			if !errors.Is(err, tc.want) || got != nil {
				t.Fatalf("want %v, got %+v / %v", tc.want, got, err)
			}
		})
	}
}

func TestUpdateSprint_MemberUpdatesAndKeepsProjectID(t *testing.T) {
	repo, gw := seededSprints(), visibleGateway()

	updated, err := service.NewSprintService(repo, gw).UpdateSprint(context.Background(), memberActor, "1",
		&domain.Sprint{Name: "New", Team: "Alpha"})

	if err != nil || updated.Name != "New" || updated.ProjectID != sprintProjectID {
		t.Fatalf("update: %+v, %v", updated, err)
	}
	if gw.gotID != sprintProjectID || gw.gotActr != memberActor {
		t.Fatalf("gateway got %q / %+v", gw.gotID, gw.gotActr)
	}
}

func TestUpdateSprint_ErrorCasesDoNotPersist(t *testing.T) {
	cases := []struct {
		name  string
		actor domain.Actor
		id    string
		gw    *fakeGateway
		want  error
	}{
		{"unauthenticated", domain.Actor{}, "1", visibleGateway(), domain.ErrUnauthenticated},
		{"auditor forbidden", sprintAuditor, "1", visibleGateway(), domain.ErrForbidden},
		{"auditor unknown sprint", sprintAuditor, "99", visibleGateway(), domain.ErrNotFound},
		{"unknown sprint", memberActor, "99", visibleGateway(), domain.ErrNotFound},
		{"non-member", memberActor, "1", hiddenProjectGW, domain.ErrNotProjectMember},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := seededSprints()

			_, err := service.NewSprintService(repo, tc.gw).UpdateSprint(context.Background(), tc.actor, tc.id,
				&domain.Sprint{Name: "New", Team: "Alpha"})

			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
			if repo.updateCalls != 0 || repo.sprints[0].Name != "Old" {
				t.Fatalf("must not persist: calls=%d sprints=%+v", repo.updateCalls, repo.sprints)
			}
		})
	}
}

func TestUpdateSprint_InvalidSprintReportsViolations(t *testing.T) {
	repo := seededSprints()

	_, err := service.NewSprintService(repo, visibleGateway()).UpdateSprint(context.Background(), memberActor, "1",
		&domain.Sprint{Name: " ", Team: ""})

	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) || len(invalid.Violations) != 2 ||
		invalid.Violations[0].Field != "name" || invalid.Violations[1].Field != "team" {
		t.Fatalf("want name+team violations, got %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatal("invalid update must not persist")
	}
}

func TestDeleteSprint_MemberDeletes(t *testing.T) {
	repo := seededSprints()
	svc := service.NewSprintService(repo, visibleGateway())

	if err := svc.DeleteSprint(context.Background(), memberActor, "1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetSprint(context.Background(), memberActor, "1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("sprint still present after delete: %v", err)
	}
}

func TestDeleteSprint_ErrorCasesDoNotDelete(t *testing.T) {
	cases := []struct {
		name  string
		actor domain.Actor
		id    string
		gw    *fakeGateway
		want  error
	}{
		{"unauthenticated", domain.Actor{}, "1", visibleGateway(), domain.ErrUnauthenticated},
		{"auditor forbidden", sprintAuditor, "1", visibleGateway(), domain.ErrForbidden},
		{"auditor unknown sprint", sprintAuditor, "99", visibleGateway(), domain.ErrNotFound},
		{"unknown sprint", memberActor, "99", visibleGateway(), domain.ErrNotFound},
		{"non-member", memberActor, "1", hiddenProjectGW, domain.ErrNotProjectMember},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := seededSprints()

			err := service.NewSprintService(repo, tc.gw).DeleteSprint(context.Background(), tc.actor, tc.id)

			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
			if repo.deleteCalls != 0 || len(repo.sprints) != 1 {
				t.Fatalf("must not delete: calls=%d sprints=%+v", repo.deleteCalls, repo.sprints)
			}
		})
	}
}

func TestSprintService_WrapsRepositoryErrorsWithContext(t *testing.T) {
	storageErr := errors.New("connection reset")
	newSvc := func() *service.SprintService {
		repo := seededSprints()
		repo.opErr = storageErr
		return service.NewSprintService(repo, visibleGateway())
	}
	ctx := context.Background()
	sprint := &domain.Sprint{ProjectID: sprintProjectID, Name: "New", Team: "Alpha"}
	cases := []struct {
		name   string
		prefix string
		call   func(*service.SprintService) error
	}{
		{"create", "create sprint: ", func(s *service.SprintService) error {
			_, err := s.CreateSprint(ctx, memberActor, sprint)
			return err
		}},
		{"list", "list sprints: ", func(s *service.SprintService) error {
			_, err := s.ListSprints(ctx, memberActor, sprintProjectID)
			return err
		}},
		{"update", "update sprint: ", func(s *service.SprintService) error {
			_, err := s.UpdateSprint(ctx, memberActor, "1", sprint)
			return err
		}},
		{"delete", "delete sprint: ", func(s *service.SprintService) error {
			return s.DeleteSprint(ctx, memberActor, "1")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call(newSvc())

			if !errors.Is(err, storageErr) || !strings.HasPrefix(err.Error(), tc.prefix) {
				t.Fatalf("want %q-prefixed wrap of %v, got %v", tc.prefix, storageErr, err)
			}
		})
	}
}

func TestGetSprint_WrapsNotFoundWithContext(t *testing.T) {
	_, err := service.NewSprintService(seededSprints(), visibleGateway()).GetSprint(context.Background(), memberActor, "99")

	if !errors.Is(err, domain.ErrNotFound) || !strings.HasPrefix(err.Error(), "get sprint: ") {
		t.Fatalf("got %v", err)
	}
}

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
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

var fixedNow = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func newProjectService(repo *fakeProjectRepo, dir *fakeDirectory) *service.ProjectConfigurationService {
	return service.NewProjectConfigurationService(repo, dir,
		service.WithIDGenerator(func() string { return "11111111-1111-4111-8111-111111111111" }),
		service.WithClock(func() time.Time { return fixedNow }),
	)
}

func newProjectEnv() (*fakeProjectRepo, *fakeDirectory, *service.ProjectConfigurationService) {
	repo := newFakeProjectRepo()
	dir := newFakeDirectory(repo.log)
	return repo, dir, newProjectService(repo, dir)
}

func TestCreateProjectBoard_SETemplate_CreatorBecomesProductOwner(t *testing.T) {
	repo, _, svc := newProjectEnv()

	got, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{
		Name: "  Chura  ", Description: "  agile board \n", TemplateID: "se",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := &domain.Project{
		ID: "11111111-1111-4111-8111-111111111111", Name: "Chura", Description: "agile board",
		TemplateID: "se", Mode: "se", CreatedBy: "u1", GroupID: "group-1",
		Members:   []domain.Member{{UserID: "u1", Email: "u1@example.com", Role: "product_owner"}},
		CreatedAt: fixedNow, UpdatedAt: fixedNow,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("project mismatch\n got: %+v\nwant: %+v", got, want)
	}
	stored, err := repo.GetByID(context.Background(), want.ID)
	if err != nil || stored.GroupID != "group-1" || stored.Name != "Chura" {
		t.Fatalf("row not persisted with group id: %+v, %v", stored, err)
	}
}

func TestCreateProjectBoard_GeneralTemplate_CreatorBecomesOwner(t *testing.T) {
	_, _, svc := newProjectEnv()

	got, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{
		Name: "Thesis", TemplateID: "general",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Mode != "general" || got.TemplateID != "general" || got.Description != "" {
		t.Fatalf("unexpected project: %+v", got)
	}
	wantMembers := []domain.Member{{UserID: "u1", Email: "u1@example.com", Role: "owner"}}
	if !reflect.DeepEqual(got.Members, wantMembers) {
		t.Fatalf("members = %+v, want %+v", got.Members, wantMembers)
	}
}

func TestCreateProjectBoard_WritesInOrderAndUndoesInReverseOnFailure(t *testing.T) {
	errKeycloak := fmt.Errorf("keycloak down: %w", domain.ErrUnavailable)
	errDB := errors.New("db down")
	const (
		create   = "CreateProjectGroup(11111111-1111-4111-8111-111111111111)"
		add      = "AddMember(group-1,u1,product_owner)"
		repoCall = "repo.Create(11111111-1111-4111-8111-111111111111)"
		remove   = "RemoveMember(group-1,u1)"
		del      = "DeleteProjectGroup(group-1)"
	)
	cases := []struct {
		name      string
		dirErrs   map[string]error
		repoErr   error
		wantCalls []string
		wantErr   error // nil means success
	}{
		{"success", nil, nil, []string{create, add, repoCall}, nil},
		{"group creation fails", map[string]error{"CreateProjectGroup": errKeycloak}, nil, []string{create}, domain.ErrUnavailable},
		{"member add fails", map[string]error{"AddMember": errKeycloak}, nil, []string{create, add, del}, domain.ErrUnavailable},
		{"row insert fails", nil, errDB, []string{create, add, repoCall, remove, del}, errDB},
		{"row insert fails and undo fails", map[string]error{"RemoveMember": errKeycloak, "DeleteProjectGroup": errKeycloak}, errDB, []string{create, add, repoCall, remove, del}, errDB},
		{"member add fails and undo fails", map[string]error{"AddMember": errKeycloak, "DeleteProjectGroup": errors.New("also down")}, nil, []string{create, add, del}, domain.ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, dir, svc := newProjectEnv()
			for method, err := range tc.dirErrs {
				dir.errs[method] = err
			}
			repo.createErr = tc.repoErr

			_, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: "se"})

			if tc.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want it to match %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(repo.log.calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", repo.log.calls, tc.wantCalls)
			}
		})
	}
}

func createInvalid(t *testing.T, repo *fakeProjectRepo, input in.CreateProjectInput) []domain.Violation {
	t.Helper()
	_, err := newProjectService(repo, newFakeDirectory(repo.log)).CreateProjectBoard(context.Background(), teamMember, input)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) {
		t.Fatalf("want *InvalidInputError, got %T", err)
	}
	if len(repo.log.calls) != 0 {
		t.Fatalf("invalid input must not touch the repository or directory, calls = %v", repo.log.calls)
	}
	return invalid.Violations
}

func TestCreateProjectBoard_BlankName_ReportsNameViolation(t *testing.T) {
	got := createInvalid(t, newFakeProjectRepo(), in.CreateProjectInput{Name: "   ", TemplateID: "se"})

	want := []domain.Violation{{Field: "name", Message: "name is required"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %+v, want %+v", got, want)
	}
}

func TestCreateProjectBoard_NameOf101Runes_ReportsNameViolation(t *testing.T) {
	got := createInvalid(t, newFakeProjectRepo(), in.CreateProjectInput{Name: strings.Repeat("a", 101), TemplateID: "se"})

	want := []domain.Violation{{Field: "name", Message: "name must be at most 100 characters"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %+v, want %+v", got, want)
	}
}

func TestCreateProjectBoard_NameOf101ThaiRunes_ReportsNameViolation(t *testing.T) {
	got := createInvalid(t, newFakeProjectRepo(), in.CreateProjectInput{Name: strings.Repeat("ก", 101), TemplateID: "se"})

	want := []domain.Violation{{Field: "name", Message: "name must be at most 100 characters"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %+v, want %+v", got, want)
	}
}

func TestCreateProjectBoard_NameOf100ThaiRunes_IsAccepted(t *testing.T) {
	name := strings.Repeat("ก", 100)

	_, _, svc := newProjectEnv()

	got, err := svc.CreateProjectBoard(context.Background(), teamMember,
		in.CreateProjectInput{Name: name, TemplateID: "se"})
	if err != nil {
		t.Fatalf("100 Thai runes (300 bytes) must be accepted: %v", err)
	}
	if got.Name != name {
		t.Fatalf("name altered: %q", got.Name)
	}
}

func TestCreateProjectBoard_DescriptionOf2001Runes_ReportsDescriptionViolation(t *testing.T) {
	got := createInvalid(t, newFakeProjectRepo(), in.CreateProjectInput{
		Name: "ok", Description: strings.Repeat("ก", 2001), TemplateID: "se",
	})

	want := []domain.Violation{{Field: "description", Message: "description must be at most 2000 characters"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %+v, want %+v", got, want)
	}
}

func TestCreateProjectBoard_UnknownTemplate_ReportsTemplateViolation(t *testing.T) {
	got := createInvalid(t, newFakeProjectRepo(), in.CreateProjectInput{Name: "ok", TemplateID: "kanban"})

	want := []domain.Violation{{Field: "template_id", Message: "unknown template"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %+v, want %+v", got, want)
	}
}

func TestCreateProjectBoard_SeveralViolations_AreAllReportedInFieldOrder(t *testing.T) {
	got := createInvalid(t, newFakeProjectRepo(), in.CreateProjectInput{
		Name: " ", Description: strings.Repeat("d", 2001), TemplateID: "",
	})

	want := []domain.Violation{
		{Field: "name", Message: "name is required"},
		{Field: "description", Message: "description must be at most 2000 characters"},
		{Field: "template_id", Message: "unknown template"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %+v, want %+v", got, want)
	}
}

func TestCreateProjectBoard_Auditor_IsForbidden(t *testing.T) {
	repo := newFakeProjectRepo()
	auditor := domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}

	_, err := newProjectService(repo, newFakeDirectory(repo.log)).CreateProjectBoard(context.Background(), auditor,
		in.CreateProjectInput{Name: "ok", TemplateID: "se"})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
	if len(repo.log.calls) != 0 {
		t.Fatalf("forbidden create must not touch anything, calls = %v", repo.log.calls)
	}
}

func TestCreateProjectBoard_UnauthenticatedActor_IsRejected(t *testing.T) {
	repo := newFakeProjectRepo()

	_, err := newProjectService(repo, newFakeDirectory(repo.log)).CreateProjectBoard(context.Background(), domain.Actor{},
		in.CreateProjectInput{Name: "ok", TemplateID: "se"})

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
	if len(repo.log.calls) != 0 {
		t.Fatalf("unauthenticated create must not touch anything, calls = %v", repo.log.calls)
	}
}

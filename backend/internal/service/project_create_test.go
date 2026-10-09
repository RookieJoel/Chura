package service_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

var fixedNow = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func newProjectService(repo *fakeProjectRepo) *service.ProjectConfigurationService {
	return service.NewProjectConfigurationService(repo,
		service.WithIDGenerator(func() string { return "11111111-1111-4111-8111-111111111111" }),
		service.WithClock(func() time.Time { return fixedNow }),
	)
}

func TestCreateProjectBoard_SETemplate_CreatorBecomesProductOwner(t *testing.T) {
	repo := newFakeProjectRepo()
	svc := newProjectService(repo)

	got, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{
		Name: "  Chura  ", Description: "  agile board \n", TemplateID: "se",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := &domain.Project{
		ID: "11111111-1111-4111-8111-111111111111", Name: "Chura", Description: "agile board",
		TemplateID: "se", Mode: "se", CreatedBy: "u1",
		Members:   []domain.Member{{UserID: "u1", Role: "product_owner", AddedAt: fixedNow}},
		CreatedAt: fixedNow, UpdatedAt: fixedNow,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("project mismatch\n got: %+v\nwant: %+v", got, want)
	}
	stored, err := repo.GetByID(context.Background(), want.ID)
	if err != nil || !reflect.DeepEqual(stored, want) {
		t.Fatalf("not persisted as returned: %+v, %v", stored, err)
	}
}

func TestCreateProjectBoard_GeneralTemplate_CreatorBecomesOwner(t *testing.T) {
	svc := newProjectService(newFakeProjectRepo())

	got, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{
		Name: "Thesis", TemplateID: "general",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Mode != "general" || got.TemplateID != "general" || got.Description != "" {
		t.Fatalf("unexpected project: %+v", got)
	}
	wantMembers := []domain.Member{{UserID: "u1", Role: "owner", AddedAt: fixedNow}}
	if !reflect.DeepEqual(got.Members, wantMembers) {
		t.Fatalf("members = %+v, want %+v", got.Members, wantMembers)
	}
}

func createInvalid(t *testing.T, repo *fakeProjectRepo, input in.CreateProjectInput) []domain.Violation {
	t.Helper()
	_, err := newProjectService(repo).CreateProjectBoard(context.Background(), teamMember, input)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) {
		t.Fatalf("want *InvalidInputError, got %T", err)
	}
	if len(repo.projects) != 0 {
		t.Fatalf("invalid input must not persist anything")
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

	got, err := newProjectService(newFakeProjectRepo()).CreateProjectBoard(context.Background(), teamMember,
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
	auditor := domain.Actor{UserID: "a1", Role: domain.SystemRoleAuditor}

	_, err := newProjectService(repo).CreateProjectBoard(context.Background(), auditor,
		in.CreateProjectInput{Name: "ok", TemplateID: "se"})

	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
	if len(repo.projects) != 0 {
		t.Fatal("forbidden create must not persist")
	}
}

func TestCreateProjectBoard_UnauthenticatedActor_IsRejected(t *testing.T) {
	repo := newFakeProjectRepo()

	_, err := newProjectService(repo).CreateProjectBoard(context.Background(), domain.Actor{},
		in.CreateProjectInput{Name: "ok", TemplateID: "se"})

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
	if len(repo.projects) != 0 {
		t.Fatal("unauthenticated create must not persist")
	}
}

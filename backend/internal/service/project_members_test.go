package service_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

var auditor = domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}

// projectWithTemplate seeds a project (creator u1) from the given template.
func projectWithTemplate(t *testing.T, templateID string) (*fakeProjectRepo, *service.ProjectConfigurationService) {
	t.Helper()
	repo := newFakeProjectRepo()
	svc := newProjectService(repo)
	if _, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: templateID}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return repo, svc
}

func TestAddProjectMember_AppliesTemplateDefaultRole(t *testing.T) {
	cases := []struct {
		templateID  string
		wantDefault domain.ProjectRole
		wantCreator domain.ProjectRole
	}{
		{"se", "developer", "product_owner"},
		{"general", "member", "owner"},
	}
	for _, tc := range cases {
		t.Run(tc.templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, tc.templateID)

			got, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "  u2 ")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := []domain.Member{
				{UserID: "u1", Role: tc.wantCreator, AddedAt: fixedNow},
				{UserID: "u2", Role: tc.wantDefault, AddedAt: fixedNow},
			}
			if !reflect.DeepEqual(got.Members, want) {
				t.Fatalf("members = %+v, want %+v", got.Members, want)
			}
		})
	}
}

var bothTemplates = []string{"se", "general"}

func TestAddProjectMember_Auditor_IsForbidden(t *testing.T) {
	for _, templateID := range bothTemplates {
		t.Run(templateID, func(t *testing.T) {
			repo, svc := projectWithTemplate(t, templateID)

			_, err := svc.AddProjectMember(context.Background(), auditor, seededProjectID, "u2")

			if !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("want ErrForbidden, got %v", err)
			}
			if stored, _ := repo.GetByID(context.Background(), seededProjectID); len(stored.Members) != 1 {
				t.Fatalf("auditor must not mutate, members = %+v", stored.Members)
			}
		})
	}
}

func TestAddProjectMember_OutsiderTeamMember_IsNotFound(t *testing.T) {
	for _, templateID := range bothTemplates {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)
			outsider := domain.Actor{UserID: "u9", Email: "u9@example.com", Role: domain.SystemRoleTeamMember}

			_, err := svc.AddProjectMember(context.Background(), outsider, seededProjectID, "u2")

			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("want ErrNotFound, got %v", err)
			}
		})
	}
}

func TestAddProjectMember_AlreadyMember_IsConflict(t *testing.T) {
	for _, templateID := range bothTemplates {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)

			_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "u1")

			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("want ErrConflict, got %v", err)
			}
		})
	}
}

func TestAddProjectMember_InvalidUserID_ReportsViolationOnUserID(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"blank":       "   ",
		"101 runes":   strings.Repeat("é", 101),
		"101 trimmed": " " + strings.Repeat("a", 101) + " ",
	}
	for _, templateID := range bothTemplates {
		for name, userID := range cases {
			t.Run(templateID+"/"+name, func(t *testing.T) {
				_, svc := projectWithTemplate(t, templateID)

				_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, userID)

				var invalid *domain.InvalidInputError
				if !errors.As(err, &invalid) || len(invalid.Violations) != 1 || invalid.Violations[0].Field != "user_id" {
					t.Fatalf("want single violation on user_id, got %v", err)
				}
			})
		}
	}
}

func TestAddProjectMember_100RuneUserID_IsAccepted(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	got, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, strings.Repeat("é", 100))

	if err != nil || len(got.Members) != 2 {
		t.Fatalf("want accepted, got %v / %+v", err, got)
	}
}

func TestAddProjectMember_UnauthenticatedActor_IsRejected(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	_, err := svc.AddProjectMember(context.Background(), domain.Actor{UserID: "u1", Email: "u1@example.com", Role: "admin"}, seededProjectID, "u2")

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
}

func TestAddProjectMember_MalformedProjectID_IsInvalidID(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	_, err := svc.AddProjectMember(context.Background(), teamMember, "not-a-uuid", "u2")

	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("want ErrInvalidID, got %v", err)
	}
}

func TestAddProjectMember_UnknownProject_IsNotFound(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	_, err := svc.AddProjectMember(context.Background(), teamMember, "22222222-2222-4222-8222-222222222222", "u2")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

// templateRoles holds literal expectations per template from the spec.
var templateRoles = map[string]struct {
	creator, other, foreign domain.ProjectRole
}{
	"se":      {creator: "product_owner", other: "scrum_master", foreign: "member"},
	"general": {creator: "owner", other: "member", foreign: "scrum_master"},
}

func TestAssignProjectRole_ChangesMemberRole(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)
			ctx := context.Background()
			if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2"); err != nil {
				t.Fatalf("seed: %v", err)
			}

			got, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", roles.creator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			m, _ := got.MemberByID("u2")
			if m.Role != roles.creator {
				t.Fatalf("u2 role = %q, want %q", m.Role, roles.creator)
			}
		})
	}
}

func TestAssignProjectRole_RoleNotInTemplate_ReportsViolationOnRole(t *testing.T) {
	for templateID, roles := range templateRoles {
		for name, role := range map[string]domain.ProjectRole{"other template's role": roles.foreign, "unknown": "wizard", "empty": ""} {
			t.Run(templateID+"/"+name, func(t *testing.T) {
				_, svc := projectWithTemplate(t, templateID)

				_, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "u1", role)

				var invalid *domain.InvalidInputError
				if !errors.As(err, &invalid) || len(invalid.Violations) != 1 || invalid.Violations[0].Field != "role" {
					t.Fatalf("want single violation on role, got %v", err)
				}
			})
		}
	}
}

func TestAssignProjectRole_TargetNotAMember_IsNotFound(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			repo, svc := projectWithTemplate(t, templateID)

			_, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "ghost", roles.other)

			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("want ErrNotFound, got %v", err)
			}
			if stored, _ := repo.GetByID(context.Background(), seededProjectID); len(stored.Members) != 1 {
				t.Fatalf("must not add members, got %+v", stored.Members)
			}
		})
	}
}

func TestAssignProjectRole_Auditor_IsForbidden(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			repo, svc := projectWithTemplate(t, templateID)

			_, err := svc.AssignProjectRole(context.Background(), auditor, seededProjectID, "u1", roles.other)

			if !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("want ErrForbidden, got %v", err)
			}
			if stored, _ := repo.GetByID(context.Background(), seededProjectID); stored.Members[0].Role != roles.creator {
				t.Fatalf("auditor must not mutate, got %+v", stored.Members)
			}
		})
	}
}

func TestAssignProjectRole_OutsiderTeamMember_IsNotFound(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)
			outsider := domain.Actor{UserID: "u9", Email: "u9@example.com", Role: domain.SystemRoleTeamMember}

			_, err := svc.AssignProjectRole(context.Background(), outsider, seededProjectID, "u1", roles.other)

			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("want ErrNotFound, got %v", err)
			}
		})
	}
}

func TestAssignProjectRole_DemotingSoleCreatorRoleHolder_IsConflict(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			repo, svc := projectWithTemplate(t, templateID)

			_, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "u1", roles.other)

			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("want ErrConflict, got %v", err)
			}
			if stored, _ := repo.GetByID(context.Background(), seededProjectID); stored.Members[0].Role != roles.creator {
				t.Fatalf("role must be unchanged, got %+v", stored.Members)
			}
		})
	}
}

func TestAssignProjectRole_DemotingCreatorRoleWhenAnotherHolderExists_Succeeds(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)
			ctx := context.Background()
			if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2"); err != nil {
				t.Fatalf("seed: %v", err)
			}
			if _, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", roles.creator); err != nil {
				t.Fatalf("seed promote: %v", err)
			}

			got, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u1", roles.other)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			u1, _ := got.MemberByID("u1")
			if u1.Role != roles.other || got.CountRole(roles.creator) != 1 {
				t.Fatalf("members = %+v", got.Members)
			}
		})
	}
}

func TestAssignProjectRole_SameRole_IsNoOpSuccess(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)

			got, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "u1", roles.creator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			u1, _ := got.MemberByID("u1")
			if u1.Role != roles.creator {
				t.Fatalf("role = %q, want %q", u1.Role, roles.creator)
			}
		})
	}
}

func TestAssignProjectRole_UnauthenticatedAndBadID(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	_, unauthErr := svc.AssignProjectRole(context.Background(), domain.Actor{UserID: "u1", Email: "u1@example.com", Role: "admin"}, seededProjectID, "u1", "developer")
	_, idErr := svc.AssignProjectRole(context.Background(), teamMember, "not-a-uuid", "u1", "developer")

	if !errors.Is(unauthErr, domain.ErrUnauthenticated) || !errors.Is(idErr, domain.ErrInvalidID) {
		t.Fatalf("got %v, %v", unauthErr, idErr)
	}
}

func TestAssignProjectRole_TrimsUserID(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")
	ctx := context.Background()
	if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "  u2 ", "scrum_master")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u2, _ := got.MemberByID("u2"); u2.Role != "scrum_master" {
		t.Fatalf("members = %+v", got.Members)
	}
}

// A concurrent demotion of the other creator-role holder lands after the
// service read the project; the locked re-check must still refuse.
func TestAssignProjectRole_LastCreatorGuardIsRecheckedAtUpdateTime(t *testing.T) {
	repo, svc := projectWithTemplate(t, "se")
	ctx := context.Background()
	if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", "product_owner"); err != nil {
		t.Fatalf("seed promote: %v", err)
	}
	repo.beforeUpdate = func(r *fakeProjectRepo) {
		p := r.projects[seededProjectID]
		p.Members[1].Role = "developer"
		r.projects[seededProjectID] = p
	}

	_, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u1", "developer")

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
	if u1, _ := repo.projects[seededProjectID].MemberByID("u1"); u1.Role != "product_owner" {
		t.Fatalf("u1 must keep product_owner, got %+v", repo.projects[seededProjectID].Members)
	}
}

func TestProjectOperations_UnknownStoredTemplate_IsInternalError(t *testing.T) {
	repo := newFakeProjectRepo()
	repo.projects[seededProjectID] = domain.Project{
		ID: seededProjectID, TemplateID: "vanished",
		Members: []domain.Member{{UserID: "u1", Role: "owner"}},
	}
	svc := newProjectService(repo)
	ctx := context.Background()

	_, addErr := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2")
	_, assignErr := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u1", "owner")
	_, validateErr := svc.ValidateWorkItemAttributes(ctx, teamMember, seededProjectID, domain.WorkItemAttributes{})

	for name, err := range map[string]error{"add": addErr, "assign": assignErr, "validate": validateErr} {
		if err == nil || !strings.Contains(err.Error(), "vanished") {
			t.Fatalf("%s: want internal error naming the template, got %v", name, err)
		}
		for _, sentinel := range []error{domain.ErrNotFound, domain.ErrInvalidInput, domain.ErrConflict, domain.ErrForbidden} {
			if errors.Is(err, sentinel) {
				t.Fatalf("%s: must not map to a client error, got %v", name, err)
			}
		}
	}
}

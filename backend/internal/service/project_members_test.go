package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

var auditor = domain.Actor{UserID: "a1", Email: "a1@example.com", Role: domain.SystemRoleAuditor}

// projectWithTemplate seeds a project (creator u1) from the given template.
func projectWithTemplate(t *testing.T, templateID string) (*fakeDirectory, *service.ProjectConfigurationService) {
	t.Helper()
	_, dir, svc := newProjectEnv()
	if _, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: templateID}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return dir, svc
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

			got, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "  u2@example.com ")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want := []domain.Member{
				{UserID: "u1", Email: "u1@example.com", Name: "User u1", Role: tc.wantCreator},
				{UserID: "u2", Email: "u2@example.com", Name: "User u2", Role: tc.wantDefault},
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
			dir, svc := projectWithTemplate(t, templateID)

			_, err := svc.AddProjectMember(context.Background(), auditor, seededProjectID, "u2@example.com")

			if !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("want ErrForbidden, got %v", err)
			}
			if members := dir.members["group-1"]; len(members) != 1 {
				t.Fatalf("auditor must not mutate, members = %+v", members)
			}
		})
	}
}

func TestAddProjectMember_OutsiderTeamMember_IsNotProjectMember(t *testing.T) {
	for _, templateID := range bothTemplates {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)
			outsider := domain.Actor{UserID: "u9", Email: "u9@example.com", Role: domain.SystemRoleTeamMember}

			_, err := svc.AddProjectMember(context.Background(), outsider, seededProjectID, "u2@example.com")

			if !errors.Is(err, domain.ErrNotProjectMember) || !errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("want ErrNotProjectMember (and ErrForbidden), got %v", err)
			}
		})
	}
}

func TestAddProjectMember_AlreadyMember_IsConflict(t *testing.T) {
	for _, templateID := range bothTemplates {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)

			_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "u1@example.com")

			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("want ErrConflict, got %v", err)
			}
		})
	}
}

func TestAddProjectMember_InvalidEmail_ReportsViolationOnEmail(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"blank":       "   ",
		"no at sign":  "pim.example.com",
		"255 runes":   strings.Repeat("é", 243) + "@example.com",
		"255 trimmed": " " + strings.Repeat("a", 243) + "@example.com ",
	}
	for _, templateID := range bothTemplates {
		for name, email := range cases {
			t.Run(templateID+"/"+name, func(t *testing.T) {
				_, svc := projectWithTemplate(t, templateID)

				_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, email)

				var invalid *domain.InvalidInputError
				if !errors.As(err, &invalid) || len(invalid.Violations) != 1 || invalid.Violations[0].Field != "email" {
					t.Fatalf("want single violation on email, got %v", err)
				}
			})
		}
	}
}

func TestAddProjectMember_254RuneEmail_IsLookedUp(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")
	email := strings.Repeat("a", 242) + "@example.com"

	_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, email)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound from lookup, got %v", err)
	}
}

func TestAddProjectMember_UnknownEmail_IsNotFound(t *testing.T) {
	for _, templateID := range bothTemplates {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)

			_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "ghost@example.com")

			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("want ErrNotFound, got %v", err)
			}
		})
	}
}

func TestAddProjectMember_UserWithoutTeamMemberRole_ReportsViolationOnEmail(t *testing.T) {
	cases := map[string]domain.DirectoryUser{
		"auditor":   {ID: "a2", Email: "a2@example.com", Role: domain.SystemRoleAuditor},
		"role-less": {ID: "n1", Email: "n1@example.com", Role: ""},
	}
	for name, user := range cases {
		t.Run(name, func(t *testing.T) {
			dir, svc := projectWithTemplate(t, "se")
			dir.users[user.Email] = user

			_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, user.Email)

			var invalid *domain.InvalidInputError
			if !errors.As(err, &invalid) || !reflect.DeepEqual(invalid.Violations, []domain.Violation{{Field: "email", Message: "user is not a Team Member"}}) {
				t.Fatalf("want email violation 'user is not a Team Member', got %v", err)
			}
			if members := dir.members["group-1"]; len(members) != 1 {
				t.Fatalf("must not add members, got %+v", members)
			}
		})
	}
}

func TestAddProjectMember_AddsByDirectoryUserID_NotEmail(t *testing.T) {
	repo, dir, svc := newProjectEnv()
	dir.users["pim@example.com"] = domain.DirectoryUser{ID: "kc-pim-id", Email: "pim@example.com", Role: domain.SystemRoleTeamMember}
	if _, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: "se"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "pim@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slices.Contains(repo.log.calls, "AddMember(group-1,kc-pim-id,developer)") {
		t.Fatalf("calls = %v", repo.log.calls)
	}
}

func TestMembershipWrites_TakeTheProjectLock(t *testing.T) {
	repo, _, svc := newProjectEnv()
	ctx := context.Background()
	if _, err := svc.CreateProjectBoard(ctx, teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: "se"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if repo.locks != 1 {
		t.Fatalf("locks after add = %d, want 1", repo.locks)
	}
	if _, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", "scrum_master"); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if repo.locks != 2 {
		t.Fatalf("locks after assign = %d, want 2", repo.locks)
	}
}

func TestMembershipWrites_RejectedBeforeLock_DoNotTakeIt(t *testing.T) {
	repo, _, svc := newProjectEnv()
	ctx := context.Background()
	if _, err := svc.CreateProjectBoard(ctx, teamMember, in.CreateProjectInput{Name: "Chura", TemplateID: "se"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, _ = svc.AddProjectMember(ctx, auditor, seededProjectID, "u2@example.com")
	_, _ = svc.AddProjectMember(ctx, teamMember, seededProjectID, "nobody")
	_, _ = svc.AssignProjectRole(ctx, auditor, seededProjectID, "u1", "developer")

	if repo.locks != 0 {
		t.Fatalf("locks = %d, want 0", repo.locks)
	}
}

// Two owners each demote the other concurrently. Without the lock both pass
// the last-owner guard against the same member list; with it exactly one wins.
func TestAssignProjectRole_ConcurrentDemotionOfLastTwoOwners_ExactlyOneSucceeds(t *testing.T) {
	for range 50 {
		_, svc := projectWithTemplate(t, "general")
		ctx := context.Background()
		if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
			t.Fatalf("seed add: %v", err)
		}
		if _, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", "owner"); err != nil {
			t.Fatalf("seed promote: %v", err)
		}
		u2Actor := domain.Actor{UserID: "u2", Email: "u2@example.com", Role: domain.SystemRoleTeamMember}

		errs := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", "member")
			errs <- err
		}()
		go func() {
			defer wg.Done()
			_, err := svc.AssignProjectRole(ctx, u2Actor, seededProjectID, "u1", "member")
			errs <- err
		}()
		wg.Wait()
		close(errs)

		var succeeded, conflicts int
		for err := range errs {
			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, domain.ErrConflict):
				conflicts++
			default:
				t.Fatalf("unexpected error: %v", err)
			}
		}
		if succeeded != 1 || conflicts != 1 {
			t.Fatalf("succeeded = %d, conflicts = %d, want 1 and 1", succeeded, conflicts)
		}
	}
}

func TestAddProjectMember_ConcurrentSameEmail_ExactlyOneSucceeds(t *testing.T) {
	for range 50 {
		dir, svc := projectWithTemplate(t, "se")
		ctx := context.Background()

		errs := make(chan error, 2)
		var wg sync.WaitGroup
		wg.Add(2)
		for range 2 {
			go func() {
				defer wg.Done()
				_, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com")
				errs <- err
			}()
		}
		wg.Wait()
		close(errs)

		var succeeded, conflicts int
		for err := range errs {
			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, domain.ErrConflict):
				conflicts++
			default:
				t.Fatalf("unexpected error: %v", err)
			}
		}
		if succeeded != 1 || conflicts != 1 || len(dir.members["group-1"]) != 2 {
			t.Fatalf("succeeded = %d, conflicts = %d, members = %+v", succeeded, conflicts, dir.members["group-1"])
		}
	}
}

func TestAddProjectMember_UnauthenticatedActor_IsRejected(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	_, err := svc.AddProjectMember(context.Background(), domain.Actor{UserID: "u1", Email: "u1@example.com", Role: "admin"}, seededProjectID, "u2@example.com")

	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("want ErrUnauthenticated, got %v", err)
	}
}

func TestAddProjectMember_MalformedProjectID_IsInvalidID(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	_, err := svc.AddProjectMember(context.Background(), teamMember, "not-a-uuid", "u2@example.com")

	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("want ErrInvalidID, got %v", err)
	}
}

func TestAddProjectMember_UnknownProject_IsNotFound(t *testing.T) {
	_, svc := projectWithTemplate(t, "se")

	_, err := svc.AddProjectMember(context.Background(), teamMember, "22222222-2222-4222-8222-222222222222", "u2@example.com")

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
			if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
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
			dir, svc := projectWithTemplate(t, templateID)

			_, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "ghost", roles.other)

			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("want ErrNotFound, got %v", err)
			}
			if members := dir.members["group-1"]; len(members) != 1 {
				t.Fatalf("must not add members, got %+v", members)
			}
		})
	}
}

func TestAssignProjectRole_Auditor_IsForbidden(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			dir, svc := projectWithTemplate(t, templateID)

			_, err := svc.AssignProjectRole(context.Background(), auditor, seededProjectID, "u1", roles.other)

			if !errors.Is(err, domain.ErrForbidden) {
				t.Fatalf("want ErrForbidden, got %v", err)
			}
			if members := dir.members["group-1"]; members[0].Role != roles.creator {
				t.Fatalf("auditor must not mutate, got %+v", members)
			}
		})
	}
}

func TestAssignProjectRole_OutsiderTeamMember_IsNotProjectMember(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)
			outsider := domain.Actor{UserID: "u9", Email: "u9@example.com", Role: domain.SystemRoleTeamMember}

			_, err := svc.AssignProjectRole(context.Background(), outsider, seededProjectID, "u1", roles.other)

			if !errors.Is(err, domain.ErrNotProjectMember) || !errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("want ErrNotProjectMember (and ErrForbidden), got %v", err)
			}
		})
	}
}

func TestAssignProjectRole_DemotingSoleCreatorRoleHolder_IsConflict(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			dir, svc := projectWithTemplate(t, templateID)

			_, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "u1", roles.other)

			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("want ErrConflict, got %v", err)
			}
			if members := dir.members["group-1"]; members[0].Role != roles.creator {
				t.Fatalf("role must be unchanged, got %+v", members)
			}
		})
	}
}

func TestAssignProjectRole_DemotingCreatorRoleWhenAnotherHolderExists_Succeeds(t *testing.T) {
	for templateID, roles := range templateRoles {
		t.Run(templateID, func(t *testing.T) {
			_, svc := projectWithTemplate(t, templateID)
			ctx := context.Background()
			if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
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
	if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
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

func TestMembershipWrites_DirectoryDown_IsUnavailable(t *testing.T) {
	errKeycloak := fmt.Errorf("keycloak down: %w", domain.ErrUnavailable)
	cases := []struct {
		name   string
		method string
		call   func(svc *service.ProjectConfigurationService) error
	}{
		{"add: lookup fails", "FindUserByEmail", func(svc *service.ProjectConfigurationService) error {
			_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "u2@example.com")
			return err
		}},
		{"add: list fails", "ListMembers", func(svc *service.ProjectConfigurationService) error {
			_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "u2@example.com")
			return err
		}},
		{"add: write fails", "AddMember", func(svc *service.ProjectConfigurationService) error {
			_, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "u2@example.com")
			return err
		}},
		{"assign: list fails", "ListMembers", func(svc *service.ProjectConfigurationService) error {
			_, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "u1", "product_owner")
			return err
		}},
		{"assign: write fails", "SetMemberRole", func(svc *service.ProjectConfigurationService) error {
			_, err := svc.AssignProjectRole(context.Background(), teamMember, seededProjectID, "u1", "product_owner")
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir, svc := projectWithTemplate(t, "se")
			dir.errs[tc.method] = errKeycloak

			err := tc.call(svc)

			if !errors.Is(err, domain.ErrUnavailable) {
				t.Fatalf("want ErrUnavailable, got %v", err)
			}
		})
	}
}

func TestProjectOperations_UnknownStoredTemplate_IsInternalError(t *testing.T) {
	repo := newFakeProjectRepo()
	dir := newFakeDirectory(repo.log)
	repo.projects[seededProjectID] = domain.Project{ID: seededProjectID, TemplateID: "vanished", GroupID: "g"}
	dir.members["g"] = []domain.Member{{UserID: "u1", Role: "owner"}}
	svc := newProjectService(repo, dir)
	ctx := context.Background()

	_, addErr := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com")
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

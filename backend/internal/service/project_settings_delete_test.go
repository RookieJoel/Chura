package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func projectEnvWithCleanup(t *testing.T) (*fakeProjectRepo, *fakeDirectory, *fakeSprintStore, *fakeWorkItemStore, *service.ProjectConfigurationService) {
	t.Helper()
	repo := newFakeProjectRepo()
	dir := newFakeDirectory(repo.log)
	sprints := &fakeSprintStore{counts: map[string]int{}, log: repo.log}
	workItems := &fakeWorkItemStore{items: map[string][]domain.WorkItem{}, log: repo.log}
	svc := service.NewProjectConfigurationService(repo, dir,
		service.WithIDGenerator(func() string { return seededProjectID }),
		service.WithClock(func() time.Time { return fixedNow }),
		service.WithSprintStore(sprints),
		service.WithWorkItemStore(workItems),
	)
	if _, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{
		Name: "Chura", Description: "desc", TemplateID: "se",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return repo, dir, sprints, workItems, svc
}

func TestUpdateProject_OwnerCanEditNameAndDescription(t *testing.T) {
	_, _, _, _, svc := projectEnvWithCleanup(t)
	name := " Renamed "
	description := " new desc "
	got, err := svc.UpdateProject(context.Background(), teamMember, seededProjectID, in.UpdateProjectInput{
		Name: &name, Description: &description,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Name != "Renamed" || got.Description != "new desc" {
		t.Fatalf("got %+v", got)
	}
}

func TestUpdateProject_TemplateChangeRejectedWhenProjectHasSprintsOrWorkItems(t *testing.T) {
	_, _, sprints, workItems, svc := projectEnvWithCleanup(t)
	templateID := "general"
	ctx := context.Background()

	sprints.counts[seededProjectID] = 2
	_, err := svc.UpdateProject(ctx, teamMember, seededProjectID, in.UpdateProjectInput{TemplateID: &templateID})
	assertTemplateViolation(t, err, "This Project has 2 Sprints. Delete them before changing the template.")

	sprints.counts[seededProjectID] = 0
	workItems.items[seededProjectID] = []domain.WorkItem{{ID: "wi-1", ProjectID: seededProjectID}}
	_, err = svc.UpdateProject(ctx, teamMember, seededProjectID, in.UpdateProjectInput{TemplateID: &templateID})
	assertTemplateViolation(t, err, "This Project has 1 Work Item. Delete it before changing the template.")
}

func assertTemplateViolation(t *testing.T, err error, want string) {
	t.Helper()
	var invalid *domain.InvalidInputError
	if !errors.As(err, &invalid) || len(invalid.Violations) != 1 {
		t.Fatalf("err = %v, want one template_id violation", err)
	}
	if v := invalid.Violations[0]; v.Field != "template_id" || v.Message != want {
		t.Fatalf("violation = %+v, want template_id %q", v, want)
	}
}

func TestUpdateProject_SecondOwnerCanChangeTemplateAfterCreatorLeaves(t *testing.T) {
	_, _, _, _, svc := projectEnvWithCleanup(t)
	ctx := context.Background()
	templateID := "general"
	if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
		t.Fatalf("add member: %v", err)
	}
	if _, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", "product_owner"); err != nil {
		t.Fatalf("promote: %v", err)
	}
	if _, err := svc.RemoveProjectMember(ctx, teamMember, seededProjectID, "u1"); err != nil {
		t.Fatalf("creator leaves: %v", err)
	}
	secondOwner := domain.Actor{UserID: "u2", Email: "u2@example.com", Role: domain.SystemRoleTeamMember}
	got, err := svc.UpdateProject(ctx, secondOwner, seededProjectID, in.UpdateProjectInput{TemplateID: &templateID})
	if err != nil {
		t.Fatalf("template change after creator left: %v", err)
	}
	if m, _ := got.MemberByID("u2"); m.Role != "owner" {
		t.Fatalf("second owner role = %q, want owner", m.Role)
	}
}

func TestUpdateProject_TemplateChangeSucceedsWhenEligible(t *testing.T) {
	_, dir, _, _, svc := projectEnvWithCleanup(t)
	templateID := "general"
	got, err := svc.UpdateProject(context.Background(), teamMember, seededProjectID, in.UpdateProjectInput{TemplateID: &templateID})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.TemplateID != "general" || got.Mode != domain.ModeGeneral {
		t.Fatalf("got %+v", got)
	}
	member, ok := got.MemberByID("u1")
	if !ok || member.Role != "owner" {
		t.Fatalf("creator role = %+v", member)
	}
	_ = dir
}

func TestUpdateProject_TemplateChangeSucceedsWithMembersWithoutRoles(t *testing.T) {
	_, _, _, _, svc := projectEnvWithCleanup(t)
	ctx := context.Background()
	templateID := "general"
	if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
		t.Fatalf("add member: %v", err)
	}
	got, err := svc.UpdateProject(ctx, teamMember, seededProjectID, in.UpdateProjectInput{TemplateID: &templateID})
	if err != nil {
		t.Fatalf("a member without a role must not block the template change: %v", err)
	}
	if got.TemplateID != "general" || got.Mode != domain.ModeGeneral {
		t.Fatalf("got %+v", got)
	}
}

func TestUpdateProject_TemplateChangeClearsEveryRoleButTheOwners(t *testing.T) {
	repo, _, _, _, svc := projectEnvWithCleanup(t)
	ctx := context.Background()
	templateID := "general"
	if _, err := svc.AddProjectMember(ctx, teamMember, seededProjectID, "u2@example.com"); err != nil {
		t.Fatalf("add member: %v", err)
	}
	if _, err := svc.AssignProjectRole(ctx, teamMember, seededProjectID, "u2", "scrum_master"); err != nil {
		t.Fatalf("assign role: %v", err)
	}
	got, err := svc.UpdateProject(ctx, teamMember, seededProjectID, in.UpdateProjectInput{TemplateID: &templateID})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if m, _ := got.MemberByID("u1"); m.Role != "owner" {
		t.Fatalf("creator role = %q, want owner", m.Role)
	}
	if m, _ := got.MemberByID("u2"); m.Role != "" {
		t.Fatalf("other member role = %q, want it cleared", m.Role)
	}
	if !containsCall(repo.log.calls, "repo.WithMembershipLock("+seededProjectID+")") {
		t.Fatalf("template change must run under the membership lock; calls = %v", repo.log.calls)
	}
}

func TestUpdateProject_TemplateChangeRollsBackRolesWhenUpdateFails(t *testing.T) {
	repo, _, _, _, svc := projectEnvWithCleanup(t)
	ctx := context.Background()
	templateID := "general"
	repo.updateErr = domain.ErrUnavailable
	if _, err := svc.UpdateProject(ctx, teamMember, seededProjectID, in.UpdateProjectInput{TemplateID: &templateID}); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("err = %v, want unavailable", err)
	}
	repo.updateErr = nil

	// The creator must still hold the stored template's creator role, so the
	// owner can retry instead of being locked out.
	name := "Retry"
	got, err := svc.UpdateProject(ctx, teamMember, seededProjectID, in.UpdateProjectInput{Name: &name})
	if err != nil {
		t.Fatalf("owner locked out after failed template change: %v", err)
	}
	if m, _ := got.MemberByID("u1"); m.Role != "product_owner" || got.TemplateID != "se" {
		t.Fatalf("got template %q role %q, want se/product_owner", got.TemplateID, m.Role)
	}
}

func TestUpdateProject_AuditorForbidden(t *testing.T) {
	_, _, _, _, svc := projectEnvWithCleanup(t)
	name := "Nope"
	_, err := svc.UpdateProject(context.Background(), auditor, seededProjectID, in.UpdateProjectInput{Name: &name})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden", err)
	}
}

func TestDeleteProject_OwnerCascadesAndRemovesGroup(t *testing.T) {
	repo, _, sprints, workItems, svc := projectEnvWithCleanup(t)
	workItems.items[seededProjectID] = []domain.WorkItem{{ID: "wi-1", ProjectID: seededProjectID}}
	sprints.counts[seededProjectID] = 2
	repo.auditors[seededProjectID] = []domain.ProjectAuditorAssignment{{ProjectID: seededProjectID, AuditorUserID: "a1"}}
	repo.feedback[seededProjectID] = []domain.ProjectFeedback{{ID: "f1", ProjectID: seededProjectID, Body: "note"}}

	if err := svc.DeleteProject(context.Background(), teamMember, seededProjectID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.GetByID(context.Background(), seededProjectID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("project still present: %v", err)
	}
	if sprints.counts[seededProjectID] != 0 {
		t.Fatalf("sprints not cleared")
	}
	if len(workItems.items[seededProjectID]) != 0 {
		t.Fatalf("work items not cleared")
	}
	if !containsCall(repo.log.calls, "DeleteProjectGroup(group-1)") {
		t.Fatalf("calls = %v, want DeleteProjectGroup", repo.log.calls)
	}
}

func TestDeleteProject_KeycloakFailureSurfacesWithoutSuccess(t *testing.T) {
	repo, dir, _, _, svc := projectEnvWithCleanup(t)
	dir.errs["DeleteProjectGroup"] = domain.ErrUnavailable
	err := svc.DeleteProject(context.Background(), teamMember, seededProjectID)
	if !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("err = %v, want unavailable", err)
	}
	if _, getErr := repo.GetByID(context.Background(), seededProjectID); getErr != nil {
		t.Fatalf("project should remain after group failure: %v", getErr)
	}
}

func TestDeleteProject_NonOwnerForbidden(t *testing.T) {
	_, _, _, _, svc := projectEnvWithCleanup(t)
	if _, err := svc.AddProjectMember(context.Background(), teamMember, seededProjectID, "u2@example.com"); err != nil {
		t.Fatalf("add: %v", err)
	}
	nonOwner := domain.Actor{UserID: "u2", Email: "u2@example.com", Role: domain.SystemRoleTeamMember}
	err := svc.DeleteProject(context.Background(), nonOwner, seededProjectID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden", err)
	}
}

func containsCall(calls []string, want string) bool {
	for _, call := range calls {
		if call == want {
			return true
		}
	}
	return false
}

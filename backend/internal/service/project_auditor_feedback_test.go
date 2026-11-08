package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func seedProjectForAuditor(t *testing.T) (*fakeProjectRepo, *fakeDirectory, *service.ProjectConfigurationService) {
	t.Helper()
	repo, dir, svc := newProjectEnv()
	if _, err := svc.CreateProjectBoard(context.Background(), teamMember, in.CreateProjectInput{
		Name: "Chura", TemplateID: "se",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	dir.users["a1@example.com"] = domain.DirectoryUser{
		ID: "a1", Email: "a1@example.com", Name: "Auditor One", Role: domain.SystemRoleAuditor,
	}
	return repo, dir, svc
}

func TestSelfAssignAuditor_Idempotent(t *testing.T) {
	repo, _, svc := seedProjectForAuditor(t)

	first, err := svc.SelfAssignAuditor(context.Background(), auditor, seededProjectID)
	if err != nil {
		t.Fatalf("first assign: %v", err)
	}
	if first.AuditorUserID != "a1" || first.ProjectID != seededProjectID || first.AssignedAt != fixedNow {
		t.Fatalf("unexpected assignment: %+v", first)
	}

	second, err := svc.SelfAssignAuditor(context.Background(), auditor, seededProjectID)
	if err != nil {
		t.Fatalf("second assign: %v", err)
	}
	if second != first {
		t.Fatalf("second = %+v, want %+v", second, first)
	}
	if len(repo.auditors[seededProjectID]) != 1 {
		t.Fatalf("assignments = %d, want 1", len(repo.auditors[seededProjectID]))
	}
}

func TestSelfAssignAuditor_TeamMemberForbidden(t *testing.T) {
	_, _, svc := seedProjectForAuditor(t)
	_, err := svc.SelfAssignAuditor(context.Background(), teamMember, seededProjectID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden", err)
	}
}

func TestSubmitProjectFeedback_RequiresAssignmentAndValidBody(t *testing.T) {
	_, _, svc := seedProjectForAuditor(t)

	_, err := svc.SubmitProjectFeedback(context.Background(), auditor, seededProjectID, "looks good")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unassigned submit err = %v, want forbidden", err)
	}

	if _, err := svc.SelfAssignAuditor(context.Background(), auditor, seededProjectID); err != nil {
		t.Fatalf("assign: %v", err)
	}

	_, err = svc.SubmitProjectFeedback(context.Background(), auditor, seededProjectID, "   ")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("blank body err = %v, want invalid input", err)
	}

	tooLong := strings.Repeat("a", domain.MaxProjectFeedbackBodyRunes+1)
	_, err = svc.SubmitProjectFeedback(context.Background(), auditor, seededProjectID, tooLong)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("long body err = %v, want invalid input", err)
	}

	got, err := svc.SubmitProjectFeedback(context.Background(), auditor, seededProjectID, "  Ship it  ")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if got.Body != "Ship it" || got.AuthorUserID != "a1" || got.AuthorName != "Auditor One" {
		t.Fatalf("unexpected feedback: %+v", got)
	}
	if utf8.RuneCountInString(got.Body) == 0 {
		t.Fatal("body empty")
	}
}

func TestListProjectFeedback_NewestFirst_MembersCanRead(t *testing.T) {
	repo, _, svc := seedProjectForAuditor(t)
	if _, err := svc.SelfAssignAuditor(context.Background(), auditor, seededProjectID); err != nil {
		t.Fatalf("assign: %v", err)
	}
	first, err := svc.SubmitProjectFeedback(context.Background(), auditor, seededProjectID, "first")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	// Force newest-first ordering in the fake by creating a second entry with a later id.
	second := domain.ProjectFeedback{
		ID: "22222222-2222-4222-8222-222222222222", ProjectID: seededProjectID,
		AuthorUserID: "a1", AuthorName: "Auditor One", AuthorEmail: "a1@example.com",
		Body: "second", CreatedAt: fixedNow.Add(time.Minute),
	}
	if err := repo.CreateFeedback(context.Background(), &second); err != nil {
		t.Fatalf("second: %v", err)
	}

	items, err := svc.ListProjectFeedback(context.Background(), teamMember, seededProjectID)
	if err != nil {
		t.Fatalf("list as member: %v", err)
	}
	if len(items) != 2 || items[0].Body != "second" || items[1].ID != first.ID {
		t.Fatalf("items = %+v", items)
	}

	_, err = svc.SubmitProjectFeedback(context.Background(), teamMember, seededProjectID, "nope")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("member submit err = %v, want forbidden", err)
	}
}

func TestListProjectAuditorsAndFeedback_NonMemberForbidden(t *testing.T) {
	_, _, svc := seedProjectForAuditor(t)
	outsider := domain.Actor{UserID: "u9", Email: "u9@example.com", Role: domain.SystemRoleTeamMember}
	_, err := svc.ListProjectAuditors(context.Background(), outsider, seededProjectID)
	if !errors.Is(err, domain.ErrNotProjectMember) {
		t.Fatalf("auditors err = %v, want not project member", err)
	}
	_, err = svc.ListProjectFeedback(context.Background(), outsider, seededProjectID)
	if !errors.Is(err, domain.ErrNotProjectMember) {
		t.Fatalf("feedback err = %v, want not project member", err)
	}
}

func TestAssignProjectAuditor_TeamMemberAssignsAnAuditor(t *testing.T) {
	repo, _, svc := seedProjectForAuditor(t)

	got, err := svc.AssignProjectAuditor(context.Background(), teamMember, seededProjectID, "a1")
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	if got.AuditorUserID != "a1" || got.AuditorName != "Auditor One" || got.AuditorEmail != "a1@example.com" {
		t.Fatalf("assignment = %+v", got)
	}
	// Picking the same Auditor again (another Work Item) keeps one assignment.
	if _, err := svc.AssignProjectAuditor(context.Background(), teamMember, seededProjectID, "a1"); err != nil {
		t.Fatalf("second assign: %v", err)
	}
	if len(repo.auditors[seededProjectID]) != 1 {
		t.Fatalf("assignments = %d, want 1", len(repo.auditors[seededProjectID]))
	}
}

func TestAssignProjectAuditor_RejectsUsersWhoAreNotAuditors(t *testing.T) {
	repo, _, svc := seedProjectForAuditor(t)

	for _, userID := range []string{"u2", "missing", " "} {
		_, err := svc.AssignProjectAuditor(context.Background(), teamMember, seededProjectID, userID)
		var invalid *domain.InvalidInputError
		if !errors.As(err, &invalid) || invalid.Violations[0].Field != "auditor_user_id" {
			t.Fatalf("user %q: err = %v, want auditor_user_id violation", userID, err)
		}
	}
	if len(repo.auditors[seededProjectID]) != 0 {
		t.Fatalf("no assignment expected, got %d", len(repo.auditors[seededProjectID]))
	}
}

func TestAssignProjectAuditor_RequiresProjectMembership(t *testing.T) {
	_, _, svc := seedProjectForAuditor(t)
	outsider := domain.Actor{UserID: "u2", Email: "u2@example.com", Role: domain.SystemRoleTeamMember}

	if _, err := svc.AssignProjectAuditor(context.Background(), outsider, seededProjectID, "a1"); !errors.Is(err, domain.ErrNotProjectMember) {
		t.Fatalf("outsider err = %v, want not a project member", err)
	}
	if _, err := svc.AssignProjectAuditor(context.Background(), auditor, seededProjectID, "a1"); err == nil {
		t.Fatal("an Auditor must not use the Team Member assignment path")
	}
}

func TestListAuditorDirectory_ListsOnlyAuditorsByName(t *testing.T) {
	_, dir, svc := seedProjectForAuditor(t)
	dir.users["a0@example.com"] = domain.DirectoryUser{ID: "a0", Email: "a0@example.com", Name: "Aaron Auditor", Role: domain.SystemRoleAuditor}

	got, err := svc.ListAuditorDirectory(context.Background(), teamMember)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 || got[0].ID != "a0" || got[1].ID != "a1" {
		t.Fatalf("auditors = %+v, want a0 then a1", got)
	}
}

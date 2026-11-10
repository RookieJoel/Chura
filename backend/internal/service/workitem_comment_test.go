package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

const commentWorkItemID = "wi-1"
const commentProjectID = "11111111-1111-4111-8111-111111111111"

type fakeCommentRepo struct {
	items []domain.WorkItemComment
	err   error
}

func (r *fakeCommentRepo) Create(_ context.Context, comment *domain.WorkItemComment) error {
	if r.err != nil {
		return r.err
	}
	r.items = append(r.items, *comment)
	return nil
}

func (r *fakeCommentRepo) ListByWorkItemID(_ context.Context, workItemID string) ([]domain.WorkItemComment, error) {
	if r.err != nil {
		return nil, r.err
	}
	var outItems []domain.WorkItemComment
	for i := len(r.items) - 1; i >= 0; i-- {
		if r.items[i].WorkItemID == workItemID {
			outItems = append(outItems, r.items[i])
		}
	}
	return outItems, nil
}

func (r *fakeCommentRepo) DeleteByWorkItemID(_ context.Context, workItemID string) error {
	if r.err != nil {
		return r.err
	}
	kept := r.items[:0]
	for _, item := range r.items {
		if item.WorkItemID != workItemID {
			kept = append(kept, item)
		}
	}
	r.items = kept
	return nil
}

type fakeWorkItemRepo struct {
	item domain.WorkItem
	err  error
}

func (r *fakeWorkItemRepo) Create(item domain.WorkItem) (domain.WorkItem, error) {
	return item, nil
}
func (r *fakeWorkItemRepo) Get(id string) (domain.WorkItem, error) {
	if r.err != nil {
		return domain.WorkItem{}, r.err
	}
	if r.item.ID != id {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	return r.item, nil
}
func (r *fakeWorkItemRepo) List(out.WorkItemFilter) ([]domain.WorkItem, error) {
	return nil, nil
}
func (r *fakeWorkItemRepo) Update(item domain.WorkItem) (domain.WorkItem, error) {
	return item, nil
}
func (r *fakeWorkItemRepo) Delete(string) error { return nil }

type commentProjectGateway struct {
	project      *domain.Project
	retrieveErr  error
	assigned     bool
	assignedErr  error
	gotProjectID string
	gotActor     domain.Actor
}

func (g *commentProjectGateway) RetrieveProjectData(_ context.Context, a domain.Actor, projectID string) (*domain.Project, error) {
	g.gotActor, g.gotProjectID = a, projectID
	if g.retrieveErr != nil {
		return nil, g.retrieveErr
	}
	return g.project, nil
}

func (g *commentProjectGateway) IsAuditorAssigned(_ context.Context, projectID, _ string) (bool, error) {
	g.gotProjectID = projectID
	return g.assigned, g.assignedErr
}

type commentDirectory struct {
	users map[string]domain.DirectoryUser
}

func (d *commentDirectory) CreateProjectGroup(context.Context, string) (string, error) {
	return "", nil
}
func (d *commentDirectory) DeleteProjectGroup(context.Context, string) error { return nil }
func (d *commentDirectory) ListProjectMemberships(context.Context, string) ([]string, error) {
	return nil, nil
}
func (d *commentDirectory) ListAuditors(context.Context) ([]domain.DirectoryUser, error) {
	return nil, nil
}
func (d *commentDirectory) FindUserByID(_ context.Context, userID string) (domain.DirectoryUser, error) {
	if user, ok := d.users[userID]; ok {
		return user, nil
	}
	return domain.DirectoryUser{}, domain.ErrNotFound
}
func (d *commentDirectory) FindUserByEmail(context.Context, string) (domain.DirectoryUser, error) {
	return domain.DirectoryUser{}, domain.ErrNotFound
}
func (d *commentDirectory) ListMembers(context.Context, string, string) ([]domain.Member, error) {
	return nil, nil
}
func (d *commentDirectory) AddMember(context.Context, string, string, string, domain.ProjectRole) error {
	return nil
}
func (d *commentDirectory) SetMemberRole(context.Context, string, string, domain.ProjectRole) error {
	return nil
}
func (d *commentDirectory) RemoveMember(context.Context, string, string, string) error { return nil }

func newCommentEnv() (*fakeCommentRepo, *fakeWorkItemRepo, *commentProjectGateway, *commentDirectory, *service.WorkItemCommentService) {
	comments := &fakeCommentRepo{}
	workItems := &fakeWorkItemRepo{item: domain.WorkItem{
		ID: commentWorkItemID, ProjectID: commentProjectID, Title: "Task",
	}}
	gateway := &commentProjectGateway{
		project: &domain.Project{ID: commentProjectID},
	}
	directory := &commentDirectory{users: map[string]domain.DirectoryUser{
		"u1": {ID: "u1", Email: "u1@example.com", Name: "Member One", Role: domain.SystemRoleTeamMember},
		"a1": {ID: "a1", Email: "a1@example.com", Name: "Auditor One", Role: domain.SystemRoleAuditor},
	}}
	svc := service.NewWorkItemCommentService(comments, workItems, gateway, directory)
	return comments, workItems, gateway, directory, svc
}

func TestCreateWorkItemComment_MemberAndAssignedAuditor(t *testing.T) {
	comments, _, gateway, _, svc := newCommentEnv()

	got, err := svc.CreateWorkItemComment(context.Background(), teamMember, commentWorkItemID, "  Looks good  ")
	if err != nil {
		t.Fatalf("member create: %v", err)
	}
	if got.Body != "Looks good" || got.AuthorUserID != "u1" || got.AuthorName != "Member One" || got.WorkItemID != commentWorkItemID {
		t.Fatalf("unexpected comment: %+v", got)
	}
	if gateway.gotProjectID != commentProjectID {
		t.Fatalf("project id = %q", gateway.gotProjectID)
	}

	gateway.assigned = true
	auditorComment, err := svc.CreateWorkItemComment(context.Background(), auditor, commentWorkItemID, "Auditor note")
	if err != nil {
		t.Fatalf("auditor create: %v", err)
	}
	if auditorComment.AuthorUserID != "a1" || auditorComment.Body != "Auditor note" {
		t.Fatalf("unexpected auditor comment: %+v", auditorComment)
	}
	if len(comments.items) != 2 {
		t.Fatalf("stored = %d, want 2", len(comments.items))
	}
}

func TestCreateWorkItemComment_UnassignedAuditorForbidden(t *testing.T) {
	_, _, gateway, _, svc := newCommentEnv()
	gateway.assigned = false

	_, err := svc.CreateWorkItemComment(context.Background(), auditor, commentWorkItemID, "nope")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden", err)
	}
}

func TestCreateWorkItemComment_ValidatesBody(t *testing.T) {
	_, _, _, _, svc := newCommentEnv()

	_, err := svc.CreateWorkItemComment(context.Background(), teamMember, commentWorkItemID, "   ")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("blank body err = %v, want invalid input", err)
	}

	tooLong := strings.Repeat("a", domain.MaxWorkItemCommentBodyRunes+1)
	_, err = svc.CreateWorkItemComment(context.Background(), teamMember, commentWorkItemID, tooLong)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("long body err = %v, want invalid input", err)
	}
}

func TestCreateWorkItemComment_WorkItemMissing(t *testing.T) {
	_, workItems, _, _, svc := newCommentEnv()
	workItems.err = domain.ErrNotFound

	_, err := svc.CreateWorkItemComment(context.Background(), teamMember, commentWorkItemID, "hi")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestCreateWorkItemComment_NonMemberForbidden(t *testing.T) {
	_, _, gateway, _, svc := newCommentEnv()
	gateway.retrieveErr = fmt.Errorf("project %s: %w", commentProjectID, domain.ErrNotProjectMember)

	outsider := domain.Actor{UserID: "u9", Email: "u9@example.com", Role: domain.SystemRoleTeamMember}
	_, err := svc.CreateWorkItemComment(context.Background(), outsider, commentWorkItemID, "hi")
	if !errors.Is(err, domain.ErrNotProjectMember) {
		t.Fatalf("err = %v, want not project member", err)
	}
}

func TestListWorkItemComments_NewestFirst_MemberAndAssignedAuditor(t *testing.T) {
	comments, _, gateway, _, svc := newCommentEnv()
	older := domain.WorkItemComment{
		ID: "c1", WorkItemID: commentWorkItemID, AuthorUserID: "u1",
		AuthorName: "Member One", AuthorEmail: "u1@example.com",
		Body: "older", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	newer := domain.WorkItemComment{
		ID: "c2", WorkItemID: commentWorkItemID, AuthorUserID: "u1",
		AuthorName: "Member One", AuthorEmail: "u1@example.com",
		Body: "newer", CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	comments.items = []domain.WorkItemComment{older, newer}

	items, err := svc.ListWorkItemComments(context.Background(), teamMember, commentWorkItemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 || items[0].Body != "newer" || items[1].Body != "older" {
		t.Fatalf("items = %+v", items)
	}

	gateway.assigned = true
	items, err = svc.ListWorkItemComments(context.Background(), auditor, commentWorkItemID)
	if err != nil || len(items) != 2 {
		t.Fatalf("auditor list err=%v items=%+v", err, items)
	}

	gateway.assigned = false
	_, err = svc.ListWorkItemComments(context.Background(), auditor, commentWorkItemID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unassigned auditor err = %v, want forbidden", err)
	}
}

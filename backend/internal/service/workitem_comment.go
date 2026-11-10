package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/google/uuid"
)

var _ in.WorkItemCommentService = (*WorkItemCommentService)(nil)

// WorkItemCommentService creates and lists append-only Work Item comments.
type WorkItemCommentService struct {
	comments  out.WorkItemCommentRepository
	workItems out.WorkItemRepository
	projects  out.ProjectGateway
	directory out.ProjectDirectory
	newID     func() string
	now       func() time.Time
}

// NewWorkItemCommentService wires comment persistence against Work Items and Project access.
func NewWorkItemCommentService(
	comments out.WorkItemCommentRepository,
	workItems out.WorkItemRepository,
	projects out.ProjectGateway,
	directory out.ProjectDirectory,
) *WorkItemCommentService {
	return &WorkItemCommentService{
		comments:  comments,
		workItems: workItems,
		projects:  projects,
		directory: directory,
		newID:     uuid.NewString,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *WorkItemCommentService) ListWorkItemComments(ctx context.Context, actor domain.Actor, workItemID string) ([]domain.WorkItemComment, error) {
	if _, err := s.requireCommentAccess(ctx, actor, workItemID); err != nil {
		return nil, fmt.Errorf("list work item comments: %w", err)
	}
	items, err := s.comments.ListByWorkItemID(ctx, workItemID)
	if err != nil {
		return nil, fmt.Errorf("list work item comments: %w", err)
	}
	if items == nil {
		items = []domain.WorkItemComment{}
	}
	return items, nil
}

func (s *WorkItemCommentService) CreateWorkItemComment(ctx context.Context, actor domain.Actor, workItemID, body string) (domain.WorkItemComment, error) {
	if _, err := s.requireCommentAccess(ctx, actor, workItemID); err != nil {
		return domain.WorkItemComment{}, fmt.Errorf("create work item comment: %w", err)
	}
	trimmed := strings.TrimSpace(body)
	switch {
	case trimmed == "":
		return domain.WorkItemComment{}, fmt.Errorf("create work item comment: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "body", Message: "body is required"},
		}})
	case utf8.RuneCountInString(trimmed) > domain.MaxWorkItemCommentBodyRunes:
		return domain.WorkItemComment{}, fmt.Errorf("create work item comment: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "body", Message: fmt.Sprintf("body must be at most %d characters", domain.MaxWorkItemCommentBodyRunes)},
		}})
	}

	name, email := resolveAuthor(ctx, s.directory, actor)

	comment := domain.WorkItemComment{
		ID:           s.newID(),
		WorkItemID:   workItemID,
		AuthorUserID: actor.UserID,
		AuthorName:   name,
		AuthorEmail:  email,
		Body:         trimmed,
		CreatedAt:    s.now(),
	}
	if err := s.comments.Create(ctx, &comment); err != nil {
		return domain.WorkItemComment{}, fmt.Errorf("create work item comment: %w", err)
	}
	return comment, nil
}

// requireCommentAccess loads the Work Item and ensures the actor is a Project
// member or an assigned Auditor.
func (s *WorkItemCommentService) requireCommentAccess(ctx context.Context, actor domain.Actor, workItemID string) (domain.WorkItem, error) {
	if err := actor.Validate(); err != nil {
		return domain.WorkItem{}, err
	}
	if strings.TrimSpace(workItemID) == "" {
		return domain.WorkItem{}, domain.ErrInvalidID
	}
	item, err := s.workItems.Get(workItemID)
	if err != nil {
		return domain.WorkItem{}, err
	}
	if _, err := s.projects.RetrieveProjectData(ctx, actor, item.ProjectID); err != nil {
		return domain.WorkItem{}, err
	}
	if actor.Role == domain.SystemRoleAuditor {
		assigned, err := s.projects.IsAuditorAssigned(ctx, item.ProjectID, actor.UserID)
		if err != nil {
			return domain.WorkItem{}, err
		}
		if !assigned {
			return domain.WorkItem{}, domain.ErrForbidden
		}
	}
	return item, nil
}

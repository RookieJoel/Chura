package in

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// WorkItemCommentService lists and creates append-only Work Item comments.
type WorkItemCommentService interface {
	ListWorkItemComments(ctx context.Context, actor domain.Actor, workItemID string) ([]domain.WorkItemComment, error)
	CreateWorkItemComment(ctx context.Context, actor domain.Actor, workItemID, body string) (domain.WorkItemComment, error)
}

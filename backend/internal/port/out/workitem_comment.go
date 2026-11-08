package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// WorkItemCommentRepository persists append-only Work Item comments.
type WorkItemCommentRepository interface {
	Create(ctx context.Context, comment *domain.WorkItemComment) error
	ListByWorkItemID(ctx context.Context, workItemID string) ([]domain.WorkItemComment, error)
	DeleteByWorkItemID(ctx context.Context, workItemID string) error
}

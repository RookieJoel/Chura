package out

import (
	"context"
)

// ProjectSprintStore counts and permanently deletes Sprints owned by a Project.
type ProjectSprintStore interface {
	CountByProject(ctx context.Context, projectID string) (int, error)
	HardDeleteByProject(ctx context.Context, projectID string) error
}

// ProjectWorkItemStore counts and permanently deletes Work Items (and their
// comments) owned by a Project.
type ProjectWorkItemStore interface {
	CountByProject(ctx context.Context, projectID string) (int, error)
	DeleteByProject(ctx context.Context, projectID string) error
}

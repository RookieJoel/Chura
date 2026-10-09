package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type ProjectRepository interface {
	// Create stores the project and its creator member in one transaction.
	Create(ctx context.Context, project *domain.Project) error
	// GetByID returns the project with members ordered by added_at; domain.ErrNotFound if missing.
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	// AddMember returns domain.ErrConflict on duplicate and domain.ErrNotFound if the project is gone.
	AddMember(ctx context.Context, projectID string, member domain.Member) error
	// UpdateMemberRole atomically locks the project's members, passes them to
	// guard (whose error aborts the change unchanged) and then updates the role.
	// It returns domain.ErrNotFound if the user is not a member.
	UpdateMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole, guard func(members []domain.Member) error) error
}

// ProjectGateway lets other services read a Project with the caller's visibility rules.
type ProjectGateway interface {
	RetrieveProjectData(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error)
}

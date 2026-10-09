package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type ProjectRepository interface {
	// Create stores the project row; membership lives in the ProjectDirectory.
	Create(ctx context.Context, project *domain.Project) error
	// GetByID returns the project row with Members nil; domain.ErrNotFound if missing.
	GetByID(ctx context.Context, id string) (*domain.Project, error)
}

// ProjectDirectory is the identity store holding Project membership and Project Role.
// Every failure that is not a definite answer is wrapped with domain.ErrUnavailable.
type ProjectDirectory interface {
	CreateProjectGroup(ctx context.Context, projectID string) (groupID string, err error)
	DeleteProjectGroup(ctx context.Context, groupID string) error
	FindUserByEmail(ctx context.Context, email string) (domain.DirectoryUser, error) // ErrNotFound
	ListMembers(ctx context.Context, groupID, projectID string) ([]domain.Member, error)
	AddMember(ctx context.Context, groupID, projectID, userID string, role domain.ProjectRole) error
	SetMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error
	RemoveMember(ctx context.Context, groupID, projectID, userID string) error // undo only
}

// ProjectGateway lets other services read a Project with the caller's visibility rules.
type ProjectGateway interface {
	RetrieveProjectData(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error)
}

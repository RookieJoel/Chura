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
	// ListSummaries applies the optional project-ID scope before counting and paging.
	ListSummaries(ctx context.Context, query ProjectListQuery) ([]domain.ProjectSummary, int, error)
	// WithMembershipLock runs fn while holding a transaction-scoped advisory
	// lock for the project, serialising membership changes across instances.
	// fn's error is returned unchanged.
	WithMembershipLock(ctx context.Context, projectID string, fn func(ctx context.Context) error) error
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, projectID string) error
	AssignAuditor(ctx context.Context, assignment domain.ProjectAuditorAssignment) (domain.ProjectAuditorAssignment, error)
	ListAuditors(ctx context.Context, projectID string) ([]domain.ProjectAuditorAssignment, error)
	IsAuditorAssigned(ctx context.Context, projectID, auditorUserID string) (bool, error)
	CreateFeedback(ctx context.Context, feedback *domain.ProjectFeedback) error
	ListFeedback(ctx context.Context, projectID string) ([]domain.ProjectFeedback, error)
}

type ProjectListQuery struct {
	ProjectIDs []string
	Offset     int
	Limit      int
}

// ProjectDirectory is the identity store holding Project membership and Project Role.
// Every failure that is not a definite answer is wrapped with domain.ErrUnavailable.
type ProjectDirectory interface {
	CreateProjectGroup(ctx context.Context, projectID string) (groupID string, err error)
	DeleteProjectGroup(ctx context.Context, groupID string) error
	ListProjectMemberships(ctx context.Context, userID string) ([]string, error)
	FindUserByID(ctx context.Context, userID string) (domain.DirectoryUser, error)
	FindUserByEmail(ctx context.Context, email string) (domain.DirectoryUser, error) // ErrNotFound
	// ListAuditors returns every user holding the Auditor system role.
	ListAuditors(ctx context.Context) ([]domain.DirectoryUser, error)
	ListMembers(ctx context.Context, groupID, projectID string) ([]domain.Member, error)
	AddMember(ctx context.Context, groupID, projectID, userID string, role domain.ProjectRole) error
	SetMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole) error
	RemoveMember(ctx context.Context, groupID, projectID, userID string) error // undo only
}

// ProjectGateway lets other services read a Project with the caller's visibility rules.
type ProjectGateway interface {
	RetrieveProjectData(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error)
	IsAuditorAssigned(ctx context.Context, projectID, auditorUserID string) (bool, error)
}

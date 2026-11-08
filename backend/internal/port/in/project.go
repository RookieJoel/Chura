package in

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type CreateProjectInput struct {
	Name        string
	Description string
	TemplateID  string
}

const (
	DefaultProjectPageSize = 15
	MaxProjectPageSize     = 30
)

type ProjectListQuery struct {
	Page     int
	PageSize int
}

type ProjectListResult struct {
	Items    []domain.ProjectSummary `json:"items"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	Total    int                     `json:"total"`
}

type UpdateProjectInput struct {
	Name        *string
	Description *string
	TemplateID  *string
}

type ProjectConfigurationService interface {
	GetAvailableTemplates(ctx context.Context, actor domain.Actor) ([]domain.Template, error)
	RetrieveTemplateDefinition(ctx context.Context, actor domain.Actor, templateID string) (domain.Template, error)
	ListAccessibleProjects(ctx context.Context, actor domain.Actor, query ProjectListQuery) (ProjectListResult, error)
	CreateProjectBoard(ctx context.Context, actor domain.Actor, in CreateProjectInput) (*domain.Project, error)
	AddProjectMember(ctx context.Context, actor domain.Actor, projectID, email string) (*domain.Project, error)
	AssignProjectRole(ctx context.Context, actor domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error)
	RemoveProjectMember(ctx context.Context, actor domain.Actor, projectID, userID string) (*domain.Project, error)
	RetrieveProjectData(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error)
	ValidateWorkItemAttributes(ctx context.Context, actor domain.Actor, projectID string, attrs domain.WorkItemAttributes) (domain.ValidationResult, error)
	SelfAssignAuditor(ctx context.Context, actor domain.Actor, projectID string) (domain.ProjectAuditorAssignment, error)
	ListProjectAuditors(ctx context.Context, actor domain.Actor, projectID string) ([]domain.ProjectAuditorAssignment, error)
	// ListAuditorDirectory lists every Auditor in the system, for picking one
	// as a Work Item's Auditor.
	ListAuditorDirectory(ctx context.Context, actor domain.Actor) ([]domain.DirectoryUser, error)
	// AssignProjectAuditor lets a Team Member of the Project assign an Auditor
	// to it (picking that Auditor on a Work Item). Assigning twice is a no-op.
	AssignProjectAuditor(ctx context.Context, actor domain.Actor, projectID, auditorUserID string) (domain.ProjectAuditorAssignment, error)
	SubmitProjectFeedback(ctx context.Context, actor domain.Actor, projectID, body string) (domain.ProjectFeedback, error)
	ListProjectFeedback(ctx context.Context, actor domain.Actor, projectID string) ([]domain.ProjectFeedback, error)
	UpdateProject(ctx context.Context, actor domain.Actor, projectID string, in UpdateProjectInput) (*domain.Project, error)
	DeleteProject(ctx context.Context, actor domain.Actor, projectID string) error
}

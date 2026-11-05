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

type ProjectConfigurationService interface {
	GetAvailableTemplates(ctx context.Context, actor domain.Actor) ([]domain.Template, error)
	RetrieveTemplateDefinition(ctx context.Context, actor domain.Actor, templateID string) (domain.Template, error)
	ListAccessibleProjects(ctx context.Context, actor domain.Actor, query ProjectListQuery) (ProjectListResult, error)
	CreateProjectBoard(ctx context.Context, actor domain.Actor, in CreateProjectInput) (*domain.Project, error)
	AddProjectMember(ctx context.Context, actor domain.Actor, projectID, email string) (*domain.Project, error)
	AssignProjectRole(ctx context.Context, actor domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error)
	RetrieveProjectData(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error)
	ValidateWorkItemAttributes(ctx context.Context, actor domain.Actor, projectID string, attrs domain.WorkItemAttributes) (domain.ValidationResult, error)
}

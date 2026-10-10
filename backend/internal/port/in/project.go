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

type ProjectConfigurationService interface {
	GetAvailableTemplates(ctx context.Context, actor domain.Actor) ([]domain.Template, error)
	RetrieveTemplateDefinition(ctx context.Context, actor domain.Actor, templateID string) (domain.Template, error)
	CreateProjectBoard(ctx context.Context, actor domain.Actor, in CreateProjectInput) (*domain.Project, error)
	AddProjectMember(ctx context.Context, actor domain.Actor, projectID, email string) (*domain.Project, error)
	AssignProjectRole(ctx context.Context, actor domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error)
	RetrieveProjectData(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error)
	ValidateWorkItemAttributes(ctx context.Context, actor domain.Actor, projectID string, attrs domain.WorkItemAttributes) (domain.ValidationResult, error)
}

package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type ProjectConfigurationService struct {
	repo  out.ProjectRepository
	newID func() string
	now   func() time.Time
}

var (
	_ in.ProjectConfigurationService = (*ProjectConfigurationService)(nil)
	_ out.ProjectGateway             = (*ProjectConfigurationService)(nil)
)

// ProjectOption customises a ProjectConfigurationService (used by tests).
type ProjectOption func(*ProjectConfigurationService)

func WithIDGenerator(newID func() string) ProjectOption {
	return func(s *ProjectConfigurationService) { s.newID = newID }
}

func WithClock(now func() time.Time) ProjectOption {
	return func(s *ProjectConfigurationService) { s.now = now }
}

func NewProjectConfigurationService(repo out.ProjectRepository, opts ...ProjectOption) *ProjectConfigurationService {
	s := &ProjectConfigurationService{
		repo:  repo,
		newID: uuid.NewString,
		now:   func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *ProjectConfigurationService) GetAvailableTemplates(ctx context.Context, actor domain.Actor) ([]domain.Template, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("get available templates: %w", err)
	}
	return domain.AvailableTemplates(), nil
}

func (s *ProjectConfigurationService) RetrieveTemplateDefinition(ctx context.Context, actor domain.Actor, templateID string) (domain.Template, error) {
	if err := actor.Validate(); err != nil {
		return domain.Template{}, fmt.Errorf("retrieve template definition: %w", err)
	}
	template, ok := domain.TemplateByID(templateID)
	if !ok {
		return domain.Template{}, fmt.Errorf("template %q: %w", templateID, domain.ErrNotFound)
	}
	return template, nil
}

func (s *ProjectConfigurationService) CreateProjectBoard(ctx context.Context, actor domain.Actor, input in.CreateProjectInput) (*domain.Project, error) {
	if err := actor.RequireTeamMember(); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	template, err := domain.ValidateProjectInput(input.Name, input.Description, input.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	now := s.now()
	project := &domain.Project{
		ID:          s.newID(),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		TemplateID:  template.ID,
		Mode:        template.Mode,
		CreatedBy:   actor.UserID,
		Members:     []domain.Member{{UserID: actor.UserID, Role: template.CreatorRole, AddedAt: now}},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (s *ProjectConfigurationService) AddProjectMember(ctx context.Context, actor domain.Actor, projectID, userID string) (*domain.Project, error) {
	project, err := s.loadMutable(ctx, actor, projectID)
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}
	userID = strings.TrimSpace(userID)
	if userID == "" || utf8.RuneCountInString(userID) > domain.MaxUserIDRunes {
		return nil, fmt.Errorf("add project member: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "user_id", Message: fmt.Sprintf("user_id is required and must be at most %d characters", domain.MaxUserIDRunes)},
		}})
	}
	if _, isMember := project.MemberByID(userID); isMember {
		return nil, fmt.Errorf("add project member: user %q is already a member: %w", userID, domain.ErrConflict)
	}
	template, err := templateOf(project)
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}
	member := domain.Member{UserID: userID, Role: template.DefaultRole, AddedAt: s.now()}
	if err := s.repo.AddMember(ctx, project.ID, member); err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}
	return s.repo.GetByID(ctx, project.ID)
}

func (s *ProjectConfigurationService) AssignProjectRole(ctx context.Context, actor domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error) {
	project, err := s.loadMutable(ctx, actor, projectID)
	if err != nil {
		return nil, fmt.Errorf("assign project role: %w", err)
	}
	template, err := templateOf(project)
	if err != nil {
		return nil, fmt.Errorf("assign project role: %w", err)
	}
	if !slices.Contains(template.Roles, role) {
		return nil, fmt.Errorf("assign project role: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "role", Message: fmt.Sprintf("role must be one of %v", template.Roles)},
		}})
	}
	userID = strings.TrimSpace(userID)
	guard := keepCreatorRoleGuard(userID, role, template.CreatorRole)
	if err := s.repo.UpdateMemberRole(ctx, project.ID, userID, role, guard); err != nil {
		return nil, fmt.Errorf("assign project role: %w", err)
	}
	return s.repo.GetByID(ctx, project.ID)
}

// keepCreatorRoleGuard is evaluated by the repository against the locked,
// current members: the target must be a member, and the last holder of the
// creator role may not be moved off it.
func keepCreatorRoleGuard(userID string, role, creatorRole domain.ProjectRole) func([]domain.Member) error {
	return func(members []domain.Member) error {
		current := domain.Project{Members: members}
		target, isMember := current.MemberByID(userID)
		if !isMember {
			return fmt.Errorf("user %q is not a member: %w", userID, domain.ErrNotFound)
		}
		if target.Role == creatorRole && role != creatorRole && current.CountRole(creatorRole) == 1 {
			return fmt.Errorf("project must keep a %s: %w", creatorRole, domain.ErrConflict)
		}
		return nil
	}
}

func (s *ProjectConfigurationService) RetrieveProjectData(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error) {
	project, err := s.loadVisible(ctx, actor, projectID)
	if err != nil {
		return nil, fmt.Errorf("retrieve project data: %w", err)
	}
	return project, nil
}

// loadVisible loads a project the actor may see: validated actor, well-formed
// id, then Auditors see everything and Team Members only their own projects
// (existence of others is hidden behind ErrNotFound).
func (s *ProjectConfigurationService) loadVisible(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error) {
	if err := actor.Validate(); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(projectID); err != nil {
		return nil, fmt.Errorf("project id %q: %w", projectID, domain.ErrInvalidID)
	}
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if actor.Role == domain.SystemRoleAuditor {
		return project, nil
	}
	if _, isMember := project.MemberByID(actor.UserID); !isMember {
		return nil, fmt.Errorf("project %s: %w", projectID, domain.ErrNotFound)
	}
	return project, nil
}

// loadMutable loads a visible project and additionally requires the actor to
// be a Team Member (loadVisible already guarantees they belong to it).
func (s *ProjectConfigurationService) loadMutable(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, error) {
	project, err := s.loadVisible(ctx, actor, projectID)
	if err != nil {
		return nil, err
	}
	if err := actor.RequireTeamMember(); err != nil {
		return nil, fmt.Errorf("project %s: %w", projectID, err)
	}
	return project, nil
}

func (s *ProjectConfigurationService) ValidateWorkItemAttributes(ctx context.Context, actor domain.Actor, projectID string, attrs domain.WorkItemAttributes) (domain.ValidationResult, error) {
	project, err := s.loadVisible(ctx, actor, projectID)
	if err != nil {
		return domain.ValidationResult{}, fmt.Errorf("validate work item attributes: %w", err)
	}
	template, err := templateOf(project)
	if err != nil {
		return domain.ValidationResult{}, fmt.Errorf("validate work item attributes: %w", err)
	}
	return domain.ValidateWorkItemAttributes(template, attrs), nil
}

// templateOf returns the project's Template. A stored project whose template
// is missing from the catalog is a data-integrity fault, reported as an
// internal error rather than a client error.
func templateOf(project *domain.Project) (domain.Template, error) {
	template, ok := domain.TemplateByID(project.TemplateID)
	if !ok {
		return domain.Template{}, fmt.Errorf("project %s references unknown template %q", project.ID, project.TemplateID)
	}
	return template, nil
}

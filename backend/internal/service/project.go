package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	repo      out.ProjectRepository
	directory out.ProjectDirectory
	sprints   out.ProjectSprintStore
	workItems out.ProjectWorkItemStore
	newID     func() string
	now       func() time.Time
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

func WithSprintStore(store out.ProjectSprintStore) ProjectOption {
	return func(s *ProjectConfigurationService) { s.sprints = store }
}

func WithWorkItemStore(store out.ProjectWorkItemStore) ProjectOption {
	return func(s *ProjectConfigurationService) { s.workItems = store }
}

func NewProjectConfigurationService(repo out.ProjectRepository, directory out.ProjectDirectory, opts ...ProjectOption) *ProjectConfigurationService {
	s := &ProjectConfigurationService{
		repo:      repo,
		directory: directory,
		newID:     uuid.NewString,
		now:       func() time.Time { return time.Now().UTC() },
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

func (s *ProjectConfigurationService) ListAccessibleProjects(ctx context.Context, actor domain.Actor, query in.ProjectListQuery) (in.ProjectListResult, error) {
	if err := actor.Validate(); err != nil {
		return in.ProjectListResult{}, fmt.Errorf("list accessible projects: %w", err)
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > in.MaxProjectPageSize {
		return in.ProjectListResult{}, fmt.Errorf("list accessible projects: %w", domain.ErrInvalidInput)
	}
	maxInt := int(^uint(0) >> 1)
	if query.Page-1 > maxInt/query.PageSize {
		return in.ProjectListResult{}, fmt.Errorf("list accessible projects: %w", domain.ErrInvalidInput)
	}

	repositoryQuery := out.ProjectListQuery{
		Offset: (query.Page - 1) * query.PageSize,
		Limit:  query.PageSize,
	}
	if actor.Role == domain.SystemRoleTeamMember {
		projectIDs, err := s.directory.ListProjectMemberships(ctx, actor.UserID)
		if err != nil {
			return in.ProjectListResult{}, fmt.Errorf("list accessible projects: %w", err)
		}
		if len(projectIDs) == 0 {
			return in.ProjectListResult{
				Items: []domain.ProjectSummary{}, Page: query.Page, PageSize: query.PageSize,
			}, nil
		}
		repositoryQuery.ProjectIDs = projectIDs
	}

	items, total, err := s.repo.ListSummaries(ctx, repositoryQuery)
	if err != nil {
		return in.ProjectListResult{}, fmt.Errorf("list accessible projects: %w", err)
	}
	if items == nil {
		items = []domain.ProjectSummary{}
	}
	return in.ProjectListResult{Items: items, Page: query.Page, PageSize: query.PageSize, Total: total}, nil
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
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	// Keycloak has no transactions: write in order and undo in reverse on failure.
	groupID, err := s.directory.CreateProjectGroup(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	project.GroupID = groupID
	if err := s.directory.AddMember(ctx, groupID, project.ID, actor.UserID, template.CreatorRole); err != nil {
		s.undoCreate(ctx, project)
		return nil, fmt.Errorf("create project: %w", err)
	}
	if err := s.repo.Create(ctx, project); err != nil {
		s.undoCreate(ctx, project)
		return nil, fmt.Errorf("create project: %w", err)
	}
	project.Members = []domain.Member{{UserID: actor.UserID, Email: actor.Email, Role: template.CreatorRole}}
	return project, nil
}

// undoCreate best-effort reverts the directory writes of a failed create.
// Failures are logged for manual clean-up and never returned: the caller must
// see the original error.
//
// RemoveMember is always attempted (it is idempotent): a timed-out AddMember
// may still have written the role attribute.
func (s *ProjectConfigurationService) undoCreate(ctx context.Context, project *domain.Project) {
	ctx = context.WithoutCancel(ctx)
	if err := s.directory.RemoveMember(ctx, project.GroupID, project.ID, project.CreatedBy); err != nil {
		slog.Error("undo create project: remove creator failed", "project_id", project.ID, "group_id", project.GroupID, "error", err)
	}
	if err := s.directory.DeleteProjectGroup(ctx, project.GroupID); err != nil {
		slog.Error("undo create project: delete group failed", "project_id", project.ID, "group_id", project.GroupID, "error", err)
	}
}

func (s *ProjectConfigurationService) AddProjectMember(ctx context.Context, actor domain.Actor, projectID, email string) (*domain.Project, error) {
	project, template, err := s.loadOwnerMutable(ctx, actor, projectID)
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}
	email = strings.TrimSpace(email)
	if email == "" || utf8.RuneCountInString(email) > domain.MaxEmailRunes || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("add project member: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "email", Message: fmt.Sprintf("email is required, must contain @ and be at most %d characters", domain.MaxEmailRunes)},
		}})
	}
	user, err := s.directory.FindUserByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("add project member: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "email", Message: "this user does not exist"},
		}})
	}
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}
	if user.Role != domain.SystemRoleTeamMember {
		return nil, fmt.Errorf("add project member: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "email", Message: "user is not a Team Member"},
		}})
	}
	err = s.repo.WithMembershipLock(ctx, project.ID, func(ctx context.Context) error {
		if err := s.refreshMembers(ctx, project); err != nil {
			return err
		}
		if err := requireProjectOwner(actor, project, template); err != nil {
			return err
		}
		if _, isMember := project.MemberByID(user.ID); isMember {
			return fmt.Errorf("user %q is already a member: %w", user.ID, domain.ErrAlreadyMember)
		}
		return s.directory.AddMember(ctx, project.GroupID, project.ID, user.ID, "")
	})
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}
	return s.loadVisible(ctx, actor, project.ID)
}

func (s *ProjectConfigurationService) AssignProjectRole(ctx context.Context, actor domain.Actor, projectID, userID string, role domain.ProjectRole) (*domain.Project, error) {
	project, template, err := s.loadOwnerMutable(ctx, actor, projectID)
	if err != nil {
		return nil, fmt.Errorf("assign project role: %w", err)
	}
	if role != "" && !slices.Contains(template.Roles, role) {
		return nil, fmt.Errorf("assign project role: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "role", Message: fmt.Sprintf("role must be one of %v", template.Roles)},
		}})
	}
	userID = strings.TrimSpace(userID)
	err = s.repo.WithMembershipLock(ctx, project.ID, func(ctx context.Context) error {
		if err := s.refreshMembers(ctx, project); err != nil {
			return err
		}
		if err := requireProjectOwner(actor, project, template); err != nil {
			return err
		}
		if err := keepCreatorRoleGuard(userID, role, template.CreatorRole)(project.Members); err != nil {
			return err
		}
		return s.directory.SetMemberRole(ctx, project.ID, userID, role)
	})
	if err != nil {
		return nil, fmt.Errorf("assign project role: %w", err)
	}
	return s.loadVisible(ctx, actor, project.ID)
}

func (s *ProjectConfigurationService) RemoveProjectMember(ctx context.Context, actor domain.Actor, projectID, userID string) (*domain.Project, error) {
	project, template, err := s.loadOwnerMutable(ctx, actor, projectID)
	if err != nil {
		return nil, fmt.Errorf("remove project member: %w", err)
	}
	userID = strings.TrimSpace(userID)
	if userID == "" || utf8.RuneCountInString(userID) > domain.MaxUserIDRunes {
		return nil, fmt.Errorf("remove project member: %w", domain.ErrInvalidInput)
	}
	err = s.repo.WithMembershipLock(ctx, project.ID, func(ctx context.Context) error {
		if err := s.refreshMembers(ctx, project); err != nil {
			return err
		}
		if err := requireProjectOwner(actor, project, template); err != nil {
			return err
		}
		member, exists := project.MemberByID(userID)
		if !exists {
			return fmt.Errorf("user %q is not a member: %w", userID, domain.ErrNotFound)
		}
		if member.Role == template.CreatorRole && project.CountRole(template.CreatorRole) == 1 {
			return fmt.Errorf("project must keep a %s: %w", template.CreatorRole, domain.ErrConflict)
		}
		if err := s.directory.RemoveMember(ctx, project.GroupID, project.ID, userID); err != nil {
			return err
		}
		project.Members = slices.DeleteFunc(project.Members, func(member domain.Member) bool {
			return member.UserID == userID
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("remove project member: %w", err)
	}
	if actor.UserID == userID {
		return project, nil
	}
	return s.loadVisible(ctx, actor, project.ID)
}

// refreshMembers replaces project.Members with the directory's current members.
func (s *ProjectConfigurationService) refreshMembers(ctx context.Context, project *domain.Project) error {
	members, err := s.directory.ListMembers(ctx, project.GroupID, project.ID)
	if err != nil {
		return fmt.Errorf("list members of project %s: %w", project.ID, err)
	}
	project.Members = members
	return nil
}

// keepCreatorRoleGuard is evaluated against the current members: the target
// must be a member, and the last holder of the creator role may not be moved
// off it.
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

func (s *ProjectConfigurationService) IsAuditorAssigned(ctx context.Context, projectID, auditorUserID string) (bool, error) {
	return s.repo.IsAuditorAssigned(ctx, projectID, auditorUserID)
}

// loadVisible loads a project the actor may see: validated actor, well-formed
// id, then Auditors see everything and Team Members only their own projects
// (a non-member gets ErrNotProjectMember; a missing project is ErrNotFound).
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
	if err := s.refreshMembers(ctx, project); err != nil {
		return nil, err
	}
	if actor.Role == domain.SystemRoleAuditor {
		return project, nil
	}
	if _, isMember := project.MemberByID(actor.UserID); !isMember {
		return nil, fmt.Errorf("project %s: %w", projectID, domain.ErrNotProjectMember)
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

func (s *ProjectConfigurationService) loadOwnerMutable(ctx context.Context, actor domain.Actor, projectID string) (*domain.Project, domain.Template, error) {
	project, err := s.loadMutable(ctx, actor, projectID)
	if err != nil {
		return nil, domain.Template{}, err
	}
	template, err := templateOf(project)
	if err != nil {
		return nil, domain.Template{}, err
	}
	if err := requireProjectOwner(actor, project, template); err != nil {
		return nil, domain.Template{}, err
	}
	return project, template, nil
}

func requireProjectOwner(actor domain.Actor, project *domain.Project, template domain.Template) error {
	member, isMember := project.MemberByID(actor.UserID)
	if !isMember {
		return fmt.Errorf("project %s: %w", project.ID, domain.ErrNotProjectMember)
	}
	if member.Role != template.CreatorRole {
		return fmt.Errorf("project %s requires the %s Project Role: %w", project.ID, template.CreatorRole, domain.ErrForbidden)
	}
	return nil
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

func (s *ProjectConfigurationService) SelfAssignAuditor(ctx context.Context, actor domain.Actor, projectID string) (domain.ProjectAuditorAssignment, error) {
	if err := actor.Validate(); err != nil {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("self-assign auditor: %w", err)
	}
	if actor.Role != domain.SystemRoleAuditor {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("self-assign auditor: %w", domain.ErrForbidden)
	}
	project, err := s.loadVisible(ctx, actor, projectID)
	if err != nil {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("self-assign auditor: %w", err)
	}
	name, email := resolveAuthor(ctx, s.directory, actor)
	assignment, err := s.repo.AssignAuditor(ctx, domain.ProjectAuditorAssignment{
		ProjectID:     project.ID,
		AuditorUserID: actor.UserID,
		AuditorName:   name,
		AuditorEmail:  email,
		AssignedAt:    s.now(),
	})
	if err != nil {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("self-assign auditor: %w", err)
	}
	return assignment, nil
}

func (s *ProjectConfigurationService) ListAuditorDirectory(ctx context.Context, actor domain.Actor) ([]domain.DirectoryUser, error) {
	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("list auditor directory: %w", err)
	}
	auditors, err := s.directory.ListAuditors(ctx)
	if err != nil {
		return nil, fmt.Errorf("list auditor directory: %w", err)
	}
	slices.SortFunc(auditors, func(a, b domain.DirectoryUser) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	if auditors == nil {
		auditors = []domain.DirectoryUser{}
	}
	return auditors, nil
}

func (s *ProjectConfigurationService) AssignProjectAuditor(ctx context.Context, actor domain.Actor, projectID, auditorUserID string) (domain.ProjectAuditorAssignment, error) {
	project, err := s.loadMutable(ctx, actor, projectID)
	if err != nil {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("assign project auditor: %w", err)
	}
	auditorUserID = strings.TrimSpace(auditorUserID)
	notAuditor := &domain.InvalidInputError{Violations: []domain.Violation{
		{Field: "auditor_user_id", Message: "choose a user with the Auditor role"},
	}}
	if auditorUserID == "" || utf8.RuneCountInString(auditorUserID) > domain.MaxUserIDRunes {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("assign project auditor: %w", notAuditor)
	}
	auditors, err := s.directory.ListAuditors(ctx)
	if err != nil {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("assign project auditor: %w", err)
	}
	index := slices.IndexFunc(auditors, func(user domain.DirectoryUser) bool { return user.ID == auditorUserID })
	if index < 0 {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("assign project auditor: %w", notAuditor)
	}
	auditor := auditors[index]
	assignment, err := s.repo.AssignAuditor(ctx, domain.ProjectAuditorAssignment{
		ProjectID:     project.ID,
		AuditorUserID: auditor.ID,
		AuditorName:   truncateRunes(auditor.Name, maxAuthorNameRunes),
		AuditorEmail:  truncateRunes(auditor.Email, maxAuthorEmailRunes),
		AssignedAt:    s.now(),
	})
	if err != nil {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("assign project auditor: %w", err)
	}
	return assignment, nil
}

func (s *ProjectConfigurationService) ListProjectAuditors(ctx context.Context, actor domain.Actor, projectID string) ([]domain.ProjectAuditorAssignment, error) {
	if _, err := s.loadVisible(ctx, actor, projectID); err != nil {
		return nil, fmt.Errorf("list project auditors: %w", err)
	}
	assignments, err := s.repo.ListAuditors(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project auditors: %w", err)
	}
	if assignments == nil {
		assignments = []domain.ProjectAuditorAssignment{}
	}
	return assignments, nil
}

func (s *ProjectConfigurationService) SubmitProjectFeedback(ctx context.Context, actor domain.Actor, projectID, body string) (domain.ProjectFeedback, error) {
	if err := actor.Validate(); err != nil {
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", err)
	}
	if actor.Role != domain.SystemRoleAuditor {
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", domain.ErrForbidden)
	}
	project, err := s.loadVisible(ctx, actor, projectID)
	if err != nil {
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", err)
	}
	assigned, err := s.repo.IsAuditorAssigned(ctx, project.ID, actor.UserID)
	if err != nil {
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", err)
	}
	if !assigned {
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", domain.ErrForbidden)
	}
	trimmed := strings.TrimSpace(body)
	switch {
	case trimmed == "":
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "body", Message: "body is required"},
		}})
	case utf8.RuneCountInString(trimmed) > domain.MaxProjectFeedbackBodyRunes:
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "body", Message: fmt.Sprintf("body must be at most %d characters", domain.MaxProjectFeedbackBodyRunes)},
		}})
	}
	name, email := resolveAuthor(ctx, s.directory, actor)
	feedback := domain.ProjectFeedback{
		ID:           s.newID(),
		ProjectID:    project.ID,
		AuthorUserID: actor.UserID,
		AuthorName:   name,
		AuthorEmail:  email,
		Body:         trimmed,
		CreatedAt:    s.now(),
	}
	if err := s.repo.CreateFeedback(ctx, &feedback); err != nil {
		return domain.ProjectFeedback{}, fmt.Errorf("submit project feedback: %w", err)
	}
	return feedback, nil
}

func (s *ProjectConfigurationService) ListProjectFeedback(ctx context.Context, actor domain.Actor, projectID string) ([]domain.ProjectFeedback, error) {
	if _, err := s.loadVisible(ctx, actor, projectID); err != nil {
		return nil, fmt.Errorf("list project feedback: %w", err)
	}
	items, err := s.repo.ListFeedback(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project feedback: %w", err)
	}
	if items == nil {
		items = []domain.ProjectFeedback{}
	}
	return items, nil
}

func (s *ProjectConfigurationService) UpdateProject(ctx context.Context, actor domain.Actor, projectID string, input in.UpdateProjectInput) (*domain.Project, error) {
	project, template, err := s.loadOwnerMutable(ctx, actor, projectID)
	if err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	name := project.Name
	description := project.Description
	if input.Name != nil {
		name = *input.Name
	}
	if input.Description != nil {
		description = *input.Description
	}
	templateID := project.TemplateID
	if input.TemplateID != nil {
		templateID = *input.TemplateID
	}
	nextTemplate, err := domain.ValidateProjectInput(name, description, templateID)
	if err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	project.Name = strings.TrimSpace(name)
	project.Description = strings.TrimSpace(description)
	project.TemplateID = nextTemplate.ID
	project.Mode = nextTemplate.Mode
	project.UpdatedAt = s.now()

	if nextTemplate.ID == template.ID {
		if err := s.repo.Update(ctx, project); err != nil {
			return nil, fmt.Errorf("update project: %w", err)
		}
		return s.loadVisible(ctx, actor, project.ID)
	}

	// A template change rewrites every member's Project Role, so it runs under
	// the membership lock against freshly read members, like the other
	// membership mutations.
	err = s.repo.WithMembershipLock(ctx, project.ID, func(ctx context.Context) error {
		if err := s.refreshMembers(ctx, project); err != nil {
			return err
		}
		if err := requireProjectOwner(actor, project, template); err != nil {
			return err
		}
		if err := s.ensureTemplateChangeEligible(ctx, project); err != nil {
			return err
		}
		return s.changeTemplate(ctx, project, template, nextTemplate)
	})
	if err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	return s.loadVisible(ctx, actor, project.ID)
}

// changeTemplate moves the owners onto nextTemplate's creator role and clears
// every other member's Project Role (the templates' roles differ, so the owner
// reassigns them afterwards), then persists the project. Keycloak has no transactions, so when any step fails the roles
// already rewritten are put back (best effort, logged) and the original error
// is returned: the owner keeps the creator role of the template still stored.
func (s *ProjectConfigurationService) changeTemplate(ctx context.Context, project *domain.Project, current, next domain.Template) error {
	type change struct {
		userID   string
		previous domain.ProjectRole
	}
	var applied []change
	undo := func() {
		undoCtx := context.WithoutCancel(ctx)
		for i := len(applied) - 1; i >= 0; i-- {
			c := applied[i]
			if err := s.directory.SetMemberRole(undoCtx, project.ID, c.userID, c.previous); err != nil {
				slog.Error("undo template change: restore member role failed",
					"project_id", project.ID, "user_id", c.userID, "role", c.previous, "error", err)
			}
		}
	}
	for _, member := range project.Members {
		role := roleAfterTemplateChange(member.Role, current, next)
		if role == member.Role {
			continue
		}
		if err := s.directory.SetMemberRole(ctx, project.ID, member.UserID, role); err != nil {
			undo()
			return err
		}
		applied = append(applied, change{userID: member.UserID, previous: member.Role})
	}
	if err := s.repo.Update(ctx, project); err != nil {
		undo()
		return err
	}
	return nil
}

// roleAfterTemplateChange keeps owners as owners (the next template's creator
// role) and clears everyone else's role.
func roleAfterTemplateChange(role domain.ProjectRole, current, next domain.Template) domain.ProjectRole {
	if role == current.CreatorRole {
		return next.CreatorRole
	}
	return ""
}

// ensureTemplateChangeEligible refuses a template change while the Project has
// Sprints or Work Items: both are shaped by the template. Refusals are field
// violations so the owner is told what to remove, not a bare "conflict".
func (s *ProjectConfigurationService) ensureTemplateChangeEligible(ctx context.Context, project *domain.Project) error {
	if s.sprints == nil || s.workItems == nil {
		return fmt.Errorf("template change eligibility checks unavailable: %w", domain.ErrUnavailable)
	}
	sprintCount, err := s.sprints.CountByProject(ctx, project.ID)
	if err != nil {
		return err
	}
	if sprintCount > 0 {
		return templateChangeViolation(fmt.Sprintf(
			"This Project has %d %s. Delete %s before changing the template.",
			sprintCount, plural(sprintCount, "Sprint", "Sprints"), plural(sprintCount, "it", "them"),
		))
	}
	itemCount, err := s.workItems.CountByProject(ctx, project.ID)
	if err != nil {
		return errorsJoinUnavailable(err)
	}
	if itemCount > 0 {
		return templateChangeViolation(fmt.Sprintf(
			"This Project has %d %s. Delete %s before changing the template.",
			itemCount, plural(itemCount, "Work Item", "Work Items"), plural(itemCount, "it", "them"),
		))
	}
	return nil
}

func templateChangeViolation(message string) error {
	return &domain.InvalidInputError{Violations: []domain.Violation{{Field: "template_id", Message: message}}}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func (s *ProjectConfigurationService) DeleteProject(ctx context.Context, actor domain.Actor, projectID string) error {
	project, _, err := s.loadOwnerMutable(ctx, actor, projectID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if s.sprints == nil || s.workItems == nil {
		return fmt.Errorf("delete project: cleanup dependencies unavailable: %w", domain.ErrUnavailable)
	}
	// Every step is idempotent and the project row goes last, so a failure
	// part-way leaves a still-existing project whose delete can simply be
	// retried to finish the cleanup.
	if err := s.workItems.DeleteByProject(ctx, project.ID); err != nil {
		return fmt.Errorf("delete project: delete work items: %w", errorsJoinUnavailable(err))
	}
	if err := s.sprints.HardDeleteByProject(ctx, project.ID); err != nil {
		return fmt.Errorf("delete project: delete sprints: %w", err)
	}
	// Delete the identity group before the project row so a failed DB delete
	// remains retryable: group deletion is idempotent and the project still
	// exists for a later attempt.
	if err := s.directory.DeleteProjectGroup(ctx, project.GroupID); err != nil {
		return fmt.Errorf("delete project: delete membership group: %w", err)
	}
	if err := s.repo.Delete(ctx, project.ID); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

func errorsJoinUnavailable(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(domain.ErrUnavailable, err)
}

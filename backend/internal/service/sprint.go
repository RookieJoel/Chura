package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type SprintService struct {
	repository out.SprintRepository
	gateway    out.ProjectGateway
}

var _ in.SprintService = (*SprintService)(nil)

func NewSprintService(repository out.SprintRepository, gateway out.ProjectGateway) *SprintService {
	return &SprintService{
		repository: repository,
		gateway:    gateway,
	}
}

func (s *SprintService) CreateSprint(
	ctx context.Context,
	actor domain.Actor,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {

	if err := actor.RequireTeamMember(); err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	projectID, err := parseProjectID(sprint.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	if _, err := s.gateway.RetrieveProjectData(ctx, actor, projectID); err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	toCreate := *sprint
	toCreate.ProjectID = projectID
	if err := toCreate.Validate(); err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	created, err := s.repository.Create(ctx, &toCreate)
	if err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	return created, nil
}

func (s *SprintService) GetSprint(
	ctx context.Context,
	actor domain.Actor,
	id string,
) (*domain.Sprint, error) {

	sprint, err := s.loadVisible(ctx, actor, id)
	if err != nil {
		return nil, fmt.Errorf("get sprint: %w", err)
	}

	return sprint, nil
}

func (s *SprintService) ListSprints(
	ctx context.Context,
	actor domain.Actor,
	projectID string,
) ([]domain.Sprint, error) {

	if err := actor.Validate(); err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}

	projectID, err := parseProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}

	if _, err := s.gateway.RetrieveProjectData(ctx, actor, projectID); err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}

	sprints, err := s.repository.List(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}

	return sprints, nil
}

func (s *SprintService) UpdateSprint(
	ctx context.Context,
	actor domain.Actor,
	id string,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {

	if _, err := s.loadMutable(ctx, actor, id); err != nil {
		return nil, fmt.Errorf("update sprint: %w", err)
	}

	toUpdate := *sprint
	if err := toUpdate.Validate(); err != nil {
		return nil, fmt.Errorf("update sprint: %w", err)
	}

	updated, err := s.repository.Update(ctx, id, &toUpdate)
	if err != nil {
		return nil, fmt.Errorf("update sprint: %w", err)
	}

	return updated, nil
}

func (s *SprintService) DeleteSprint(
	ctx context.Context,
	actor domain.Actor,
	id string,
) error {

	if _, err := s.loadMutable(ctx, actor, id); err != nil {
		return fmt.Errorf("delete sprint: %w", err)
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete sprint: %w", err)
	}

	return nil
}

// loadVisible returns the sprint when the actor may see its Project; sprints
// of Projects the actor cannot see are reported as ErrNotFound.
func (s *SprintService) loadVisible(ctx context.Context, actor domain.Actor, id string) (*domain.Sprint, error) {
	if err := actor.Validate(); err != nil {
		return nil, err
	}

	sprint, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if _, err := s.gateway.RetrieveProjectData(ctx, actor, sprint.ProjectID); err != nil {
		return nil, err
	}

	return sprint, nil
}

// loadMutable loads a visible sprint and then requires a Team Member, matching
// the project gate order: a missing sprint is ErrNotFound for every actor and
// an Auditor on an existing sprint gets ErrForbidden.
func (s *SprintService) loadMutable(ctx context.Context, actor domain.Actor, id string) (*domain.Sprint, error) {
	sprint, err := s.loadVisible(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	if err := actor.RequireTeamMember(); err != nil {
		return nil, fmt.Errorf("sprint %s: %w", id, err)
	}

	return sprint, nil
}

// parseProjectID trims and requires a well-formed Project UUID.
func parseProjectID(raw string) (string, error) {
	projectID := strings.TrimSpace(raw)
	if projectID == "" {
		return "", &domain.InvalidInputError{Violations: []domain.Violation{
			{Field: "project_id", Message: "project_id is required"},
		}}
	}

	if _, err := uuid.Parse(projectID); err != nil {
		return "", fmt.Errorf("project id %q: %w", projectID, domain.ErrInvalidID)
	}

	return projectID, nil
}

package service_test

import (
	"context"
	"fmt"
	"strconv"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// fakeSprintRepo is an in-memory out.SprintRepository.
type fakeSprintRepo struct {
	sprints     []domain.Sprint
	createCalls int
	listFilter  *string // project filter received by the last List call
	updateCalls int
	deleteCalls int
	opErr       error // returned by Create, List, Update and Delete when set
}

func (r *fakeSprintRepo) Create(_ context.Context, s *domain.Sprint) (*domain.Sprint, error) {
	r.createCalls++
	if r.opErr != nil {
		return nil, r.opErr
	}
	c := *s
	c.ID = strconv.Itoa(len(r.sprints) + 1)
	r.sprints = append(r.sprints, c)
	return &c, nil
}

func (r *fakeSprintRepo) GetByID(_ context.Context, id string) (*domain.Sprint, error) {
	for _, s := range r.sprints {
		if s.ID == id {
			c := s
			return &c, nil
		}
	}
	return nil, fmt.Errorf("sprint %s: %w", id, domain.ErrNotFound)
}

func (r *fakeSprintRepo) List(_ context.Context, projectID string) ([]domain.Sprint, error) {
	r.listFilter = &projectID
	if r.opErr != nil {
		return nil, r.opErr
	}
	return r.sprints, nil
}

func (r *fakeSprintRepo) Update(_ context.Context, id string, s *domain.Sprint) (*domain.Sprint, error) {
	r.updateCalls++
	if r.opErr != nil {
		return nil, r.opErr
	}
	for i := range r.sprints {
		if r.sprints[i].ID == id {
			r.sprints[i].Name = s.Name
			c := r.sprints[i]
			return &c, nil
		}
	}
	return nil, fmt.Errorf("sprint %s: %w", id, domain.ErrNotFound)
}

func (r *fakeSprintRepo) Delete(_ context.Context, id string) error {
	r.deleteCalls++
	if r.opErr != nil {
		return r.opErr
	}
	for i := range r.sprints {
		if r.sprints[i].ID == id {
			r.sprints = append(r.sprints[:i], r.sprints[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("sprint %s: %w", id, domain.ErrNotFound)
}

// fakeGateway is an out.ProjectGateway returning canned results and recording calls.
type fakeGateway struct {
	project *domain.Project
	err     error
	calls   int
	gotID   string
	gotActr domain.Actor
}

func (g *fakeGateway) RetrieveProjectData(_ context.Context, a domain.Actor, projectID string) (*domain.Project, error) {
	g.calls++
	g.gotID, g.gotActr = projectID, a
	return g.project, g.err
}

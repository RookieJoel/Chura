package service_test

import (
	"context"
	"fmt"
	"sort"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// fakeProjectRepo is an in-memory out.ProjectRepository.
type fakeProjectRepo struct {
	projects map[string]domain.Project
	writes   int // number of mutating calls, to prove read-only operations persist nothing
	// beforeUpdate simulates a concurrent writer acting between the service's
	// read and the locked role update.
	beforeUpdate func(r *fakeProjectRepo)
}

func newFakeProjectRepo() *fakeProjectRepo {
	return &fakeProjectRepo{projects: map[string]domain.Project{}}
}

func (r *fakeProjectRepo) Create(_ context.Context, p *domain.Project) error {
	r.writes++
	r.projects[p.ID] = clone(*p)
	return nil
}

func (r *fakeProjectRepo) GetByID(_ context.Context, id string) (*domain.Project, error) {
	p, ok := r.projects[id]
	if !ok {
		return nil, fmt.Errorf("project %s: %w", id, domain.ErrNotFound)
	}
	c := clone(p)
	sort.SliceStable(c.Members, func(i, j int) bool { return c.Members[i].AddedAt.Before(c.Members[j].AddedAt) })
	return &c, nil
}

func (r *fakeProjectRepo) AddMember(_ context.Context, projectID string, m domain.Member) error {
	r.writes++
	p, ok := r.projects[projectID]
	if !ok {
		return fmt.Errorf("project %s: %w", projectID, domain.ErrNotFound)
	}
	for _, existing := range p.Members {
		if existing.UserID == m.UserID {
			return fmt.Errorf("member %s: %w", m.UserID, domain.ErrConflict)
		}
	}
	p.Members = append(p.Members, m)
	r.projects[projectID] = p
	return nil
}

// UpdateMemberRole mirrors the real repository: guard sees the current
// members (as if locked) and vetoes the change by returning an error.
func (r *fakeProjectRepo) UpdateMemberRole(_ context.Context, projectID, userID string, role domain.ProjectRole, guard func([]domain.Member) error) error {
	r.writes++
	if r.beforeUpdate != nil {
		r.beforeUpdate(r)
	}
	p, ok := r.projects[projectID]
	if !ok {
		return fmt.Errorf("project %s: %w", projectID, domain.ErrNotFound)
	}
	if err := guard(clone(p).Members); err != nil {
		return err
	}
	for i := range p.Members {
		if p.Members[i].UserID == userID {
			p.Members[i].Role = role
			return nil
		}
	}
	return fmt.Errorf("member %s: %w", userID, domain.ErrNotFound)
}

func clone(p domain.Project) domain.Project {
	p.Members = append([]domain.Member(nil), p.Members...)
	return p
}

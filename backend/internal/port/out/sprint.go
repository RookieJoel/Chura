package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// SprintRepository returns domain.ErrNotFound from GetByID, Update and Delete
// when the sprint does not exist. List returns only the sprints of projectID,
// which callers must always supply (non-empty, already validated).
type SprintRepository interface {
	Create(ctx context.Context, sprint *domain.Sprint) (*domain.Sprint, error)
	GetByID(ctx context.Context, id string) (*domain.Sprint, error)
	List(ctx context.Context, projectID string) ([]domain.Sprint, error)
	Update(ctx context.Context, id string, sprint *domain.Sprint) (*domain.Sprint, error)
	Delete(ctx context.Context, id string) error
}

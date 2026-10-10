package in

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// SprintService applies Project visibility rules to every Sprint operation:
// reads need Project membership (or the Auditor role); writes additionally
// need the Team Member role. Sprints of invisible Projects are ErrNotFound.
type SprintService interface {
	CreateSprint(ctx context.Context, actor domain.Actor, sprint *domain.Sprint) (*domain.Sprint, error)
	GetSprint(ctx context.Context, actor domain.Actor, id string) (*domain.Sprint, error)
	// ListSprints requires projectID (a UUID).
	ListSprints(ctx context.Context, actor domain.Actor, projectID string) ([]domain.Sprint, error)
	UpdateSprint(ctx context.Context, actor domain.Actor, id string, sprint *domain.Sprint) (*domain.Sprint, error)
	DeleteSprint(ctx context.Context, actor domain.Actor, id string) error
}

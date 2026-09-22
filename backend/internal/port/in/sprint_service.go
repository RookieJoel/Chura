package in

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type SprintService interface {
	CreateSprint(ctx context.Context, sprint *domain.Sprint) (*domain.Sprint, error)
	GetSprint(ctx context.Context, id string) (*domain.Sprint, error)
	ListSprints(ctx context.Context) ([]domain.Sprint, error)
	UpdateSprint(ctx context.Context, id string, sprint *domain.Sprint) (*domain.Sprint, error)
	DeleteSprint(ctx context.Context, id string) error
}

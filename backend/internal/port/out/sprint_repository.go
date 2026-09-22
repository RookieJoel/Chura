package out

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type SprintRepository interface {
	Create(ctx context.Context, sprint *domain.Sprint) (*domain.Sprint, error)
	GetByID(ctx context.Context, id string) (*domain.Sprint, error)
	List(ctx context.Context) ([]domain.Sprint, error)
	Update(ctx context.Context, id string, sprint *domain.Sprint) (*domain.Sprint, error)
	Delete(ctx context.Context, id string) error
}

package service

import (
	"context"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type SprintService struct {
	repository out.SprintRepository
}

func NewSprintService(repository out.SprintRepository) *SprintService {
	return &SprintService{
		repository: repository,
	}
}

func (s *SprintService) CreateSprint(
	ctx context.Context,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {
	return s.repository.Create(ctx, sprint)
}

func (s *SprintService) GetSprint(
	ctx context.Context,
	id string,
) (*domain.Sprint, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *SprintService) ListSprints(
	ctx context.Context,
) ([]domain.Sprint, error) {
	return s.repository.List(ctx)
}

func (s *SprintService) UpdateSprint(
	ctx context.Context,
	id string,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {
	return s.repository.Update(ctx, id, sprint)
}

func (s *SprintService) DeleteSprint(
	ctx context.Context,
	id string,
) error {
	return s.repository.Delete(ctx, id)
}

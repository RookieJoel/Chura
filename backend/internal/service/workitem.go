package service

import (
	"errors"
	"strings"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/driven"
	"github.com/google/uuid"
)

var ErrWorkItemNotFound = domain.ErrNotFound

type WorkItemService struct{ repository driven.WorkItemRepository }

func NewWorkItemService(repository driven.WorkItemRepository) *WorkItemService {
	return &WorkItemService{repository: repository}
}

func (service *WorkItemService) CreateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	if err := item.Validate(); err != nil {
		return domain.WorkItem{}, err
	}
	now := time.Now().UTC()
	item.ID, item.CreatedAt, item.UpdatedAt = uuid.NewString(), now, now
	return service.repository.Create(item)
}

func (service *WorkItemService) GetWorkItem(id string) (domain.WorkItem, error) {
	if strings.TrimSpace(id) == "" {
		return domain.WorkItem{}, errors.New("id is required")
	}
	return service.repository.Get(id)
}

func (service *WorkItemService) ListWorkItems(projectID string) ([]domain.WorkItem, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, errors.New("project_id is required")
	}
	return service.repository.List(projectID)
}

func (service *WorkItemService) UpdateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	if strings.TrimSpace(item.ID) == "" {
		return domain.WorkItem{}, errors.New("id is required")
	}
	if err := item.Validate(); err != nil {
		return domain.WorkItem{}, err
	}
	existing, err := service.repository.Get(item.ID)
	if err != nil {
		return domain.WorkItem{}, err
	}
	item.CreatedAt = existing.CreatedAt
	item.UpdatedAt = time.Now().UTC()
	return service.repository.Update(item)
}

func (service *WorkItemService) DeleteWorkItem(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}
	return service.repository.Delete(id)
}

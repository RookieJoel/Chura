package repository

import (
	"sync"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/driven"
)

var _ driven.WorkItemRepository = (*MemoryWorkItemRepository)(nil)

type MemoryWorkItemRepository struct {
	mu    sync.RWMutex
	items map[string]domain.WorkItem
}

func NewMemoryWorkItemRepository() *MemoryWorkItemRepository {
	return &MemoryWorkItemRepository{items: make(map[string]domain.WorkItem)}
}

func (repository *MemoryWorkItemRepository) Create(item domain.WorkItem) (domain.WorkItem, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.items[item.ID] = item
	return item, nil
}

func (repository *MemoryWorkItemRepository) Get(id string) (domain.WorkItem, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	item, ok := repository.items[id]
	if !ok {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (repository *MemoryWorkItemRepository) List(projectID string) ([]domain.WorkItem, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	items := make([]domain.WorkItem, 0)
	for _, item := range repository.items {
		if item.ProjectID == projectID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (repository *MemoryWorkItemRepository) Update(item domain.WorkItem) (domain.WorkItem, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, ok := repository.items[item.ID]; !ok {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	repository.items[item.ID] = item
	return item, nil
}

func (repository *MemoryWorkItemRepository) Delete(id string) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, ok := repository.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(repository.items, id)
	return nil
}

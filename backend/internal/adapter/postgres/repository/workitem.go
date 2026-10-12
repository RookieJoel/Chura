package memory

import (
	"sync"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type WorkItemRepository struct {
	mu    sync.RWMutex
	items map[string]domain.WorkItem
}

func NewWorkItemRepository() *WorkItemRepository {
	return &WorkItemRepository{items: make(map[string]domain.WorkItem)}
}

func (repository *WorkItemRepository) Create(item domain.WorkItem) (domain.WorkItem, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.items[item.ID] = item
	return item, nil
}

func (repository *WorkItemRepository) Get(id string) (domain.WorkItem, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	item, ok := repository.items[id]
	if !ok {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (repository *WorkItemRepository) List(filter out.WorkItemFilter) ([]domain.WorkItem, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	items := make([]domain.WorkItem, 0)
	for _, item := range repository.items {
		if matchesFilter(item, filter) {
			items = append(items, item)
		}
	}
	return items, nil
}

func matchesFilter(item domain.WorkItem, filter out.WorkItemFilter) bool {
	return (filter.ProjectID == "" || item.ProjectID == filter.ProjectID) &&
		(filter.SprintID == "" || item.SprintID == filter.SprintID)
}

func (repository *WorkItemRepository) Update(item domain.WorkItem) (domain.WorkItem, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, ok := repository.items[item.ID]; !ok {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	repository.items[item.ID] = item
	return item, nil
}

func (repository *WorkItemRepository) Delete(id string) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, ok := repository.items[id]; !ok {
		return domain.ErrNotFound
	}
	delete(repository.items, id)
	return nil
}

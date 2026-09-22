package out

import (
	"errors"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

var ErrNotFound = errors.New("work item not found")

type WorkItemRepository interface {
	Create(item domain.WorkItem) (domain.WorkItem, error)
	Get(id string) (domain.WorkItem, error)
	List(projectID string) ([]domain.WorkItem, error)
	Update(item domain.WorkItem) (domain.WorkItem, error)
	Delete(id string) error
}
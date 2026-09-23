package driven

import (
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type WorkItemRepository interface {
	Create(item domain.WorkItem) (domain.WorkItem, error)
	Get(id string) (domain.WorkItem, error)
	List(projectID string) ([]domain.WorkItem, error)
	Update(item domain.WorkItem) (domain.WorkItem, error)
	Delete(id string) error
}

type WorkItemGateway interface {
	CreateWorkItem(item domain.WorkItem) (domain.WorkItem, error)
	GetWorkItem(id string) (domain.WorkItem, error)
	ListWorkItems(projectID string) ([]domain.WorkItem, error)
	UpdateWorkItem(item domain.WorkItem) (domain.WorkItem, error)
	DeleteWorkItem(id string) error
}

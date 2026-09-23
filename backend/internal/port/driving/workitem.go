package driving

import (
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type WorkItemService interface {
	CreateWorkItem(input domain.WorkItem) (domain.WorkItem, error)
	GetWorkItem(id string) (domain.WorkItem, error)
	ListWorkItems(projectID string) ([]domain.WorkItem, error)
	UpdateWorkItem(item domain.WorkItem) (domain.WorkItem, error)
	DeleteWorkItem(id string) error
}

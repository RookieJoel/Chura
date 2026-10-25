package grpc

import (
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

// LocalWorkItemGateway implements out.WorkItemGateway by calling WorkItemService in-process.
type LocalWorkItemGateway struct {
	service in.WorkItemService
}

func NewLocalWorkItemGateway(service in.WorkItemService) *LocalWorkItemGateway {
	return &LocalWorkItemGateway{service: service}
}

func (gateway *LocalWorkItemGateway) CreateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	return gateway.service.CreateWorkItem(item)
}

func (gateway *LocalWorkItemGateway) GetWorkItem(id string) (domain.WorkItem, error) {
	return gateway.service.GetWorkItem(id)
}

func (gateway *LocalWorkItemGateway) ListWorkItems(projectID string) ([]domain.WorkItem, error) {
	return gateway.service.ListWorkItems(projectID)
}

func (gateway *LocalWorkItemGateway) UpdateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	return gateway.service.UpdateWorkItem(item)
}

func (gateway *LocalWorkItemGateway) DeleteWorkItem(id string) error {
	return gateway.service.DeleteWorkItem(id)
}

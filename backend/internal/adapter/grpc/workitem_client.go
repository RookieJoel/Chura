package grpc

import (
	"context"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/adapter/grpc/pb/workitem"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type WorkItemGateway struct {
	client workitem.WorkItemServiceClient
}

func NewWorkItemGateway(connection grpc.ClientConnInterface) *WorkItemGateway {
	return &WorkItemGateway{client: workitem.NewWorkItemServiceClient(connection)}
}

func (gateway *WorkItemGateway) CreateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	stream, err := gateway.client.CreateWorkItems(context.Background())
	if err != nil {
		return domain.WorkItem{}, err
	}
	if err := stream.Send(&workitem.CreateWorkItemRequest{WorkItem: workItemInputFromDomain(item)}); err != nil {
		return domain.WorkItem{}, err
	}
	if err := stream.CloseSend(); err != nil {
		return domain.WorkItem{}, err
	}
	response, err := stream.Recv()
	if err != nil {
		return domain.WorkItem{}, err
	}
	return workItemFromProto(response), nil
}

func (gateway *WorkItemGateway) GetWorkItem(id string) (domain.WorkItem, error) {
	response, err := gateway.client.GetWorkItem(context.Background(), &workitem.GetWorkItemRequest{Id: id})
	if err != nil {
		return domain.WorkItem{}, err
	}
	return workItemFromProto(response), nil
}

func (gateway *WorkItemGateway) ListWorkItems(projectID string) ([]domain.WorkItem, error) {
	response, err := gateway.client.ListWorkItems(context.Background(), &workitem.ListWorkItemsRequest{ProjectId: projectID})
	if err != nil {
		return nil, err
	}
	items := make([]domain.WorkItem, 0, len(response.GetWorkItems()))
	for _, item := range response.GetWorkItems() {
		items = append(items, workItemFromProto(item))
	}
	return items, nil
}

func (gateway *WorkItemGateway) UpdateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	response, err := gateway.client.UpdateWorkItem(context.Background(), &workitem.UpdateWorkItemRequest{
		Id:       item.ID,
		WorkItem: workItemInputFromDomain(item),
	})
	if err != nil {
		return domain.WorkItem{}, err
	}
	return workItemFromProto(response), nil
}

func (gateway *WorkItemGateway) DeleteWorkItem(id string) error {
	_, err := gateway.client.DeleteWorkItem(context.Background(), &workitem.DeleteWorkItemRequest{Id: id})
	return err
}

func workItemInputFromDomain(item domain.WorkItem) *workitem.WorkItemInput {
	var features *structpb.Struct
	if item.Features != nil {
		features, _ = structpb.NewStruct(item.Features)
	}
	reporterID := ""
	if len(item.ReporterIDs) > 0 {
		reporterID = item.ReporterIDs[0]
	}
	return &workitem.WorkItemInput{
		ProjectId:   item.ProjectID,
		Title:       item.Title,
		Description: item.Description,
		Type:        typeValue(item.Type),
		Status:      statusValue(item.Status),
		Priority:    priorityValue(item.Priority),
		AssigneeId:  item.AssigneeID,
		ReporterId:  reporterID,
		StoryPoints: item.StoryPoints,
		Features:    features,
	}
}

func workItemFromProto(item *workitem.WorkItem) domain.WorkItem {
	var features map[string]any
	if item.GetFeatures() != nil {
		features = item.GetFeatures().AsMap()
	}
	reporterIDs := []string(nil)
	if item.GetReporterId() != "" {
		reporterIDs = []string{item.GetReporterId()}
	}
	return domain.WorkItem{
		ID:          item.GetId(),
		ProjectID:   item.GetProjectId(),
		Title:       item.GetTitle(),
		Description: item.GetDescription(),
		Type:        domain.WorkItemType(typeName(item.GetType())),
		Status:      domain.WorkItemStatus(statusName(item.GetStatus())),
		Priority:    domain.WorkItemPriority(priorityName(item.GetPriority())),
		AssigneeID:  item.GetAssigneeId(),
		ReporterIDs: reporterIDs,
		StoryPoints: item.GetStoryPoints(),
		Features:    features,
		CreatedAt:   timestampFromProto(item.GetCreatedAt()),
		UpdatedAt:   timestampFromProto(item.GetUpdatedAt()),
	}
}

func timestampFromProto(timestamp *timestamppb.Timestamp) time.Time {
	if timestamp == nil {
		return time.Time{}
	}
	return timestamp.AsTime()
}

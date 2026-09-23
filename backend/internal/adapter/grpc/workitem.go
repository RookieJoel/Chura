package grpc

import (
	"context"
	"errors"
	"io"

	"github.com/RookieJoel/Chura/backend/internal/adapter/grpc/pb/workitem"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/driving"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	workitem.UnimplementedWorkItemServiceServer
	service driving.WorkItemService
}

func NewServer(service driving.WorkItemService) *Server {
	return &Server{service: service}
}

func (server *Server) CreateWorkItems(stream workitem.WorkItemService_CreateWorkItemsServer) error {
	for {
		request, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return status.Error(codes.Unknown, err.Error())
		}

		item, err := server.service.CreateWorkItem(workItemFromInput(request.GetWorkItem()))
		if err != nil {
			return grpcError(err)
		}
		response, err := workItemToProto(item)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (server *Server) GetWorkItem(_ context.Context, request *workitem.GetWorkItemRequest) (*workitem.WorkItem, error) {
	item, err := server.service.GetWorkItem(request.GetId())
	if err != nil {
		return nil, grpcError(err)
	}
	return workItemToProto(item)
}

func (server *Server) ListWorkItems(_ context.Context, request *workitem.ListWorkItemsRequest) (*workitem.ListWorkItemsResponse, error) {
	items, err := server.service.ListWorkItems(request.GetProjectId())
	if err != nil {
		return nil, grpcError(err)
	}

	response := &workitem.ListWorkItemsResponse{WorkItems: make([]*workitem.WorkItem, 0, len(items))}
	for _, item := range items {
		workItem, err := workItemToProto(item)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		response.WorkItems = append(response.WorkItems, workItem)
	}
	return response, nil
}

func (server *Server) UpdateWorkItem(_ context.Context, request *workitem.UpdateWorkItemRequest) (*workitem.WorkItem, error) {
	item := workItemFromInput(request.GetWorkItem())
	item.ID = request.GetId()
	updated, err := server.service.UpdateWorkItem(item)
	if err != nil {
		return nil, grpcError(err)
	}
	return workItemToProto(updated)
}

func (server *Server) DeleteWorkItem(_ context.Context, request *workitem.DeleteWorkItemRequest) (*emptypb.Empty, error) {
	if err := server.service.DeleteWorkItem(request.GetId()); err != nil {
		return nil, grpcError(err)
	}
	return &emptypb.Empty{}, nil
}

func workItemFromInput(input *workitem.WorkItemInput) domain.WorkItem {
	if input == nil {
		return domain.WorkItem{}
	}

	features := map[string]any(nil)
	if input.GetFeatures() != nil {
		features = input.GetFeatures().AsMap()
	}
	reporterIDs := []string(nil)
	if input.GetReporterId() != "" {
		reporterIDs = []string{input.GetReporterId()}
	}
	return domain.WorkItem{
		ProjectID:   input.GetProjectId(),
		Title:       input.GetTitle(),
		Description: input.GetDescription(),
		Type:        domain.WorkItemType(typeName(input.GetType())),
		Status:      domain.WorkItemStatus(statusName(input.GetStatus())),
		Priority:    domain.WorkItemPriority(priorityName(input.GetPriority())),
		AssigneeID:  input.GetAssigneeId(),
		ReporterIDs: reporterIDs,
		StoryPoints: input.GetStoryPoints(),
		Features:    features,
	}
}

func workItemToProto(item domain.WorkItem) (*workitem.WorkItem, error) {
	var features *structpb.Struct
	var err error
	if item.Features != nil {
		features, err = structpb.NewStruct(item.Features)
		if err != nil {
			return nil, err
		}
	}
	reporterID := ""
	if len(item.ReporterIDs) > 0 {
		reporterID = item.ReporterIDs[0]
	}
	return &workitem.WorkItem{
		Id:          item.ID,
		ProjectId:   item.ProjectID,
		Title:       item.Title,
		Description: item.Description,
		Type:        typeValue(item.Type),
		Status:      statusValue(item.Status),
		Priority:    priorityValue(item.Priority),
		AssigneeId:  item.AssigneeID,
		ReporterId:  reporterID,
		StoryPoints: item.StoryPoints,
		CreatedAt:   timestamppb.New(item.CreatedAt),
		UpdatedAt:   timestamppb.New(item.UpdatedAt),
		Features:    features,
	}, nil
}

func grpcError(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.InvalidArgument, err.Error())
}

func typeName(value workitem.WorkItemType) string {
	switch value {
	case workitem.WorkItemType_TASK:
		return string(domain.WorkItemTypeTask)
	case workitem.WorkItemType_USER_STORY:
		return string(domain.WorkItemTypeUserStory)
	case workitem.WorkItemType_BUG:
		return string(domain.WorkItemTypeBug)
	default:
		return ""
	}
}

func statusName(value workitem.WorkItemStatus) string {
	switch value {
	case workitem.WorkItemStatus_TO_DO:
		return string(domain.WorkItemStatusToDo)
	case workitem.WorkItemStatus_IN_PROGRESS:
		return string(domain.WorkItemStatusInProgress)
	case workitem.WorkItemStatus_REVIEW:
		return string(domain.WorkItemStatusReview)
	case workitem.WorkItemStatus_DONE:
		return string(domain.WorkItemStatusDone)
	case workitem.WorkItemStatus_BLOCKED:
		return string(domain.WorkItemStatusBlocked)
	default:
		return ""
	}
}

func priorityName(value workitem.WorkItemPriority) string {
	switch value {
	case workitem.WorkItemPriority_LOW:
		return string(domain.WorkItemPriorityLow)
	case workitem.WorkItemPriority_MEDIUM:
		return string(domain.WorkItemPriorityMedium)
	case workitem.WorkItemPriority_HIGH:
		return string(domain.WorkItemPriorityHigh)
	case workitem.WorkItemPriority_CRITICAL:
		return string(domain.WorkItemPriorityCritical)
	default:
		return ""
	}
}

func typeValue(value domain.WorkItemType) workitem.WorkItemType {
	switch value {
	case domain.WorkItemTypeTask:
		return workitem.WorkItemType_TASK
	case domain.WorkItemTypeUserStory:
		return workitem.WorkItemType_USER_STORY
	case domain.WorkItemTypeBug:
		return workitem.WorkItemType_BUG
	default:
		return workitem.WorkItemType_WORK_ITEM_TYPE_UNSPECIFIED
	}
}

func statusValue(value domain.WorkItemStatus) workitem.WorkItemStatus {
	switch value {
	case domain.WorkItemStatusToDo:
		return workitem.WorkItemStatus_TO_DO
	case domain.WorkItemStatusInProgress:
		return workitem.WorkItemStatus_IN_PROGRESS
	case domain.WorkItemStatusReview:
		return workitem.WorkItemStatus_REVIEW
	case domain.WorkItemStatusDone:
		return workitem.WorkItemStatus_DONE
	case domain.WorkItemStatusBlocked:
		return workitem.WorkItemStatus_BLOCKED
	default:
		return workitem.WorkItemStatus_WORK_ITEM_STATUS_UNSPECIFIED
	}
}

func priorityValue(value domain.WorkItemPriority) workitem.WorkItemPriority {
	switch value {
	case domain.WorkItemPriorityLow:
		return workitem.WorkItemPriority_LOW
	case domain.WorkItemPriorityMedium:
		return workitem.WorkItemPriority_MEDIUM
	case domain.WorkItemPriorityHigh:
		return workitem.WorkItemPriority_HIGH
	case domain.WorkItemPriorityCritical:
		return workitem.WorkItemPriority_CRITICAL
	default:
		return workitem.WorkItemPriority_WORK_ITEM_PRIORITY_UNSPECIFIED
	}
}

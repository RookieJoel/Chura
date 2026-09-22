package workitem

import (
	"context"
	"errors"
	"io"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	UnimplementedWorkItemServiceServer
	service in.WorkItemService
}

func NewServer(service in.WorkItemService) *Server {
	return &Server{service: service}
}

func (server *Server) CreateWorkItems(stream WorkItemService_CreateWorkItemsServer) error {
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

func (server *Server) GetWorkItem(_ context.Context, request *GetWorkItemRequest) (*WorkItem, error) {
	item, err := server.service.GetWorkItem(request.GetId())
	if err != nil {
		return nil, grpcError(err)
	}
	return workItemToProto(item)
}

func (server *Server) ListWorkItems(_ context.Context, request *ListWorkItemsRequest) (*ListWorkItemsResponse, error) {
	items, err := server.service.ListWorkItems(request.GetProjectId())
	if err != nil {
		return nil, grpcError(err)
	}

	response := &ListWorkItemsResponse{WorkItems: make([]*WorkItem, 0, len(items))}
	for _, item := range items {
		workItem, err := workItemToProto(item)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		response.WorkItems = append(response.WorkItems, workItem)
	}
	return response, nil
}

func (server *Server) UpdateWorkItem(_ context.Context, request *UpdateWorkItemRequest) (*WorkItem, error) {
	item := workItemFromInput(request.GetWorkItem())
	item.ID = request.GetId()
	updated, err := server.service.UpdateWorkItem(item)
	if err != nil {
		return nil, grpcError(err)
	}
	return workItemToProto(updated)
}

func (server *Server) DeleteWorkItem(_ context.Context, request *DeleteWorkItemRequest) (*emptypb.Empty, error) {
	if err := server.service.DeleteWorkItem(request.GetId()); err != nil {
		return nil, grpcError(err)
	}
	return &emptypb.Empty{}, nil
}

func workItemFromInput(input *WorkItemInput) domain.WorkItem {
	if input == nil {
		return domain.WorkItem{}
	}

	features := map[string]any(nil)
	if input.GetFeatures() != nil {
		features = input.GetFeatures().AsMap()
	}
	return domain.WorkItem{
		ProjectID:   input.GetProjectId(),
		Title:       input.GetTitle(),
		Description: input.GetDescription(),
		Type:        domain.WorkItemType(typeName(input.GetType())),
		Status:      domain.WorkItemStatus(statusName(input.GetStatus())),
		Priority:    domain.WorkItemPriority(priorityName(input.GetPriority())),
		AssigneeID:  input.GetAssigneeId(),
		ReporterID:  input.GetReporterId(),
		StoryPoints: input.GetStoryPoints(),
		Features:    features,
	}
}

func workItemToProto(item domain.WorkItem) (*WorkItem, error) {
	var features *structpb.Struct
	var err error
	if item.Features != nil {
		features, err = structpb.NewStruct(item.Features)
		if err != nil {
			return nil, err
		}
	}
	return &WorkItem{
		Id:          item.ID,
		ProjectId:   item.ProjectID,
		Title:       item.Title,
		Description: item.Description,
		Type:        typeValue(item.Type),
		Status:      statusValue(item.Status),
		Priority:    priorityValue(item.Priority),
		AssigneeId:  item.AssigneeID,
		ReporterId:  item.ReporterID,
		StoryPoints: item.StoryPoints,
		CreatedAt:   timestamppb.New(item.CreatedAt),
		UpdatedAt:   timestamppb.New(item.UpdatedAt),
		Features:    features,
	}, nil
}

func grpcError(err error) error {
	if errors.Is(err, out.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.InvalidArgument, err.Error())
}

func typeName(value WorkItemType) string {
	switch value {
	case WorkItemType_TASK:
		return string(domain.WorkItemTypeTask)
	case WorkItemType_USER_STORY:
		return string(domain.WorkItemTypeUserStory)
	case WorkItemType_BUG:
		return string(domain.WorkItemTypeBug)
	default:
		return ""
	}
}

func statusName(value WorkItemStatus) string {
	switch value {
	case WorkItemStatus_TO_DO:
		return string(domain.WorkItemStatusToDo)
	case WorkItemStatus_IN_PROGRESS:
		return string(domain.WorkItemStatusInProgress)
	case WorkItemStatus_REVIEW:
		return string(domain.WorkItemStatusReview)
	case WorkItemStatus_DONE:
		return string(domain.WorkItemStatusDone)
	case WorkItemStatus_BLOCKED:
		return string(domain.WorkItemStatusBlocked)
	default:
		return ""
	}
}

func priorityName(value WorkItemPriority) string {
	switch value {
	case WorkItemPriority_LOW:
		return string(domain.WorkItemPriorityLow)
	case WorkItemPriority_MEDIUM:
		return string(domain.WorkItemPriorityMedium)
	case WorkItemPriority_HIGH:
		return string(domain.WorkItemPriorityHigh)
	case WorkItemPriority_CRITICAL:
		return string(domain.WorkItemPriorityCritical)
	default:
		return ""
	}
}

func typeValue(value domain.WorkItemType) WorkItemType {
	switch value {
	case domain.WorkItemTypeTask:
		return WorkItemType_TASK
	case domain.WorkItemTypeUserStory:
		return WorkItemType_USER_STORY
	case domain.WorkItemTypeBug:
		return WorkItemType_BUG
	default:
		return WorkItemType_WORK_ITEM_TYPE_UNSPECIFIED
	}
}

func statusValue(value domain.WorkItemStatus) WorkItemStatus {
	switch value {
	case domain.WorkItemStatusToDo:
		return WorkItemStatus_TO_DO
	case domain.WorkItemStatusInProgress:
		return WorkItemStatus_IN_PROGRESS
	case domain.WorkItemStatusReview:
		return WorkItemStatus_REVIEW
	case domain.WorkItemStatusDone:
		return WorkItemStatus_DONE
	case domain.WorkItemStatusBlocked:
		return WorkItemStatus_BLOCKED
	default:
		return WorkItemStatus_WORK_ITEM_STATUS_UNSPECIFIED
	}
}

func priorityValue(value domain.WorkItemPriority) WorkItemPriority {
	switch value {
	case domain.WorkItemPriorityLow:
		return WorkItemPriority_LOW
	case domain.WorkItemPriorityMedium:
		return WorkItemPriority_MEDIUM
	case domain.WorkItemPriorityHigh:
		return WorkItemPriority_HIGH
	case domain.WorkItemPriorityCritical:
		return WorkItemPriority_CRITICAL
	default:
		return WorkItemPriority_WORK_ITEM_PRIORITY_UNSPECIFIED
	}
}

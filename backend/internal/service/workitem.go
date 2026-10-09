package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/google/uuid"
)

var ErrWorkItemNotFound = domain.ErrNotFound

type WorkItemService struct {
	repository out.WorkItemRepository
	notifier   SprintNotifier
	publisher  out.EventPublisher
}

type SprintNotifier interface {
	SendSprintFinishedNotifications(sprintID string, reporterIDs []string) error
}

func NewWorkItemService(repository out.WorkItemRepository) *WorkItemService {
	return &WorkItemService{repository: repository}
}

func NewWorkItemServiceWithNotifier(
	repository out.WorkItemRepository,
	notifier SprintNotifier,
) *WorkItemService {
	return &WorkItemService{
		repository: repository,
		notifier:   notifier,
	}
}

func NewWorkItemServiceWithPublisher(
	repository out.WorkItemRepository,
	publisher out.EventPublisher,
) *WorkItemService {
	return &WorkItemService{repository: repository, publisher: publisher}
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

func (service *WorkItemService) GetWorkItemReporterID(id string) (string, error) {
	item, err := service.GetWorkItem(id)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(item.ReporterID) == "" {
		return "", errors.New("work item has no reporter")
	}
	return item.ReporterID, nil
}

func (service *WorkItemService) ListWorkItems(projectID string) ([]domain.WorkItem, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, errors.New("project_id is required")
	}
	return service.repository.List(projectID)
}

func (service *WorkItemService) ListWorkItemsBySprint(sprintID string) ([]domain.WorkItem, error) {
	if strings.TrimSpace(sprintID) == "" {
		return nil, errors.New("sprint_id is required")
	}
	return service.repository.ListBySprint(sprintID)
}

func (service *WorkItemService) CheckSprintFinished(sprintID string) (bool, error) {
	items, err := service.ListWorkItemsBySprint(sprintID)
	if err != nil {
		return false, err
	}

	reporterIDs := make([]string, 0, len(items))
	uniqueReporterIDs := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.Status != domain.WorkItemStatusDone {
			return false, nil
		}

		reporterID := strings.TrimSpace(item.ReporterID)
		if reporterID != "" {
			if _, exists := uniqueReporterIDs[reporterID]; !exists {
				uniqueReporterIDs[reporterID] = struct{}{}
				reporterIDs = append(reporterIDs, reporterID)
			}
		}
	}

	if len(reporterIDs) == 0 {
		return true, nil
	}
	if service.publisher != nil {
		payload, err := json.Marshal(struct {
			SprintID    string   `json:"sprint_id"`
			ReporterIDs []string `json:"reporter_ids"`
		}{sprintID, reporterIDs})
		if err != nil {
			return false, fmt.Errorf("encode sprint finished event: %w", err)
		}
		if err := service.publisher.Publish("sprint.finished", payload); err != nil {
			return false, fmt.Errorf("publish sprint finished event: %w", err)
		}
		return true, nil
	}
	if service.notifier == nil {
		return false, errors.New("sprint notifier or event publisher is not configured")
	}
	if err := service.notifier.SendSprintFinishedNotifications(sprintID, reporterIDs); err != nil {
		return false, fmt.Errorf("notify sprint reporters: %w", err)
	}
	return true, nil
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
	if strings.TrimSpace(item.SprintID) == "" {
		item.SprintID = existing.SprintID
	}
	item.UpdatedAt = time.Now().UTC()
	updated, err := service.repository.Update(item)
	if err != nil {
		return domain.WorkItem{}, err
	}
	if service.publisher != nil {
		if existing.Status != updated.Status {
			if err := service.publishWorkItemEvent("workitem.status_changed", updated, map[string]string{
				"old_status": string(existing.Status),
				"new_status": string(updated.Status),
			}); err != nil {
				return domain.WorkItem{}, err
			}
		}
		if existing.AssigneeID != updated.AssigneeID {
			eventType := "workitem.assigned"
			if strings.TrimSpace(updated.AssigneeID) == "" {
				eventType = "workitem.unassigned"
			}
			if err := service.publishWorkItemEvent(eventType, updated, map[string]string{
				"previous_assignee_id": existing.AssigneeID,
			}); err != nil {
				return domain.WorkItem{}, err
			}
		}
	}
	return updated, nil
}

func (service *WorkItemService) publishWorkItemEvent(eventType string, item domain.WorkItem, fields map[string]string) error {
	event := map[string]any{
		"event_type":   eventType,
		"work_item_id": item.ID,
		"title":        item.Title,
		"reporter_id":  item.ReporterID,
		"assignee_id":  item.AssigneeID,
	}
	for key, value := range fields {
		event[key] = value
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode %s event: %w", eventType, err)
	}
	if err := service.publisher.Publish(eventType, payload); err != nil {
		return fmt.Errorf("publish %s event: %w", eventType, err)
	}
	return nil
}

func (service *WorkItemService) DeleteWorkItem(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}
	return service.repository.Delete(id)
}

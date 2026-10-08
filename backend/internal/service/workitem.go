package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/driven"
	"github.com/google/uuid"
)

var ErrWorkItemNotFound = domain.ErrNotFound

type WorkItemService struct {
	repository driven.WorkItemRepository
	notifier   SprintNotifier
}

type SprintNotifier interface {
	SendSprintFinishedNotifications(sprintID string, reporterIDs []string) error
}

func NewWorkItemService(repository driven.WorkItemRepository) *WorkItemService {
	return &WorkItemService{repository: repository}
}

func NewWorkItemServiceWithNotifier(
	repository driven.WorkItemRepository,
	notifier SprintNotifier,
) *WorkItemService {
	return &WorkItemService{
		repository: repository,
		notifier:   notifier,
	}
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
	if service.notifier == nil {
		return false, errors.New("sprint notifier is not configured")
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
	return service.repository.Update(item)
}

func (service *WorkItemService) DeleteWorkItem(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}
	return service.repository.Delete(id)
}

package domain

import (
	"errors"
	"strings"
	"time"
)

type WorkItemType string
type WorkItemStatus string
type WorkItemPriority string

const (
	WorkItemTypeTask      WorkItemType = "task"
	WorkItemTypeUserStory WorkItemType = "user_story"
	WorkItemTypeBug       WorkItemType = "bug"

	WorkItemStatusToDo       WorkItemStatus = "to_do"
	WorkItemStatusInProgress WorkItemStatus = "in_progress"
	WorkItemStatusReview     WorkItemStatus = "review"
	WorkItemStatusDone       WorkItemStatus = "done"
	WorkItemStatusBlocked    WorkItemStatus = "blocked"

	WorkItemPriorityLow      WorkItemPriority = "low"
	WorkItemPriorityMedium   WorkItemPriority = "medium"
	WorkItemPriorityHigh     WorkItemPriority = "high"
	WorkItemPriorityCritical WorkItemPriority = "critical"
)

type WorkItem struct {
	ID          string
	ProjectID   string
	Title       string
	Description string
	Type        WorkItemType
	Status      WorkItemStatus
	Priority    WorkItemPriority
	AssigneeID  string
	ReporterIDs []string
	StoryPoints int32
	Features    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (item WorkItem) Validate() error {
	if strings.TrimSpace(item.ProjectID) == "" {
		return errors.New("project_id is required")
	}
	if strings.TrimSpace(item.Title) == "" {
		return errors.New("title is required")
	}
	if item.Type == "" {
		return errors.New("type is required")
	}
	if item.Status == "" {
		return errors.New("status is required")
	}
	if item.Priority == "" {
		return errors.New("priority is required")
	}
	if item.StoryPoints < 0 {
		return errors.New("story_points cannot be negative")
	}
	return nil
}

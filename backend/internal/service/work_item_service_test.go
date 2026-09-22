package service_test

import (
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/adapter/memory"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func TestWorkItemServiceCreateReturnsGeneratedID(t *testing.T) {
	workItemService := service.NewWorkItemService(memory.NewWorkItemRepository())

	item, err := workItemService.CreateWorkItem(validWorkItem())
	if err != nil {
		t.Fatalf("create work item: %v", err)
	}
	if item.ID == "" {
		t.Fatal("expected generated work item ID")
	}
	if item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() {
		t.Fatal("expected creation timestamps")
	}
}

func TestWorkItemServiceCRUD(t *testing.T) {
	workItemService := service.NewWorkItemService(memory.NewWorkItemRepository())
	created, err := workItemService.CreateWorkItem(validWorkItem())
	if err != nil {
		t.Fatalf("create work item: %v", err)
	}

	created.Title = "Updated title"
	updated, err := workItemService.UpdateWorkItem(created)
	if err != nil {
		t.Fatalf("update work item: %v", err)
	}
	if updated.Title != "Updated title" {
		t.Fatalf("got title %q", updated.Title)
	}

	items, err := workItemService.ListWorkItems(created.ProjectID)
	if err != nil || len(items) != 1 {
		t.Fatalf("list work items: got %d items, error %v", len(items), err)
	}
	if err := workItemService.DeleteWorkItem(created.ID); err != nil {
		t.Fatalf("delete work item: %v", err)
	}
	if _, err := workItemService.GetWorkItem(created.ID); err != service.ErrWorkItemNotFound {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

func TestWorkItemServiceRejectsInvalidInput(t *testing.T) {
	workItemService := service.NewWorkItemService(memory.NewWorkItemRepository())
	if _, err := workItemService.CreateWorkItem(domain.WorkItem{}); err == nil {
		t.Fatal("expected invalid work item to be rejected")
	}
}

func validWorkItem() domain.WorkItem {
	return domain.WorkItem{
		ProjectID:   "project-1",
		Title:       "Implement backlog flow",
		Description: "Create a work item through the service",
		Type:        domain.WorkItemTypeTask,
		Status:      domain.WorkItemStatusToDo,
		Priority:    domain.WorkItemPriorityMedium,
		StoryPoints: 3,
	}
}

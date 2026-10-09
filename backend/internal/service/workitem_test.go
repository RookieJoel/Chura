package service_test

import (
	"encoding/json"
	"testing"

	memory "github.com/RookieJoel/Chura/backend/internal/adapter/postgres/repository"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func TestWorkItemServiceListsWorkItemsBySprint(t *testing.T) {
	workItemService := service.NewWorkItemService(memory.NewWorkItemRepository())

	first := validWorkItem()
	first.SprintID = "sprint-1"
	if _, err := workItemService.CreateWorkItem(first); err != nil {
		t.Fatalf("create first work item: %v", err)
	}

	second := validWorkItem()
	second.Title = "Unrelated work item"
	second.SprintID = "sprint-2"
	if _, err := workItemService.CreateWorkItem(second); err != nil {
		t.Fatalf("create second work item: %v", err)
	}

	items, err := workItemService.ListWorkItemsBySprint("sprint-1")
	if err != nil {
		t.Fatalf("list sprint work items: %v", err)
	}
	if len(items) != 1 || items[0].Title != first.Title {
		t.Fatalf("expected one item from sprint-1, got %#v", items)
	}
}

func TestWorkItemServiceRequiresSprintID(t *testing.T) {
	workItemService := service.NewWorkItemService(memory.NewWorkItemRepository())

	if _, err := workItemService.ListWorkItemsBySprint(" "); err == nil {
		t.Fatal("expected sprint_id validation error")
	}
}

func TestWorkItemServiceGetsReporterIDFromWorkItem(t *testing.T) {
	workItemService := service.NewWorkItemService(memory.NewWorkItemRepository())
	item := validWorkItem()
	item.ReporterID = "reporter-1"
	created, err := workItemService.CreateWorkItem(item)
	if err != nil {
		t.Fatalf("create work item: %v", err)
	}

	reporterID, err := workItemService.GetWorkItemReporterID(created.ID)
	if err != nil {
		t.Fatalf("get reporter ID: %v", err)
	}
	if reporterID != "reporter-1" {
		t.Fatalf("expected reporter-1, got %q", reporterID)
	}
}

func TestWorkItemServiceRejectsWorkItemWithoutReporter(t *testing.T) {
	workItemService := service.NewWorkItemService(memory.NewWorkItemRepository())
	created, err := workItemService.CreateWorkItem(validWorkItem())
	if err != nil {
		t.Fatalf("create work item: %v", err)
	}

	if _, err := workItemService.GetWorkItemReporterID(created.ID); err == nil {
		t.Fatal("expected missing reporter error")
	}
}

func TestWorkItemServiceNotifiesWhenSprintIsFinished(t *testing.T) {
	notifier := &sprintNotifier{}
	workItemService := service.NewWorkItemServiceWithNotifier(
		memory.NewWorkItemRepository(),
		notifier,
	)

	first := validWorkItem()
	first.SprintID = "sprint-1"
	first.Status = domain.WorkItemStatusDone
	first.ReporterID = "reporter-1"
	if _, err := workItemService.CreateWorkItem(first); err != nil {
		t.Fatalf("create first work item: %v", err)
	}

	second := validWorkItem()
	second.SprintID = "sprint-1"
	second.Status = domain.WorkItemStatusDone
	second.ReporterID = "reporter-1"
	if _, err := workItemService.CreateWorkItem(second); err != nil {
		t.Fatalf("create second work item: %v", err)
	}

	finished, err := workItemService.CheckSprintFinished("sprint-1")
	if err != nil {
		t.Fatalf("check sprint: %v", err)
	}
	if !finished {
		t.Fatal("expected sprint to be finished")
	}
	if len(notifier.calls) != 1 || notifier.calls[0].sprintID != "sprint-1" {
		t.Fatalf("expected one notification for sprint-1, got %#v", notifier.calls)
	}
	if len(notifier.calls[0].reporterIDs) != 1 || notifier.calls[0].reporterIDs[0] != "reporter-1" {
		t.Fatalf("expected one unique reporter, got %#v", notifier.calls[0].reporterIDs)
	}
}

func TestWorkItemServicePublishesSprintFinishedEvent(t *testing.T) {
	publisher := &eventPublisher{}
	workItemService := service.NewWorkItemServiceWithPublisher(
		memory.NewWorkItemRepository(),
		publisher,
	)
	item := validWorkItem()
	item.SprintID = "sprint-published"
	item.Status = domain.WorkItemStatusDone
	item.ReporterID = "reporter-1"
	if _, err := workItemService.CreateWorkItem(item); err != nil {
		t.Fatalf("create work item: %v", err)
	}

	finished, err := workItemService.CheckSprintFinished("sprint-published")
	if err != nil {
		t.Fatalf("check sprint: %v", err)
	}
	if !finished {
		t.Fatal("expected sprint to be finished")
	}
	if len(publisher.events) != 1 || publisher.events[0].topic != "sprint.finished" {
		t.Fatalf("expected one sprint.finished event, got %#v", publisher.events)
	}

	var event struct {
		SprintID    string   `json:"sprint_id"`
		ReporterIDs []string `json:"reporter_ids"`
	}
	if err := json.Unmarshal(publisher.events[0].payload, &event); err != nil {
		t.Fatalf("decode published event: %v", err)
	}
	if event.SprintID != "sprint-published" || len(event.ReporterIDs) != 1 || event.ReporterIDs[0] != "reporter-1" {
		t.Fatalf("unexpected sprint.finished event: %#v", event)
	}
}

func TestWorkItemServicePublishesWorkItemChangeEvents(t *testing.T) {
	publisher := &eventPublisher{}
	repository := memory.NewWorkItemRepository()
	workItemService := service.NewWorkItemServiceWithPublisher(repository, publisher)

	item := validWorkItem()
	item.Status = domain.WorkItemStatusInProgress
	item.AssigneeID = "assignee-old"
	created, err := workItemService.CreateWorkItem(item)
	if err != nil {
		t.Fatalf("create work item: %v", err)
	}

	created.Status = domain.WorkItemStatusReview
	created.AssigneeID = "assignee-new"
	if _, err := workItemService.UpdateWorkItem(created); err != nil {
		t.Fatalf("update work item: %v", err)
	}

	if len(publisher.events) != 2 {
		t.Fatalf("expected status and assignment events, got %#v", publisher.events)
	}
	if publisher.events[0].topic != "workitem.status_changed" {
		t.Fatalf("expected status event, got %q", publisher.events[0].topic)
	}
	if publisher.events[1].topic != "workitem.assigned" {
		t.Fatalf("expected assignment event, got %q", publisher.events[1].topic)
	}

	var statusEvent map[string]string
	if err := json.Unmarshal(publisher.events[0].payload, &statusEvent); err != nil {
		t.Fatalf("decode status event: %v", err)
	}
	if statusEvent["old_status"] != string(domain.WorkItemStatusInProgress) ||
		statusEvent["new_status"] != string(domain.WorkItemStatusReview) {
		t.Fatalf("unexpected status event: %#v", statusEvent)
	}

	var assignmentEvent map[string]string
	if err := json.Unmarshal(publisher.events[1].payload, &assignmentEvent); err != nil {
		t.Fatalf("decode assignment event: %v", err)
	}
	if assignmentEvent["previous_assignee_id"] != "assignee-old" ||
		assignmentEvent["assignee_id"] != "assignee-new" {
		t.Fatalf("unexpected assignment event: %#v", assignmentEvent)
	}
}

func TestWorkItemServiceChecksSprintBeforeAndAfterAllItemsAreDone(t *testing.T) {
	notifier := &sprintNotifier{}
	workItemService := service.NewWorkItemServiceWithNotifier(
		memory.NewWorkItemRepository(),
		notifier,
	)

	doneItem := validWorkItem()
	doneItem.SprintID = "sprint-test"
	doneItem.Status = domain.WorkItemStatusDone
	doneItem.ReporterID = "reporter-1"
	if _, err := workItemService.CreateWorkItem(doneItem); err != nil {
		t.Fatalf("create done work item: %v", err)
	}

	unfinishedItem := validWorkItem()
	unfinishedItem.Title = "Unfinished sprint work"
	unfinishedItem.SprintID = "sprint-test"
	unfinishedItem.ReporterID = "reporter-2"
	createdUnfinishedItem, err := workItemService.CreateWorkItem(unfinishedItem)
	if err != nil {
		t.Fatalf("create unfinished work item: %v", err)
	}

	finished, err := workItemService.CheckSprintFinished("sprint-test")
	if err != nil {
		t.Fatalf("check unfinished sprint: %v", err)
	}
	if finished {
		t.Fatal("expected sprint with unfinished work to be unfinished")
	}
	if len(notifier.calls) != 0 {
		t.Fatalf("expected no notification for unfinished sprint, got %#v", notifier.calls)
	}

	createdUnfinishedItem.Status = domain.WorkItemStatusDone
	if _, err := workItemService.UpdateWorkItem(createdUnfinishedItem); err != nil {
		t.Fatalf("mark unfinished work item done: %v", err)
	}

	finished, err = workItemService.CheckSprintFinished("sprint-test")
	if err != nil {
		t.Fatalf("check finished sprint: %v", err)
	}
	if !finished {
		t.Fatal("expected sprint with all work done to be finished")
	}
	if len(notifier.calls) != 1 {
		t.Fatalf("expected one notification for finished sprint, got %#v", notifier.calls)
	}
	if len(notifier.calls[0].reporterIDs) != 2 {
		t.Fatalf("expected both reporters in notification, got %#v", notifier.calls[0].reporterIDs)
	}
}

func TestWorkItemServiceDoesNotCheckSprintWhenWorkItemBecomesDone(t *testing.T) {
	notifier := &sprintNotifier{}
	workItemService := service.NewWorkItemServiceWithNotifier(
		memory.NewWorkItemRepository(),
		notifier,
	)
	first := validWorkItem()
	first.SprintID = "sprint-1"
	first.ReporterID = "reporter-1"
	firstCreated, err := workItemService.CreateWorkItem(first)
	if err != nil {
		t.Fatalf("create first work item: %v", err)
	}
	second := validWorkItem()
	second.SprintID = "sprint-1"
	second.ReporterID = "reporter-2"
	secondCreated, err := workItemService.CreateWorkItem(second)
	if err != nil {
		t.Fatalf("create second work item: %v", err)
	}

	firstCreated.Status = domain.WorkItemStatusDone
	if _, err := workItemService.UpdateWorkItem(firstCreated); err != nil {
		t.Fatalf("update first work item: %v", err)
	}
	if len(notifier.calls) != 0 {
		t.Fatalf("expected no notification before sprint completion, got %#v", notifier.calls)
	}

	secondCreated.Status = domain.WorkItemStatusDone
	if _, err := workItemService.UpdateWorkItem(secondCreated); err != nil {
		t.Fatalf("update second work item: %v", err)
	}
	if len(notifier.calls) != 0 {
		t.Fatalf("expected no automatic notification, got %#v", notifier.calls)
	}
}

func TestWorkItemServiceDoesNotNotifyUnfinishedSprint(t *testing.T) {
	notifier := &sprintNotifier{}
	workItemService := service.NewWorkItemServiceWithNotifier(
		memory.NewWorkItemRepository(),
		notifier,
	)
	item := validWorkItem()
	item.SprintID = "sprint-1"
	item.ReporterID = "reporter-1"
	if _, err := workItemService.CreateWorkItem(item); err != nil {
		t.Fatalf("create work item: %v", err)
	}

	finished, err := workItemService.CheckSprintFinished("sprint-1")
	if err != nil {
		t.Fatalf("check sprint: %v", err)
	}
	if finished {
		t.Fatal("expected sprint to be unfinished")
	}
	if len(notifier.calls) != 0 {
		t.Fatalf("expected no notifications, got %#v", notifier.calls)
	}
}

func TestWorkItemServiceTreatsEmptySprintAsFinished(t *testing.T) {
	notifier := &sprintNotifier{}
	workItemService := service.NewWorkItemServiceWithNotifier(
		memory.NewWorkItemRepository(),
		notifier,
	)

	finished, err := workItemService.CheckSprintFinished("sprint-1")
	if err != nil {
		t.Fatalf("check sprint: %v", err)
	}
	if !finished {
		t.Fatal("expected empty sprint to be finished")
	}
	if len(notifier.calls) != 0 {
		t.Fatalf("expected no notifications, got %#v", notifier.calls)
	}
}

type sprintNotificationCall struct {
	sprintID    string
	reporterIDs []string
}

type sprintNotifier struct {
	calls []sprintNotificationCall
}

type publishedEvent struct {
	topic   string
	payload []byte
}

type eventPublisher struct {
	events []publishedEvent
}

func (publisher *eventPublisher) Publish(topic string, payload []byte) error {
	publisher.events = append(publisher.events, publishedEvent{
		topic:   topic,
		payload: append([]byte(nil), payload...),
	})
	return nil
}

func (notifier *sprintNotifier) SendSprintFinishedNotifications(
	sprintID string,
	reporterIDs []string,
) error {
	notifier.calls = append(notifier.calls, sprintNotificationCall{
		sprintID:    sprintID,
		reporterIDs: append([]string(nil), reporterIDs...),
	})
	return nil
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

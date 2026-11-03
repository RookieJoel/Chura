package grpc

import (
	"reflect"
	"testing"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// fullWorkItem sets every domain.WorkItem field, so a mapper that drops one
// fails the round-trip comparison instead of losing data silently.
func fullWorkItem() domain.WorkItem {
	createdAt := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	return domain.WorkItem{
		ID:          "item-1",
		ProjectID:   "project-1",
		SprintID:    "sprint-1",
		Title:       "Write login page",
		Description: "Form with email and password",
		Type:        domain.WorkItemTypeUserStory,
		Status:      domain.WorkItemStatusInProgress,
		Priority:    domain.WorkItemPriorityHigh,
		AssigneeID:  "user-2",
		ReporterID:  "user-1",
		StoryPoints: 5,
		Features:    map[string]any{"component": "auth"},
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt.Add(time.Hour),
	}
}

func TestWorkItemOutputRoundTripKeepsEveryField(t *testing.T) {
	want := fullWorkItem()

	message, err := workItemToProto(want)
	if err != nil {
		t.Fatalf("workItemToProto: %v", err)
	}
	got := workItemFromProto(message)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip changed work item\n got: %+v\nwant: %+v", got, want)
	}
}

func TestWorkItemInputRoundTripKeepsEveryInputField(t *testing.T) {
	item := fullWorkItem()
	// ID and timestamps are not part of WorkItemInput: the server sets them.
	want := item
	want.ID, want.CreatedAt, want.UpdatedAt = "", time.Time{}, time.Time{}

	got := workItemFromInput(workItemInputFromDomain(item))

	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip changed work item input\n got: %+v\nwant: %+v", got, want)
	}
}

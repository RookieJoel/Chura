package http

import (
	"errors"
	"testing"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// fakeWorkItemGateway returns err from every call, or canned items.
type fakeWorkItemGateway struct{ err error }

func (g fakeWorkItemGateway) CreateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	return item, g.err
}

func (g fakeWorkItemGateway) GetWorkItem(id string) (domain.WorkItem, error) {
	return domain.WorkItem{ID: id}, g.err
}

func (g fakeWorkItemGateway) ListWorkItems(string) ([]domain.WorkItem, error) {
	return []domain.WorkItem{{ID: "1"}}, g.err
}

func (g fakeWorkItemGateway) UpdateWorkItem(item domain.WorkItem) (domain.WorkItem, error) {
	return item, g.err
}

func (g fakeWorkItemGateway) DeleteWorkItem(string) error { return g.err }

func TestHandleWorkItemMessage_RoutesOperations(t *testing.T) {
	cases := []struct{ op, want string }{
		{"create", "created"}, {" GET ", "retrieved"}, {"list", "listed"},
		{"update", "updated"}, {"delete", "deleted"},
	}
	for _, tc := range cases {
		got := handleWorkItemMessage(fakeWorkItemGateway{}, workItemMessage{Operation: tc.op, ID: "1"})
		if got.Operation != tc.want || got.Error != "" {
			t.Fatalf("%s: got %+v", tc.op, got)
		}
	}
}

func TestHandleWorkItemMessage_ReportsErrors(t *testing.T) {
	gw := fakeWorkItemGateway{err: errors.New("boom")}
	for _, op := range []string{"create", "get", "list", "update", "delete"} {
		if got := handleWorkItemMessage(gw, workItemMessage{Operation: op}); got.Error != "boom" {
			t.Fatalf("%s: got %+v", op, got)
		}
	}
	if got := handleWorkItemMessage(gw, workItemMessage{Operation: "explode"}); got.Error != "unsupported operation" {
		t.Fatalf("unsupported: got %+v", got)
	}
}

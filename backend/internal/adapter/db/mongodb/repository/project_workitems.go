package repository

import (
	"context"
	"fmt"

	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ out.ProjectWorkItemStore = (*ProjectWorkItemStore)(nil)

// ProjectWorkItemStore gives the Project service project-wide access to Work
// Items: counting them (template-change eligibility) and removing them in bulk
// (permanent Project deletion).
type ProjectWorkItemStore struct {
	items    *mongo.Collection
	comments *mongo.Collection
}

func NewProjectWorkItemStore(database *mongo.Database) *ProjectWorkItemStore {
	return &ProjectWorkItemStore{
		items:    database.Collection("work_items"),
		comments: database.Collection("work_item_comments"),
	}
}

func (s *ProjectWorkItemStore) CountByProject(ctx context.Context, projectID string) (int, error) {
	n, err := s.items.CountDocuments(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return 0, fmt.Errorf("count work items of project %s: %w", projectID, err)
	}
	return int(n), nil
}

// DeleteByProject removes every Work Item of the Project and their comments.
// Comments go first so a failure part-way never leaves orphaned comments
// behind: whatever remains is still reachable from a Work Item and a retry
// finishes the job.
func (s *ProjectWorkItemStore) DeleteByProject(ctx context.Context, projectID string) error {
	filter := bson.M{"project_id": projectID}
	ids, err := s.items.Distinct(ctx, "_id", filter)
	if err != nil {
		return fmt.Errorf("list work item ids of project %s: %w", projectID, err)
	}
	if len(ids) > 0 {
		if _, err := s.comments.DeleteMany(ctx, bson.M{"work_item_id": bson.M{"$in": ids}}); err != nil {
			return fmt.Errorf("delete work item comments of project %s: %w", projectID, err)
		}
	}
	if _, err := s.items.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("delete work items of project %s: %w", projectID, err)
	}
	return nil
}

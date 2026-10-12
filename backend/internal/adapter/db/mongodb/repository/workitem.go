package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	_ out.WorkItemRepository   = (*WorkItemRepository)(nil)
	_ out.SprintWorkItemReader = (*WorkItemRepository)(nil)
)

type WorkItemRepository struct {
	collection *mongo.Collection
}

func NewWorkItemRepository(database *mongo.Database) *WorkItemRepository {
	return &WorkItemRepository{collection: database.Collection("work_items")}
}

func (repository *WorkItemRepository) EnsureIndexes() error {
	_, err := repository.collection.Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{Keys: bson.D{{Key: "project_id", Value: 1}}},
			{Keys: bson.D{{Key: "project_id", Value: 1}, {Key: "sprint_id", Value: 1}}},
		},
	)
	return err
}

type workItemDocument struct {
	ID          string                  `bson:"_id"`
	ProjectID   string                  `bson:"project_id"`
	SprintID    string                  `bson:"sprint_id,omitempty"`
	Title       string                  `bson:"title"`
	Description string                  `bson:"description"`
	Type        domain.WorkItemType     `bson:"type"`
	Status      domain.WorkItemStatus   `bson:"status"`
	Priority    domain.WorkItemPriority `bson:"priority"`
	AssigneeID  string                  `bson:"assignee_id,omitempty"`
	ReporterID  string                  `bson:"reporter_id,omitempty"`
	StoryPoints int32                   `bson:"story_points"`
	Features    map[string]any          `bson:"features,omitempty"`
	CreatedAt   time.Time               `bson:"created_at"`
	UpdatedAt   time.Time               `bson:"updated_at"`
}

func (repository *WorkItemRepository) Create(item domain.WorkItem) (domain.WorkItem, error) {
	_, err := repository.collection.InsertOne(context.Background(), documentFromDomain(item))
	if err != nil {
		return domain.WorkItem{}, err
	}
	return item, nil
}

func (repository *WorkItemRepository) Get(id string) (domain.WorkItem, error) {
	var document workItemDocument
	if err := repository.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.WorkItem{}, domain.ErrNotFound
		}
		return domain.WorkItem{}, err
	}
	return domainFromDocument(document), nil
}

func (repository *WorkItemRepository) List(filter out.WorkItemFilter) ([]domain.WorkItem, error) {
	query := bson.M{}
	if filter.ProjectID != "" {
		query["project_id"] = filter.ProjectID
	}
	if filter.SprintID != "" {
		query["sprint_id"] = filter.SprintID
	}
	return repository.find(context.Background(), query)
}

// ListBySprint returns the work items of projectID committed to sprintID.
func (repository *WorkItemRepository) ListBySprint(ctx context.Context, projectID, sprintID string) ([]domain.WorkItem, error) {
	items, err := repository.find(ctx, bson.M{"project_id": projectID, "sprint_id": sprintID})
	if err != nil {
		return nil, fmt.Errorf("list work items of sprint %s: %w", sprintID, err)
	}
	return items, nil
}

func (repository *WorkItemRepository) find(ctx context.Context, filter bson.M) ([]domain.WorkItem, error) {
	cursor, err := repository.collection.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var documents []workItemDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	items := make([]domain.WorkItem, 0, len(documents))
	for _, document := range documents {
		items = append(items, domainFromDocument(document))
	}
	return items, nil
}

func (repository *WorkItemRepository) Update(item domain.WorkItem) (domain.WorkItem, error) {
	result, err := repository.collection.ReplaceOne(
		context.Background(),
		bson.M{"_id": item.ID},
		documentFromDomain(item),
	)
	if err != nil {
		return domain.WorkItem{}, err
	}
	if result.MatchedCount == 0 {
		return domain.WorkItem{}, domain.ErrNotFound
	}
	return item, nil
}

func (repository *WorkItemRepository) Delete(id string) error {
	result, err := repository.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func documentFromDomain(item domain.WorkItem) workItemDocument {
	return workItemDocument{
		ID:          item.ID,
		ProjectID:   item.ProjectID,
		SprintID:    item.SprintID,
		Title:       item.Title,
		Description: item.Description,
		Type:        item.Type,
		Status:      item.Status,
		Priority:    item.Priority,
		AssigneeID:  item.AssigneeID,
		ReporterID:  item.ReporterID,
		StoryPoints: item.StoryPoints,
		Features:    item.Features,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func domainFromDocument(document workItemDocument) domain.WorkItem {
	return domain.WorkItem{
		ID:          document.ID,
		ProjectID:   document.ProjectID,
		SprintID:    document.SprintID,
		Title:       document.Title,
		Description: document.Description,
		Type:        document.Type,
		Status:      document.Status,
		Priority:    document.Priority,
		AssigneeID:  document.AssigneeID,
		ReporterID:  document.ReporterID,
		StoryPoints: document.StoryPoints,
		Features:    document.Features,
		CreatedAt:   document.CreatedAt,
		UpdatedAt:   document.UpdatedAt,
	}
}

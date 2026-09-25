package repository

import (
	"context"
	"errors"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ out.WorkItemRepository = (*WorkItemRepository)(nil)

type WorkItemRepository struct {
	collection *mongo.Collection
}

func NewWorkItemRepository(database *mongo.Database) *WorkItemRepository {
	return &WorkItemRepository{collection: database.Collection("work_items")}
}

func (repository *WorkItemRepository) EnsureIndexes() error {
	_, err := repository.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{Keys: bson.D{{Key: "project_id", Value: 1}}},
	)
	return err
}

type workItemDocument struct {
	ID          string                  `bson:"_id"`
	ProjectID   string                  `bson:"project_id"`
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

func (repository *WorkItemRepository) List(projectID string) ([]domain.WorkItem, error) {
	cursor, err := repository.collection.Find(
		context.Background(),
		bson.M{"project_id": projectID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var documents []workItemDocument
	if err := cursor.All(context.Background(), &documents); err != nil {
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

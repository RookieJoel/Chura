package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ out.WorkItemCommentRepository = (*WorkItemCommentRepository)(nil)

// WorkItemCommentRepository stores append-only comments beside Work Items.
type WorkItemCommentRepository struct {
	collection *mongo.Collection
}

func NewWorkItemCommentRepository(database *mongo.Database) *WorkItemCommentRepository {
	return &WorkItemCommentRepository{collection: database.Collection("work_item_comments")}
}

func (repository *WorkItemCommentRepository) EnsureIndexes() error {
	_, err := repository.collection.Indexes().CreateMany(
		context.Background(),
		[]mongo.IndexModel{
			{Keys: bson.D{{Key: "work_item_id", Value: 1}, {Key: "created_at", Value: -1}}},
		},
	)
	return err
}

type workItemCommentDocument struct {
	ID           string    `bson:"_id"`
	WorkItemID   string    `bson:"work_item_id"`
	AuthorUserID string    `bson:"author_user_id"`
	AuthorName   string    `bson:"author_name"`
	AuthorEmail  string    `bson:"author_email"`
	Body         string    `bson:"body"`
	CreatedAt    time.Time `bson:"created_at"`
}

func (repository *WorkItemCommentRepository) Create(ctx context.Context, comment *domain.WorkItemComment) error {
	_, err := repository.collection.InsertOne(ctx, workItemCommentDocument{
		ID:           comment.ID,
		WorkItemID:   comment.WorkItemID,
		AuthorUserID: comment.AuthorUserID,
		AuthorName:   comment.AuthorName,
		AuthorEmail:  comment.AuthorEmail,
		Body:         comment.Body,
		CreatedAt:    comment.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("create work item comment: %w", err)
	}
	return nil
}

func (repository *WorkItemCommentRepository) ListByWorkItemID(ctx context.Context, workItemID string) ([]domain.WorkItemComment, error) {
	cursor, err := repository.collection.Find(
		ctx,
		bson.M{"work_item_id": workItemID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("list work item comments: %w", err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	var documents []workItemCommentDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("list work item comments: %w", err)
	}
	items := make([]domain.WorkItemComment, 0, len(documents))
	for _, document := range documents {
		items = append(items, domain.WorkItemComment{
			ID:           document.ID,
			WorkItemID:   document.WorkItemID,
			AuthorUserID: document.AuthorUserID,
			AuthorName:   document.AuthorName,
			AuthorEmail:  document.AuthorEmail,
			Body:         document.Body,
			CreatedAt:    document.CreatedAt,
		})
	}
	return items, nil
}

func (repository *WorkItemCommentRepository) DeleteByWorkItemID(ctx context.Context, workItemID string) error {
	_, err := repository.collection.DeleteMany(ctx, bson.M{"work_item_id": workItemID})
	if err != nil {
		return fmt.Errorf("delete work item comments: %w", err)
	}
	return nil
}

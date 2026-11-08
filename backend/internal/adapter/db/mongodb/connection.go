package mongodb

import (
	"context"
	"fmt"

	repository "github.com/RookieJoel/Chura/backend/internal/adapter/db/mongodb/repository"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	Client              *mongo.Client
	Database            *mongo.Database
	WorkItemsrepository out.WorkItemRepository
	WorkItemComments    out.WorkItemCommentRepository
	SprintWorkItems     out.SprintWorkItemReader
}

func Connect(uri, databaseName string) (*Connection, error) {
	ctx := context.Background()
	if uri == "" {
		return nil, fmt.Errorf("MONGODB_URI is required")
	}
	if databaseName == "" {
		databaseName = "chura"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	database := client.Database(databaseName)
	workItemRepository := repository.NewWorkItemRepository(database)
	if err := workItemRepository.EnsureIndexes(); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	commentRepository := repository.NewWorkItemCommentRepository(database)
	if err := commentRepository.EnsureIndexes(); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return &Connection{
		Client:              client,
		Database:            database,
		WorkItemsrepository: workItemRepository,
		WorkItemComments:    commentRepository,
		SprintWorkItems:     workItemRepository,
	}, nil
}

func (connection *Connection) Close() error {
	return connection.Client.Disconnect(context.Background())
}

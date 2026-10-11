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
	workItemRepository := repository.NewWorkItemRepository(client.Database(databaseName))
	if err := workItemRepository.EnsureIndexes(); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return &Connection{
		Client:              client,
		Database:            client.Database(databaseName),
		WorkItemsrepository: workItemRepository,
	}, nil
}

func (connection *Connection) Close() error {
	return connection.Client.Disconnect(context.Background())
}

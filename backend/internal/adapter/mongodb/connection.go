package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Connect(ctx context.Context, uri string, databaseName string) (*Connection, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return &Connection{
		Client:   client,
		Database: client.Database(databaseName),
	}, nil
}

func (connection *Connection) Close(ctx context.Context) error {
	return connection.Client.Disconnect(ctx)
}

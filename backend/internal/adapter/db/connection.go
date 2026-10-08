package db

import (
	"errors"
	"fmt"

	"github.com/RookieJoel/Chura/backend/internal/adapter/db/mongodb"
	"github.com/RookieJoel/Chura/backend/internal/adapter/db/postgres"
	"gorm.io/gorm"
)

type Connections struct {
	Postgres *gorm.DB
	Mongo    *mongodb.Connection
}

func ConnectPostgresDB(postgresURL string) (*gorm.DB, error) {
	postgresDB, err := postgres.NewGormDB(postgresURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	return postgresDB, nil
}

func ConnectMongoDB() (*mongodb.Connection, error) {
	mongoConnection, err := mongodb.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}

	return mongoConnection, nil
}

func (c *Connections) Close() error {
	var errs []error

	sqlDB, err := c.Postgres.DB()
	if err != nil {
		errs = append(errs, fmt.Errorf("access postgres handle: %w", err))
	} else if err := sqlDB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close postgres: %w", err))
	}

	if err := c.Mongo.Close(); err != nil {
		errs = append(errs, fmt.Errorf("close mongodb: %w", err))
	}

	return errors.Join(errs...)
}

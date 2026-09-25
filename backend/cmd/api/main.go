package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/RookieJoel/Chura/backend/internal/adapter"
	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/grpc"
	workitempb "github.com/RookieJoel/Chura/backend/internal/adapter/grpc/pb/workitem"
	workitemhttp "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	mongodb "github.com/RookieJoel/Chura/backend/internal/adapter/mongodb"
	repository "github.com/RookieJoel/Chura/backend/internal/adapter/mongodb/repository"
	"github.com/RookieJoel/Chura/backend/internal/port/driven"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := adapter.LoadDotEnv(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	var workItemRepository driven.WorkItemRepository
	if databaseURL := os.Getenv("MONGODB_URI"); databaseURL != "" {
		databaseName := os.Getenv("MONGODB_DATABASE")
		if databaseName == "" {
			databaseName = "chura"
		}
		connection, err := mongodb.Connect(context.Background(), databaseURL, databaseName)
		if err != nil {
			log.Printf("MongoDB connection failed: %v", err)
		} else {
			mongoRepository := repository.NewWorkItemRepository(connection.Database)
			if err := mongoRepository.EnsureIndexes(); err != nil {
				_ = connection.Close(context.Background())
				log.Printf("MongoDB index setup failed: %v", err)
			} else {
				defer connection.Close(context.Background())
				workItemRepository = mongoRepository
			}
		}
	}

	workItemService := service.NewWorkItemService(workItemRepository)

	grpcServer := grpc.NewServer()
	workitempb.RegisterWorkItemServiceServer(grpcServer, workitemgrpc.NewServer(workItemService))
	grpcListener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal(err)
		}
	}()

	grpcConnection, err := grpc.NewClient("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	workitemhttp.RegisterWorkItemWebSocket(app, workitemgrpc.NewWorkItemGateway(grpcConnection))

	log.Fatal(app.Listen(":8080"))
}

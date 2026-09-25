package main

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/grpc"
	workitempb "github.com/RookieJoel/Chura/backend/internal/adapter/grpc/pb/workitem"
	workitemhttp "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	repository "github.com/RookieJoel/Chura/backend/internal/adapter/mongodb/repository"
	"github.com/RookieJoel/Chura/backend/internal/port/driven"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := loadDotEnv(); err != nil {
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
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(databaseURL))
		if err != nil {
			log.Printf("MongoDB connection failed: %v", err)
		} else {
			if err := client.Ping(context.Background(), nil); err != nil {
				_ = client.Disconnect(context.Background())
				log.Printf("MongoDB ping failed: %v", err)
			} else {
				mongoRepository := repository.NewWorkItemRepository(client.Database(databaseName))
				if err := mongoRepository.EnsureIndexes(); err != nil {
					_ = client.Disconnect(context.Background())
					log.Printf("MongoDB index setup failed: %v", err)
				} else {
					defer client.Disconnect(context.Background())
					workItemRepository = mongoRepository
				}
			}
		}
	}
	if workItemRepository == nil {
		log.Println("MONGODB_URI is not configured or MongoDB is unavailable; using in-memory repository")
		workItemRepository = repository.NewMemoryWorkItemRepository()
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

func loadDotEnv() error {
	candidates := []string{".env", filepath.Join(".", ".env")}
	for _, candidate := range candidates {
		content, err := os.ReadFile(candidate)
		if err != nil {
			continue
		}
		for _, rawLine := range strings.Split(string(content), "\n") {
			line := strings.TrimSpace(rawLine)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
		return nil
	}
	return os.ErrNotExist
}

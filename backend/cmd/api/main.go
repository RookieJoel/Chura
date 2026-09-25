package main

import (
	"log"
	"net"
	"os"

	"github.com/RookieJoel/Chura/backend/internal/adapter/db"
	memory "github.com/RookieJoel/Chura/backend/internal/adapter/db/postgres/repository"
	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/grpc"
	workitempb "github.com/RookieJoel/Chura/backend/internal/adapter/grpc/pb/workitem"
	"github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	workitemhttp "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := LoadDotEnv(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	frontendURL := os.Getenv("FRONTEND_URL")

	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	postgresDB, err := db.ConnectPostgresDB(databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	mongoConnection, err := db.ConnectMongoDB()
	if err != nil {
		log.Fatal(err)
	}

	connections := &db.Connections{Postgres: postgresDB, Mongo: mongoConnection}
	defer connections.Close()

	sprintRepository := memory.NewSprintRepository(connections.Postgres)

	sprintService := service.NewSprintService(
		sprintRepository,
	)

	sprintHandler := http.NewSprintHandler(
		sprintService,
	)

	app := http.NewRouter(
		sprintHandler,
		frontendURL,
	)

	var workItemRepository out.WorkItemRepository = connections.Mongo.WorkItemsrepository

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

	log.Printf("server running on :%s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

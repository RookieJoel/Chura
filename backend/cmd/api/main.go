package main

import (
	"context"
	"log"
	"net"
	"os"

	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/grpc"
	workitempb "github.com/RookieJoel/Chura/backend/internal/adapter/grpc/pb/workitem"
	"github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	workitemhttp "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/postgres"
	memory "github.com/RookieJoel/Chura/backend/internal/adapter/postgres/repository"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}
	ctx := context.Background()

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

	db, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sprintRepository := memory.NewSprintRepository(db)

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

	log.Printf("server running on :%s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}

	workItemRepository := memory.NewWorkItemRepository()
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

}

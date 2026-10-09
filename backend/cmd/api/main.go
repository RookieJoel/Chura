package main

import (
	"log"
	"net"
	"os"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/RookieJoel/Chura/backend/internal/adapter/db"
	memory "github.com/RookieJoel/Chura/backend/internal/adapter/db/postgres/repository"

	emailadapter "github.com/RookieJoel/Chura/backend/internal/adapter/email"
	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/handler/grpc"
	workitempb "github.com/RookieJoel/Chura/backend/internal/adapter/handler/grpc/pb/workitem"
	"github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	workitemhttp "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/messaging/rabbitmq"
	"github.com/RookieJoel/Chura/backend/internal/adapter/middleware"
	notificationmemory "github.com/RookieJoel/Chura/backend/internal/adapter/postgres/repository"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("load .env: %v; using process environment", err)
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

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	postgresDB, err := db.ConnectPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	mongoConnection, err := db.ConnectMongoDB(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatal(err)
	}

	connections := &db.Connections{Postgres: postgresDB, Mongo: mongoConnection}
	defer connections.Close()

	sprintRepository := memory.NewSprintRepository(connections.Postgres)

	sprintService := service.NewSprintService(
		sprintRepository,
	)

	sprintHandler := workitemhttp.NewSprintHandler(
		sprintService,
	)

	authHandler := http.NewAuthHandler()

	jwks, err := keyfunc.NewDefault([]string{cfg.KeycloakJWKSURL})
	if err != nil {
		log.Fatalf("Failed to initialize JWKS: %v", err)
	}

	authMiddleware := middleware.KeycloakAuth(jwks.Keyfunc)

	app := http.NewRouter(
		sprintHandler,
		authHandler,
		cfg.FrontendURL,
		authMiddleware,
	)

	var workItemRepository out.WorkItemRepository = connections.Mongo.WorkItemsrepository

	notificationService := service.NewNotificationService(
		notificationmemory.NotificationRecipientRepository{TestEmail: os.Getenv("TEST_EMAIL_TO")},
		emailadapter.SMTPSender{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     os.Getenv("SMTP_PORT"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
	)
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		log.Fatal("RABBITMQ_URL is not set")
	}
	rabbitConnection, err := rabbitmq.Connect(rabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConnection.Close()

	rabbitPublisher, err := rabbitmq.NewPublisher(rabbitConnection)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitPublisher.Close()

	notificationQueue := os.Getenv("RABBITMQ_NOTIFICATION_QUEUE")
	if notificationQueue == "" {
		notificationQueue = "chura-notifications"
	}
	notificationConsumer, err := rabbitmq.NewNotificationConsumer(rabbitConnection, notificationQueue)
	if err != nil {
		log.Fatal(err)
	}
	defer notificationConsumer.Close()
	go func() {
		if err := notificationConsumer.Consume(notificationService); err != nil {
			log.Printf("notification consumer stopped: %v", err)
		}
	}()

	workItemService := service.NewWorkItemServiceWithPublisher(workItemRepository, rabbitPublisher)

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
	http.RegisterWorkItemWebSocket(app, workitemgrpc.NewWorkItemGateway(grpcConnection))

	workitemhttp.RegisterNotificationWebSocket(app, notificationService)

	log.Printf("server running on :%s", port)
	log.Printf("server running on :%s", cfg.Port)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

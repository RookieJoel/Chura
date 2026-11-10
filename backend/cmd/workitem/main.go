package main

import (
	"log"

	"github.com/RookieJoel/Chura/backend/internal/adapter/db"
	memory "github.com/RookieJoel/Chura/backend/internal/adapter/db/postgres/repository"
	emailadapter "github.com/RookieJoel/Chura/backend/internal/adapter/email"
	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/handler/grpc"
	"github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/identity/keycloak"
	"github.com/RookieJoel/Chura/backend/internal/adapter/messaging/rabbitmq"
	"github.com/RookieJoel/Chura/backend/internal/platform/config"
	"github.com/RookieJoel/Chura/backend/internal/platform/runtime"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func main() {
	cfg, err := config.Load(config.ServiceWorkitem)
	if err != nil {
		log.Fatal(err)
	}

	connections, err := runtime.ConnectMongo(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := connections.Close(); err != nil {
			log.Printf("close connections: %v", err)
		}
	}()

	postgresDB, err := db.ConnectPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		sqlDB, dbErr := postgresDB.DB()
		if dbErr != nil {
			log.Printf("access postgres handle: %v", dbErr)
			return
		}
		if closeErr := sqlDB.Close(); closeErr != nil {
			log.Printf("close postgres: %v", closeErr)
		}
	}()

	directory := keycloak.NewDirectory(keycloak.Config{
		BaseURL:      cfg.KeycloakBaseURL,
		Realm:        cfg.KeycloakRealm,
		ClientID:     cfg.KeycloakBackendClientID,
		ClientSecret: cfg.KeycloakBackendClientSecret,
	})

	var workItemRepository out.WorkItemRepository = connections.Mongo.WorkItemsrepository

	notificationService := service.NewNotificationService(
		keycloak.NewNotificationRecipientRepository(directory),
		emailadapter.SMTPSender{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
		},
	)

	rabbitConnection, err := rabbitmq.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConnection.Close()

	rabbitPublisher, err := rabbitmq.NewPublisher(rabbitConnection)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitPublisher.Close()

	notificationHandler := service.NewNotificationEventHandler(notificationService)
	notificationConsumer, err := rabbitmq.NewNotificationConsumer(rabbitConnection, cfg.RabbitMQNotificationQueue)
	if err != nil {
		log.Fatal(err)
	}
	defer notificationConsumer.Close()
	go func() {
		if err := notificationConsumer.Consume(notificationHandler); err != nil {
			log.Printf("notification consumer stopped: %v", err)
		}
	}()

	workItemService := service.NewWorkItemServiceWithPublisher(workItemRepository, rabbitPublisher)
	gateway := workitemgrpc.NewLocalWorkItemGateway(workItemService)

	projectService := service.NewProjectConfigurationService(
		memory.NewProjectRepository(postgresDB),
		directory,
	)
	commentService := service.NewWorkItemCommentService(
		connections.Mongo.WorkItemComments,
		workItemRepository,
		projectService,
		directory,
	)

	app := http.NewBaseApp(cfg.FrontendURL, cfg.InstanceName)
	authMiddleware, err := runtime.AuthMiddleware(cfg)
	if err != nil {
		log.Fatal(err)
	}
	http.RegisterWorkItemWebSocket(app, gateway, authMiddleware)
	api := app.Group("/api/v1", authMiddleware)
	http.RegisterWorkItemCommentRoutes(api, http.NewWorkItemCommentHandler(commentService))

	if err := runtime.ListenAndServe(cfg, app); err != nil {
		log.Fatal(err)
	}
}

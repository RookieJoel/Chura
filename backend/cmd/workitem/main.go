package main

import (
	"log"

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

	app := http.NewBaseApp(cfg.FrontendURL, cfg.InstanceName)
	http.RegisterWorkItemWebSocket(app, gateway)

	if err := runtime.ListenAndServe(cfg, app); err != nil {
		log.Fatal(err)
	}
}

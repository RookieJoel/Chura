package main

import (
	"log"
	"net"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/RookieJoel/Chura/backend/internal/adapter/db"
	memory "github.com/RookieJoel/Chura/backend/internal/adapter/db/postgres/repository"
	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/handler/grpc"
	workitempb "github.com/RookieJoel/Chura/backend/internal/adapter/handler/grpc/pb/workitem"
	"github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/identity/keycloak"
	"github.com/RookieJoel/Chura/backend/internal/adapter/middleware"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
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
	defer func() {
		if err := connections.Close(); err != nil {
			log.Printf("close connections: %v", err)
		}
	}()

	sprintRepository := memory.NewSprintRepository(connections.Postgres)

	projectService := service.NewProjectConfigurationService(
		memory.NewProjectRepository(connections.Postgres),
		keycloak.NewDirectory(keycloak.Config{
			BaseURL:      cfg.KeycloakBaseURL,
			Realm:        cfg.KeycloakRealm,
			ClientID:     cfg.KeycloakBackendClientID,
			ClientSecret: cfg.KeycloakBackendClientSecret,
		}),
	)

	sprintService := service.NewSprintService(
		sprintRepository,
		projectService,
	)

	sprintHandler := http.NewSprintHandler(
		sprintService,
	)

	authHandler := http.NewAuthHandler()
	templateHandler := http.NewTemplateHandler(projectService)
	projectHandler := http.NewProjectHandler(projectService)

	jwks, err := keyfunc.NewDefault([]string{cfg.KeycloakJWKSURL})
	if err != nil {
		log.Fatalf("Failed to initialize JWKS: %v", err)
	}

	authMiddleware := middleware.KeycloakAuth(jwks.Keyfunc)

	app := http.NewRouter(
		sprintHandler,
		authHandler,
		templateHandler,
		projectHandler,
		cfg.FrontendURL,
		authMiddleware,
	)

	workItemService := service.NewWorkItemService(connections.Mongo.WorkItemsrepository)

	grpcServer := grpc.NewServer()
	workitempb.RegisterWorkItemServiceServer(grpcServer, workitemgrpc.NewServer(workItemService))
	grpcListener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal(err)
		}
	}()

	grpcConnection, err := grpc.NewClient("localhost:"+cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	http.RegisterWorkItemWebSocket(app, workitemgrpc.NewWorkItemGateway(grpcConnection))

	log.Printf("server running on :%s", cfg.Port)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

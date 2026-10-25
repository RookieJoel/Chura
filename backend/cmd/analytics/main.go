package main

import (
	"log"

	"github.com/RookieJoel/Chura/backend/internal/adapter/db"
	memory "github.com/RookieJoel/Chura/backend/internal/adapter/db/postgres/repository"
	"github.com/RookieJoel/Chura/backend/internal/adapter/exporter"
	"github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/identity/keycloak"
	"github.com/RookieJoel/Chura/backend/internal/platform/config"
	"github.com/RookieJoel/Chura/backend/internal/platform/runtime"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func main() {
	cfg, err := config.Load(config.ServiceAnalytics)
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

	directory := keycloak.NewDirectory(keycloak.Config{
		BaseURL:      cfg.KeycloakBaseURL,
		Realm:        cfg.KeycloakRealm,
		ClientID:     cfg.KeycloakBackendClientID,
		ClientSecret: cfg.KeycloakBackendClientSecret,
	})

	projectService := service.NewProjectConfigurationService(
		memory.NewProjectRepository(connections.Postgres),
		directory,
	)

	sprintRepository := memory.NewSprintRepository(connections.Postgres)
	analyticsService := service.NewAnalyticsReportingService(
		sprintRepository,
		projectService,
		projectService,
		connections.Mongo.SprintWorkItems,
		memory.NewReflectionRepository(connections.Postgres),
		[]out.ReportExporter{exporter.NewCSVExporter(), exporter.NewMarkdownExporter()},
	)

	authMiddleware, err := runtime.AuthMiddleware(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := http.NewBaseApp(cfg.FrontendURL, cfg.InstanceName)
	api := app.Group("/api/v1", authMiddleware)
	http.RegisterAnalyticsRoutes(api, http.NewAnalyticsHandler(analyticsService))

	if err := runtime.ListenAndServe(cfg, app); err != nil {
		log.Fatal(err)
	}
}

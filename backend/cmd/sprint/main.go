package main

import (
	"log"

	memory "github.com/RookieJoel/Chura/backend/internal/adapter/db/postgres/repository"
	"github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/identity/keycloak"
	"github.com/RookieJoel/Chura/backend/internal/platform/config"
	"github.com/RookieJoel/Chura/backend/internal/platform/runtime"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func main() {
	cfg, err := config.Load(config.ServiceSprint)
	if err != nil {
		log.Fatal(err)
	}

	connections, err := runtime.ConnectPostgres(cfg)
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

	projectService := service.NewProjectConfigurationService(
		memory.NewProjectRepository(connections.Postgres),
		directory,
	)

	sprintService := service.NewSprintService(
		memory.NewSprintRepository(connections.Postgres),
		projectService,
	)

	authMiddleware, err := runtime.AuthMiddleware(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := http.NewBaseApp(cfg.FrontendURL, cfg.InstanceName)
	api := app.Group("/api/v1", authMiddleware)
	http.RegisterSprintRoutes(api, http.NewSprintHandler(sprintService))

	if err := runtime.ListenAndServe(cfg, app); err != nil {
		log.Fatal(err)
	}
}

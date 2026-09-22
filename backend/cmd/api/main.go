package main

import (
	"context"
	"log"
	"os"

	httpadapter "github.com/RookieJoel/Chura/backend/internal/adapter/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/postgres"
	"github.com/RookieJoel/Chura/backend/internal/service"
)

func main() {
	ctx := context.Background()

	databaseURL := getEnv(
		"DATABASE_URL",
		"postgres://postgres:postgres@localhost:5432/chura",
	)

	frontendURL := getEnv(
		"FRONTEND_URL",
		"http://localhost:3000",
	)

	port := getEnv(
		"PORT",
		"8080",
	)

	db, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Output adapter
	sprintRepository := postgres.NewSprintRepository(db)

	// Application service
	sprintService := service.NewSprintService(
		sprintRepository,
	)

	// HTTP adapter
	sprintHandler := httpadapter.NewSprintHandler(
		sprintService,
	)

	router := httpadapter.NewRouter(
		sprintHandler,
		frontendURL,
	)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

package runtime

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/RookieJoel/Chura/backend/internal/adapter/db"
	"github.com/RookieJoel/Chura/backend/internal/adapter/middleware"
	"github.com/RookieJoel/Chura/backend/internal/platform/config"
	"github.com/RookieJoel/Chura/backend/internal/platform/consul"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware builds Keycloak JWT middleware from config.
func AuthMiddleware(cfg config.Config) (fiber.Handler, error) {
	jwks, err := keyfunc.NewDefault([]string{cfg.KeycloakJWKSURL})
	if err != nil {
		return nil, err
	}
	return middleware.KeycloakAuth(jwks.Keyfunc), nil
}

// ConnectPostgres opens Postgres when DATABASE_URL is set.
func ConnectPostgres(cfg config.Config) (*db.Connections, error) {
	postgresDB, err := db.ConnectPostgresDB(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return &db.Connections{Postgres: postgresDB}, nil
}

// ConnectMongo opens Mongo when MONGODB_URI is set.
func ConnectMongo(cfg config.Config) (*db.Connections, error) {
	mongoConnection, err := db.ConnectMongoDB(cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		return nil, err
	}
	return &db.Connections{Mongo: mongoConnection}, nil
}

// ListenAndServe starts Fiber with optional Consul registration and graceful shutdown.
func ListenAndServe(cfg config.Config, app *fiber.App) error {
	cleanup, err := consul.Register(
		cfg.ConsulServiceName,
		cfg.ConsulServiceAddress,
		cfg.ConsulServicePort,
		"/health",
	)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		addr := ":" + cfg.Port
		log.Printf("%s service listening on %s", cfg.Service, addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("%s service shutting down", cfg.Service)

	if cleanup != nil {
		cleanup()
	}
	return app.Shutdown()
}

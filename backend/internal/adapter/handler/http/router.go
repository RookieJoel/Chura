package http

import (
	"github.com/gofiber/fiber/v2"
)

func NewRouter(
	sprintHandler *SprintHandler,
	authHandler *AuthHandler,
	templateHandler *TemplateHandler,
	projectHandler *ProjectHandler,
	analyticsHandler *AnalyticsHandler,
	frontendURL string,
	instanceName string,
	authMiddleware fiber.Handler,
) *fiber.App {
	app := NewBaseApp(frontendURL, instanceName)

	api := app.Group("/api/v1", authMiddleware)

	RegisterProjectRoutes(api, templateHandler, projectHandler, authHandler)
	RegisterSprintRoutes(api, sprintHandler)
	RegisterAnalyticsRoutes(api, analyticsHandler)

	return app
}

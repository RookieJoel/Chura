package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// maxRequestBodyBytes caps every request body; larger bodies get 413.
const maxRequestBodyBytes = 1 * 1024 * 1024

func NewRouter(
	sprintHandler *SprintHandler,
	authHandler *AuthHandler,
	templateHandler *TemplateHandler,
	projectHandler *ProjectHandler,
	frontendURL string,
	authMiddleware fiber.Handler,
) *fiber.App {

	app := fiber.New(fiber.Config{BodyLimit: maxRequestBodyBytes})

	app.Use(recover.New())

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", frontendURL)
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Chura Sprint API",
			"version": "1.0.0",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})

	api := app.Group("/api/v1", authMiddleware)

	api.Get("/templates", templateHandler.List)
	api.Get("/templates/:id", templateHandler.Get)

	api.Post("/projects", projectHandler.Create)
	api.Get("/projects/:id", projectHandler.Get)
	api.Post("/projects/:id/members", projectHandler.AddMember)
	api.Put("/projects/:id/members/:userId/role", projectHandler.AssignRole)
	api.Post("/projects/:id/work-items/validate", projectHandler.ValidateWorkItem)

	api.Post("/sprints", sprintHandler.Create)
	api.Get("/sprints", sprintHandler.List)
	api.Get("/sprints/:id", sprintHandler.Get)
	api.Put("/sprints/:id", sprintHandler.Update)
	api.Delete("/sprints/:id", sprintHandler.Delete)

	api.Get("/whoami", authHandler.WhoAmI)

	return app
}

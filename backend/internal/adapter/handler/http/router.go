package http

import "github.com/gofiber/fiber/v2"

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

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", frontendURL)
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-User-Role")

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

	api := app.Group("/api/v1")

	api.Get("/templates", requireActor, templateHandler.List)
	api.Get("/templates/:id", requireActor, templateHandler.Get)

	api.Post("/projects", requireActor, projectHandler.Create)
	api.Get("/projects/:id", requireActor, projectHandler.Get)
	api.Post("/projects/:id/members", requireActor, projectHandler.AddMember)
	api.Put("/projects/:id/members/:userId/role", requireActor, projectHandler.AssignRole)
	api.Post("/projects/:id/work-items/validate", requireActor, projectHandler.ValidateWorkItem)

	api.Post("/sprints", requireActor, sprintHandler.Create)
	api.Get("/sprints", requireActor, sprintHandler.List)
	api.Get("/sprints/:id", requireActor, sprintHandler.Get)
	api.Put("/sprints/:id", requireActor, sprintHandler.Update)
	api.Delete("/sprints/:id", requireActor, sprintHandler.Delete)

	api.Get("/whoami", authMiddleware, authHandler.WhoAmI)

	return app
}

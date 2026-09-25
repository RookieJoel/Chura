package http

import "github.com/gofiber/fiber/v2"

func NewRouter(
	sprintHandler *SprintHandler,
	frontendURL string,
) *fiber.App {

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", frontendURL)
		c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type")

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

	api.Post("/sprints", sprintHandler.Create)
	api.Get("/sprints", sprintHandler.List)
	api.Get("/sprints/:id", sprintHandler.Get)
	api.Put("/sprints/:id", sprintHandler.Update)
	api.Delete("/sprints/:id", sprintHandler.Delete)

	return app
}

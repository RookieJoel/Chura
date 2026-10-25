package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// maxRequestBodyBytes caps every request body; larger bodies get 413.
const maxRequestBodyBytes = 1 * 1024 * 1024

// NewBaseApp builds shared middleware and health endpoints for a domain service.
func NewBaseApp(frontendURL string, instanceName string) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: maxRequestBodyBytes})

	app.Use(recover.New())

	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Instance", instanceName)
		return c.Next()
	})

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
			"status":   "healthy",
			"instance": instanceName,
		})
	})

	return app
}

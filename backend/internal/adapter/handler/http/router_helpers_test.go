package http_test

import (
	"github.com/gofiber/fiber/v2"

	handler "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
)

// newTestRouter builds the real router with a pass-through Keycloak middleware,
// so tests exercise routing and the placeholder Actor headers only.
func newTestRouter(sprints *handler.SprintHandler, templates *handler.TemplateHandler, projects *handler.ProjectHandler) *fiber.App {
	passThrough := func(c *fiber.Ctx) error { return c.Next() }
	return handler.NewRouter(sprints, handler.NewAuthHandler(), templates, projects, "http://localhost:3000", passThrough)
}

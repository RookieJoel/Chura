package http_test

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	handler "github.com/RookieJoel/Chura/backend/internal/adapter/handler/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/middleware"
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

// testActorMiddleware is TEST-ONLY. It stands in for the Keycloak JWT
// middleware: it builds the Actor from X-Test-Actor-* headers and answers 401
// when no valid actor is supplied. Production never reads these headers.
func testActorMiddleware(c *fiber.Ctx) error {
	actor := domain.Actor{
		UserID: strings.TrimSpace(c.Get("X-Test-Actor-ID")),
		Email:  strings.TrimSpace(c.Get("X-Test-Actor-Email")),
		Role:   domain.SystemRole(strings.TrimSpace(c.Get("X-Test-Actor-Role"))),
	}
	if err := actor.Validate(); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}
	c.Locals(middleware.ActorLocalsKey, actor)
	return c.Next()
}

// newTestRouter builds the real router with the test-only actor middleware, so
// tests exercise routing and handlers without real JWTs.
func newTestRouter(sprints *handler.SprintHandler, templates *handler.TemplateHandler, projects *handler.ProjectHandler) *fiber.App {
	return handler.NewRouter(sprints, handler.NewAuthHandler(), templates, projects, "http://localhost:3000", testActorMiddleware)
}

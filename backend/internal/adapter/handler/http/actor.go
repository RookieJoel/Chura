package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/adapter/middleware"
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func actorFrom(c *fiber.Ctx) domain.Actor {
	actor, _ := c.Locals(middleware.ActorLocalsKey).(domain.Actor)
	return actor
}

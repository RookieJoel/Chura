package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

const actorLocalsKey = "actor"

// requireActor is a placeholder for real authentication: it reads the actor
// from the trimmed X-User-ID / X-User-Role headers and rejects invalid ones
// (including blank after trimming) with 401.
func requireActor(c *fiber.Ctx) error {
	actor := domain.Actor{
		UserID: strings.TrimSpace(c.Get("X-User-ID")),
		Role:   domain.SystemRole(strings.TrimSpace(c.Get("X-User-Role"))),
	}
	if err := actor.Validate(); err != nil {
		return writeError(c, err)
	}
	c.Locals(actorLocalsKey, actor)
	return c.Next()
}

func actorFrom(c *fiber.Ctx) domain.Actor {
	actor, _ := c.Locals(actorLocalsKey).(domain.Actor)
	return actor
}

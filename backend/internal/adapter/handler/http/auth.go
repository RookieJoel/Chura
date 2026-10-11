package http

import (
	"github.com/RookieJoel/Chura/backend/internal/adapter/middleware"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) WhoAmI(c *fiber.Ctx) error {
	// 1. Reach into the Go standard library context backpack
	ctx := c.UserContext()
	user, ok := ctx.Value(middleware.UserKey).(middleware.AuthUser)

	if !ok {
		// This safety check triggers if the middleware wasn't applied to this route
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication data not found in context",
		})
	}

	// 2. Return the data to the client
	return c.JSON(fiber.Map{
		"authenticated": true,
		"user_id":       user.ID,
		"email":         user.Email,
		"roles":         user.Roles,
	})
}

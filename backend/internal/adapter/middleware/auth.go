package middleware

import (
	"context"
	"log"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type UserContextKey string

const UserKey UserContextKey = "user_credentials"

type AuthUser struct {
	ID    string
	Email string
	Roles []string
}

func KeycloakAuth(jwksKeyFunc jwt.Keyfunc) fiber.Handler {
	return jwtware.New(jwtware.Config{
		KeyFunc: jwksKeyFunc,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("JWT validation failed for %s %s: %v", c.Method(), c.OriginalURL(), err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired JWT",
			})
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			token := c.Locals("user").(*jwt.Token)
			claims := token.Claims.(jwt.MapClaims)

			var roles []string
			if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
				if rList, ok := realmAccess["roles"].([]interface{}); ok {
					for _, r := range rList {
						roles = append(roles, r.(string))
					}
				}
			}

			authUser := AuthUser{
				ID:    claims["sub"].(string),
				Email: claims["email"].(string),
				Roles: roles,
			}

			ctx := context.WithValue(c.UserContext(), UserKey, authUser)
			c.SetUserContext(ctx)

			return c.Next()
		},
	})
}

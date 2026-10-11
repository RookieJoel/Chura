package middleware

import (
	"context"
	"log"
	"strings"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type UserContextKey string

const UserKey UserContextKey = "user_credentials"

// ActorLocalsKey is the fiber Locals key under which the middleware stores the
// authenticated domain.Actor.
const ActorLocalsKey = "actor"

const (
	roleMember  = "member"
	roleAuditor = "auditor"
)

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
			return unauthenticated(c)
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			token, ok := c.Locals("user").(*jwt.Token)
			if !ok {
				return unauthenticated(c)
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return unauthenticated(c)
			}
			sub, ok := stringClaim(claims, "sub")
			if !ok {
				return unauthenticated(c)
			}
			email, ok := stringClaim(claims, "email")
			if !ok {
				return unauthenticated(c)
			}

			roles := realmRoles(claims)
			role, ok := systemRoleFrom(roles)
			if !ok {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
			}

			c.Locals(ActorLocalsKey, domain.Actor{UserID: sub, Email: email, Role: role})
			ctx := context.WithValue(c.UserContext(), UserKey, AuthUser{ID: sub, Email: email, Roles: roles})
			c.SetUserContext(ctx)

			return c.Next()
		},
	})
}

func unauthenticated(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
}

// stringClaim returns a claim that is a non-blank string.
func stringClaim(claims jwt.MapClaims, name string) (string, bool) {
	value, ok := claims[name].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}

// realmRoles returns the string entries of realm_access.roles; anything of an
// unexpected shape is ignored.
func realmRoles(claims jwt.MapClaims) []string {
	realmAccess, ok := claims["realm_access"].(map[string]any)
	if !ok {
		return nil
	}
	entries, ok := realmAccess["roles"].([]any)
	if !ok {
		return nil
	}
	var roles []string
	for _, entry := range entries {
		if role, ok := entry.(string); ok {
			roles = append(roles, role)
		}
	}
	return roles
}

// systemRoleFrom maps Keycloak realm roles to a Chura role. It reports false
// unless exactly one of the member / auditor roles is present.
func systemRoleFrom(roles []string) (domain.SystemRole, bool) {
	var isMember, isAuditor bool
	for _, role := range roles {
		switch role {
		case roleMember:
			isMember = true
		case roleAuditor:
			isAuditor = true
		}
	}
	switch {
	case isMember && !isAuditor:
		return domain.SystemRoleTeamMember, true
	case isAuditor && !isMember:
		return domain.SystemRoleAuditor, true
	default:
		return "", false
	}
}

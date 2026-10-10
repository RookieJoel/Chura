package http

import (
	"errors"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/domain"
)

type errorResponse struct {
	Error      string             `json:"error"`
	Violations []domain.Violation `json:"violations,omitempty"`
}

// errInvalidJSONBody marks a request body that could not be decoded.
var errInvalidJSONBody = fmt.Errorf("invalid JSON body: %w", domain.ErrInvalidInput)

// publicErrors maps each client-error sentinel to its status and fixed public
// message. Order matters: the first match wins, so specific sentinels precede the ones they wrap. Internal error text (ids,
// driver messages) never reaches the response.
var publicErrors = []struct {
	target  error
	status  int
	message string
}{
	{errInvalidJSONBody, fiber.StatusBadRequest, "invalid JSON body"},
	{domain.ErrInvalidID, fiber.StatusBadRequest, "invalid id"},
	{domain.ErrInvalidInput, fiber.StatusBadRequest, "invalid input"},
	{domain.ErrUnauthenticated, fiber.StatusUnauthorized, "unauthenticated"},
	{domain.ErrNotProjectMember, fiber.StatusForbidden, "you are not a member of this project"},
	{domain.ErrForbidden, fiber.StatusForbidden, "forbidden"},
	{domain.ErrNotFound, fiber.StatusNotFound, "not found"},
	{domain.ErrConflict, fiber.StatusConflict, "conflict"},
	{domain.ErrUnavailable, fiber.StatusServiceUnavailable, "service unavailable"},
}

// writeError is the single mapping from domain errors to HTTP responses.
func writeError(c *fiber.Ctx, err error) error {
	var invalid *domain.InvalidInputError
	if errors.As(err, &invalid) {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse{Error: "invalid input", Violations: invalid.Violations})
	}
	for _, e := range publicErrors {
		if errors.Is(err, e.target) {
			return c.Status(e.status).JSON(errorResponse{Error: e.message})
		}
	}
	log.Printf("internal error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{Error: "internal server error"})
}

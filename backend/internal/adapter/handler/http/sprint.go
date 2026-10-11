package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

const sprintDateLayout = "2006-01-02"

type SprintHandler struct {
	service in.SprintService
}

func NewSprintHandler(service in.SprintService) *SprintHandler {
	return &SprintHandler{
		service: service,
	}
}

type sprintRequest struct {
	ProjectID string              `json:"project_id"`
	Name      string              `json:"name"`
	Team      string              `json:"team"`
	StartDate *string             `json:"start_date"`
	EndDate   *string             `json:"end_date"`
	Status    domain.SprintStatus `json:"status"`
}

// parseDate decodes an optional YYYY-MM-DD wire value.
func parseDate(field string, value *string) (*time.Time, *domain.Violation) {
	if value == nil || *value == "" {
		return nil, nil
	}

	t, err := time.Parse(sprintDateLayout, *value)
	if err != nil {
		return nil, &domain.Violation{Field: field, Message: field + " must use YYYY-MM-DD format"}
	}

	return &t, nil
}

// decodeSprint translates the request body into a domain Sprint.
func decodeSprint(c *fiber.Ctx) (*domain.Sprint, error) {
	var req sprintRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, errInvalidJSONBody
	}

	startDate, startViolation := parseDate("start_date", req.StartDate)
	endDate, endViolation := parseDate("end_date", req.EndDate)
	var violations []domain.Violation
	for _, v := range []*domain.Violation{startViolation, endViolation} {
		if v != nil {
			violations = append(violations, *v)
		}
	}
	if len(violations) > 0 {
		return nil, &domain.InvalidInputError{Violations: violations}
	}

	return &domain.Sprint{
		ProjectID: req.ProjectID,
		Name:      req.Name,
		Team:      req.Team,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    req.Status,
	}, nil
}

func (h *SprintHandler) Create(c *fiber.Ctx) error {
	sprint, err := decodeSprint(c)
	if err != nil {
		return writeError(c, err)
	}

	result, err := h.service.CreateSprint(c.UserContext(), actorFrom(c), sprint)
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *SprintHandler) Get(c *fiber.Ctx) error {
	result, err := h.service.GetSprint(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(result)
}

func (h *SprintHandler) List(c *fiber.Ctx) error {
	// Fiber reuses the request buffer; copy before the value leaves the handler.
	projectID := utils.CopyString(c.Query("project_id"))
	results, err := h.service.ListSprints(c.UserContext(), actorFrom(c), projectID)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(results)
}

func (h *SprintHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	sprint, err := decodeSprint(c)
	if err != nil {
		return writeError(c, err)
	}

	// project_id is immutable: never taken from an update body.
	sprint.ID, sprint.ProjectID = id, ""

	result, err := h.service.UpdateSprint(c.UserContext(), actorFrom(c), id, sprint)
	if err != nil {
		return writeError(c, err)
	}

	return c.JSON(result)
}

func (h *SprintHandler) Delete(c *fiber.Ctx) error {
	if err := h.service.DeleteSprint(c.UserContext(), actorFrom(c), c.Params("id")); err != nil {
		return writeError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

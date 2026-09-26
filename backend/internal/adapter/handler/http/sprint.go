package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

type SprintHandler struct {
	service in.SprintService
}

func NewSprintHandler(service in.SprintService) *SprintHandler {
	return &SprintHandler{
		service: service,
	}
}

type sprintRequest struct {
	Name      string              `json:"name"`
	Team      string              `json:"team"`
	StartDate *string             `json:"start_date"`
	EndDate   *string             `json:"end_date"`
	Status    domain.SprintStatus `json:"status"`
}

func parseDate(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil, errors.New("date must use YYYY-MM-DD format")
	}

	return &t, nil
}

func (h *SprintHandler) Create(c *fiber.Ctx) error {
	var req sprintRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON body",
		})
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	sprint := &domain.Sprint{
		Name:      req.Name,
		Team:      req.Team,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    req.Status,
	}

	result, err := h.service.CreateSprint(c.Context(), sprint)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *SprintHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := h.service.GetSprint(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if result == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "sprint not found",
		})
	}

	return c.JSON(result)
}

func (h *SprintHandler) List(c *fiber.Ctx) error {
	results, err := h.service.ListSprints(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(results)
}

func (h *SprintHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req sprintRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON body",
		})
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	sprint := &domain.Sprint{
		ID:        id,
		Name:      req.Name,
		Team:      req.Team,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    req.Status,
	}

	result, err := h.service.UpdateSprint(c.Context(), id, sprint)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if result == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "sprint not found",
		})
	}

	return c.JSON(result)
}

func (h *SprintHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.service.DeleteSprint(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

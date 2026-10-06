package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

type AnalyticsHandler struct {
	service in.AnalyticsService
}

func NewAnalyticsHandler(service in.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

func analyticsError(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	if errors.Is(err, domain.ErrSprintNotFound) {
		status = fiber.StatusNotFound
	}

	return c.Status(status).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func (h *AnalyticsHandler) ListReflections(c *fiber.Ctx) error {
	result, err := h.service.RetrieveSprintReflection(c.Context(), c.Params("id"))
	if err != nil {
		return analyticsError(c, err)
	}

	return c.JSON(result)
}

func (h *AnalyticsHandler) ReviewSummary(c *fiber.Ctx) error {
	result, err := h.service.GenerateSprintReviewSummary(c.Context(), c.Params("id"))
	if err != nil {
		return analyticsError(c, err)
	}

	return c.JSON(result)
}

func (h *AnalyticsHandler) ExportReport(c *fiber.Ctx) error {
	data, err := h.service.ExportSummaryReport(c.Context(), c.Params("id"))
	if err != nil {
		return analyticsError(c, err)
	}

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="sprint-`+c.Params("id")+`-review.csv"`)

	return c.Send(data)
}

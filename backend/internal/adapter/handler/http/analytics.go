package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

type AnalyticsHandler struct {
	service in.AnalyticsReportingService
}

func NewAnalyticsHandler(service in.AnalyticsReportingService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

type reflectionRequest struct {
	Answers map[string]string `json:"answers"`
}

// metricsQuery reads ?metrics=planned_items,completion_rate; empty means the
// Template defaults.
func metricsQuery(c *fiber.Ctx) []domain.MetricKey {
	// Fiber reuses the request buffer; copy before the value leaves the handler.
	raw := utils.CopyString(c.Query("metrics"))
	var keys []domain.MetricKey
	for _, part := range strings.Split(raw, ",") {
		if key := strings.TrimSpace(part); key != "" {
			keys = append(keys, domain.MetricKey(key))
		}
	}
	return keys
}

func (h *AnalyticsHandler) WorkSummary(c *fiber.Ctx) error {
	result, err := h.service.CalculateSprintWorkSummary(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func (h *AnalyticsHandler) Metrics(c *fiber.Ctx) error {
	result, err := h.service.RetrieveSprintMetrics(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

// SubmitReflection creates or replaces the caller's reflection for the Sprint.
func (h *AnalyticsHandler) SubmitReflection(c *fiber.Ctx) error {
	var req reflectionRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	result, err := h.service.CreateSprintReflection(c.UserContext(), actorFrom(c), c.Params("id"), req.Answers)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func (h *AnalyticsHandler) ListReflections(c *fiber.Ctx) error {
	result, err := h.service.RetrieveSprintReflection(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func (h *AnalyticsHandler) ReviewSummary(c *fiber.Ctx) error {
	result, err := h.service.GenerateSprintReviewSummary(c.UserContext(), actorFrom(c), c.Params("id"), metricsQuery(c))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func (h *AnalyticsHandler) ExportReviewSummary(c *fiber.Ctx) error {
	format, err := domain.ParseReportFormat(c.Query("format"))
	if err != nil {
		return writeError(c, err)
	}
	report, err := h.service.ExportSummaryReport(c.UserContext(), actorFrom(c), c.Params("id"), format, metricsQuery(c))
	if err != nil {
		return writeError(c, err)
	}
	// Attachment guesses a Content-Type from the extension; set ours after it.
	c.Attachment(report.Filename)
	c.Set(fiber.HeaderContentType, report.ContentType)
	return c.Send(report.Data)
}

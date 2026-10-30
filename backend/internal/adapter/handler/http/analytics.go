package http

import (
	"fmt"
	"slices"
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
	// The body is a generated CSV/Markdown file, never HTML: refuse anything
	// else, force download and forbid MIME sniffing so a browser can't render
	// it as a page.
	if !isDownloadContentType(report.ContentType) {
		return writeError(c, fmt.Errorf("export: unexpected content type %q", report.ContentType))
	}
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+sanitizeFilename(report.Filename)+`"`)
	c.Set(fiber.HeaderContentType, report.ContentType)
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	return c.Send(report.Data)
}

// downloadContentTypes are the only types an export may be served as.
var downloadContentTypes = []string{"text/csv", "text/markdown"}

func isDownloadContentType(contentType string) bool {
	mediaType, _, _ := strings.Cut(contentType, ";")
	return slices.Contains(downloadContentTypes, strings.ToLower(strings.TrimSpace(mediaType)))
}

// sanitizeFilename reduces name to a safe download name: no path components,
// only [A-Za-z0-9._-], no leading dot or dash. Falls back to "report".
func sanitizeFilename(name string) string {
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		name = name[i+1:]
	}
	var b strings.Builder
	pendingDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			if pendingDash && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingDash = false
			b.WriteRune(r)
		default:
			pendingDash = true
		}
	}
	out := strings.TrimLeft(b.String(), ".-")
	if out == "" {
		return "report"
	}
	return out
}

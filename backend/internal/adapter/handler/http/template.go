package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

type TemplateHandler struct {
	service in.ProjectConfigurationService
}

func NewTemplateHandler(service in.ProjectConfigurationService) *TemplateHandler {
	return &TemplateHandler{service: service}
}

func (h *TemplateHandler) List(c *fiber.Ctx) error {
	templates, err := h.service.GetAvailableTemplates(c.UserContext(), actorFrom(c))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(templates)
}

func (h *TemplateHandler) Get(c *fiber.Ctx) error {
	template, err := h.service.RetrieveTemplateDefinition(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(template)
}

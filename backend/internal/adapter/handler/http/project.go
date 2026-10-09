package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

type ProjectHandler struct {
	service in.ProjectConfigurationService
}

func NewProjectHandler(service in.ProjectConfigurationService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TemplateID  string `json:"template_id"`
}

func (h *ProjectHandler) Create(c *fiber.Ctx) error {
	var req createProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	project, err := h.service.CreateProjectBoard(c.UserContext(), actorFrom(c), in.CreateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		TemplateID:  req.TemplateID,
	})
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

func (h *ProjectHandler) Get(c *fiber.Ctx) error {
	project, err := h.service.RetrieveProjectData(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(project)
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
}

func (h *ProjectHandler) AddMember(c *fiber.Ctx) error {
	var req addMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	project, err := h.service.AddProjectMember(c.UserContext(), actorFrom(c), c.Params("id"), req.UserID)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

type assignRoleRequest struct {
	Role string `json:"role"`
}

func (h *ProjectHandler) AssignRole(c *fiber.Ctx) error {
	var req assignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	project, err := h.service.AssignProjectRole(c.UserContext(), actorFrom(c), c.Params("id"), c.Params("userId"), domain.ProjectRole(req.Role))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(project)
}

type validateWorkItemRequest struct {
	Type               string   `json:"type"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	StoryPoints        *int32   `json:"story_points"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
}

func (h *ProjectHandler) ValidateWorkItem(c *fiber.Ctx) error {
	var req validateWorkItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	result, err := h.service.ValidateWorkItemAttributes(c.UserContext(), actorFrom(c), c.Params("id"), domain.WorkItemAttributes{
		Type:               domain.WorkItemType(req.Type),
		Title:              req.Title,
		Description:        req.Description,
		StoryPoints:        req.StoryPoints,
		AcceptanceCriteria: req.AcceptanceCriteria,
	})
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

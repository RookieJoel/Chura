package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/RookieJoel/Chura/backend/internal/port/in"
)

// WorkItemCommentHandler serves REST Work Item comment endpoints.
type WorkItemCommentHandler struct {
	service in.WorkItemCommentService
}

func NewWorkItemCommentHandler(service in.WorkItemCommentService) *WorkItemCommentHandler {
	return &WorkItemCommentHandler{service: service}
}

type createWorkItemCommentRequest struct {
	Body string `json:"body"`
}

func (h *WorkItemCommentHandler) List(c *fiber.Ctx) error {
	items, err := h.service.ListWorkItemComments(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(items)
}

func (h *WorkItemCommentHandler) Create(c *fiber.Ctx) error {
	var req createWorkItemCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	comment, err := h.service.CreateWorkItemComment(c.UserContext(), actorFrom(c), c.Params("id"), req.Body)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(comment)
}

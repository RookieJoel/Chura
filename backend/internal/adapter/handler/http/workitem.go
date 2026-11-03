package http

import (
	"errors"
	"strings"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type WorkItemHandler struct {
	service in.WorkItemService
}

func NewWorkItemHandler(service in.WorkItemService) *WorkItemHandler {
	return &WorkItemHandler{service: service}
}

func (h *WorkItemHandler) Create(c *fiber.Ctx) error {
	var item domain.WorkItem
	if err := c.BodyParser(&item); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	result, err := h.service.CreateWorkItem(item)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *WorkItemHandler) Get(c *fiber.Ctx) error {
	result, err := h.service.GetWorkItem(c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func (h *WorkItemHandler) List(c *fiber.Ctx) error {
	if sprintID := strings.TrimSpace(c.Query("sprint_id")); sprintID != "" {
		result, err := h.service.ListWorkItemsBySprint(sprintID)
		if err != nil {
			return writeError(c, err)
		}
		return c.JSON(result)
	}
	result, err := h.service.ListWorkItems(strings.TrimSpace(c.Query("project_id")))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func (h *WorkItemHandler) Update(c *fiber.Ctx) error {
	var item domain.WorkItem
	if err := c.BodyParser(&item); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	item.ID = c.Params("id")
	result, err := h.service.UpdateWorkItem(item)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func (h *WorkItemHandler) Delete(c *fiber.Ctx) error {
	if err := h.service.DeleteWorkItem(c.Params("id")); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

type workItemMessage struct {
	Operation string          `json:"operation"`
	ID        string          `json:"id,omitempty"`
	ProjectID string          `json:"project_id,omitempty"`
	WorkItem  domain.WorkItem `json:"work_item,omitempty"`
}

func RegisterWorkItemWebSocket(app *fiber.App, gateway out.WorkItemGateway) {
	app.Use("/ws/work-items", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/work-items", websocket.New(func(connection *websocket.Conn) {
		for {
			var message workItemMessage
			if err := connection.ReadJSON(&message); err != nil {
				return
			}
			response := handleWorkItemMessage(gateway, message)
			if err := connection.WriteJSON(response); err != nil {
				return
			}
		}
	}))
}

func handleWorkItemMessage(gateway out.WorkItemGateway, message workItemMessage) workItemResponse {
	switch strings.ToLower(strings.TrimSpace(message.Operation)) {
	case "create":
		item, err := gateway.CreateWorkItem(message.WorkItem)
		return responseForItem("created", item, err)
	case "get":
		item, err := gateway.GetWorkItem(message.ID)
		return responseForItem("retrieved", item, err)
	case "list":
		items, err := gateway.ListWorkItems(message.ProjectID)
		if err != nil {
			return workItemResponse{Operation: "list", Error: err.Error()}
		}
		return workItemResponse{Operation: "listed", WorkItems: items}
	case "update":
		if strings.TrimSpace(message.WorkItem.ID) == "" {
			message.WorkItem.ID = message.ID
		}
		item, err := gateway.UpdateWorkItem(message.WorkItem)
		return responseForItem("updated", item, err)
	case "delete":
		err := gateway.DeleteWorkItem(message.ID)
		if err != nil {
			return workItemResponse{Operation: "deleted", Error: err.Error()}
		}
		return workItemResponse{Operation: "deleted"}
	default:
		return workItemResponse{Operation: message.Operation, Error: errors.New("unsupported operation").Error()}
	}
}

func responseForItem(operation string, item domain.WorkItem, err error) workItemResponse {
	if err != nil {
		return workItemResponse{Operation: operation, Error: err.Error()}
	}
	return workItemResponse{Operation: operation, WorkItem: &item}
}

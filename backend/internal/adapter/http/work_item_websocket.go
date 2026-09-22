package http

import (
	"errors"
	"strings"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type workItemMessage struct {
	Operation string          `json:"operation"`
	ID        string          `json:"id,omitempty"`
	ProjectID string          `json:"project_id,omitempty"`
	WorkItem  domain.WorkItem `json:"work_item,omitempty"`
}

type workItemResponse struct {
	Operation string          `json:"operation"`
	WorkItem  *domain.WorkItem `json:"work_item,omitempty"`
	WorkItems []domain.WorkItem `json:"work_items,omitempty"`
	Error     string          `json:"error,omitempty"`
}

func RegisterWorkItemWebSocket(app *fiber.App, service in.WorkItemService) {
	app.Use("/ws/work-items", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) { return c.Next() }
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/work-items", websocket.New(func(connection *websocket.Conn) {
		for {
			var message workItemMessage
			if err := connection.ReadJSON(&message); err != nil { return }
			response := handleWorkItemMessage(service, message)
			if err := connection.WriteJSON(response); err != nil { return }
		}
	}))
}

func handleWorkItemMessage(service in.WorkItemService, message workItemMessage) workItemResponse {
	switch strings.ToLower(strings.TrimSpace(message.Operation)) {
	case "create":
		item, err := service.CreateWorkItem(message.WorkItem)
		return responseForItem("created", item, err)
	case "get":
		item, err := service.GetWorkItem(message.ID)
		return responseForItem("retrieved", item, err)
	case "list":
		items, err := service.ListWorkItems(message.ProjectID)
		if err != nil { return workItemResponse{Operation: "list", Error: err.Error()} }
		return workItemResponse{Operation: "listed", WorkItems: items}
	case "update":
		item, err := service.UpdateWorkItem(message.WorkItem)
		return responseForItem("updated", item, err)
	case "delete":
		err := service.DeleteWorkItem(message.ID)
		if err != nil { return workItemResponse{Operation: "deleted", Error: err.Error()} }
		return workItemResponse{Operation: "deleted"}
	default:
		return workItemResponse{Operation: message.Operation, Error: errors.New("unsupported operation").Error()}
	}
}

func responseForItem(operation string, item domain.WorkItem, err error) workItemResponse {
	if err != nil { return workItemResponse{Operation: operation, Error: err.Error()} }
	return workItemResponse{Operation: operation, WorkItem: &item}
}
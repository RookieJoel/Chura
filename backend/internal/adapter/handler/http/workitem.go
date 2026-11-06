package http

import (
	"errors"
	"strings"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type workItemMessage struct {
	RequestID string          `json:"request_id"`
	Operation string          `json:"operation"`
	ID        string          `json:"id,omitempty"`
	ProjectID string          `json:"project_id,omitempty"`
	SprintID  string          `json:"sprint_id,omitempty"`
	WorkItem  domain.WorkItem `json:"work_item,omitempty"`
}

type workItemSprintLister interface {
	ListWorkItemsBySprint(sprintID string) ([]domain.WorkItem, error)
}

func RegisterWorkItemWebSocket(app *fiber.App, gateway out.WorkItemGateway, authMiddleware fiber.Handler) {
	app.Use("/ws/work-items", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			if err := setWebSocketBearer(c); err != nil {
				return err
			}
			return authMiddleware(c)
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/work-items", websocket.New(func(connection *websocket.Conn) {
		for {
			var message workItemMessage
			if err := connection.ReadJSON(&message); err != nil {
				return
			}
			response := workItemResponse{}
			if strings.TrimSpace(message.RequestID) == "" {
				response.Operation = message.Operation
				response.Error = "request_id is required"
			} else {
				response = handleWorkItemMessage(gateway, message)
			}
			response.RequestID = message.RequestID
			if err := connection.WriteJSON(response); err != nil {
				return
			}
		}
	}, websocket.Config{Subprotocols: []string{"chura-auth"}}))
}

func setWebSocketBearer(c *fiber.Ctx) error {
	for _, protocol := range strings.Split(c.Get("Sec-WebSocket-Protocol"), ",") {
		protocol = strings.TrimSpace(protocol)
		if strings.HasPrefix(protocol, "auth.") {
			token := strings.TrimPrefix(protocol, "auth.")
			if token == "" {
				return fiber.ErrUnauthorized
			}
			c.Request().Header.Set("Authorization", "Bearer "+token)
			return nil
		}
	}
	if token := strings.TrimSpace(c.Cookies("chura_ws_token")); token != "" {
		c.Request().Header.Set("Authorization", "Bearer "+token)
		return nil
	}
	if token := strings.TrimSpace(c.Query("access_token")); token != "" {
		c.Request().Header.Set("Authorization", "Bearer "+token)
		return nil
	}
	return fiber.ErrUnauthorized
}

func handleWorkItemMessage(gateway out.WorkItemGateway, message workItemMessage) workItemResponse {
	switch strings.ToLower(strings.TrimSpace(message.Operation)) {
	case "create":
		item, err := gateway.CreateWorkItem(message.WorkItem)
		return responseForItem("created", item, err)
	case "validate":
		valid := message.WorkItem.Validate() == nil
		response := workItemResponse{Operation: "validated", Valid: &valid}
		return response
	case "get":
		item, err := gateway.GetWorkItem(message.ID)
		return responseForItem("retrieved", item, err)
	case "list":
		if strings.TrimSpace(message.SprintID) != "" {
			lister, ok := gateway.(workItemSprintLister)
			if !ok {
				return workItemResponse{Operation: "list", Error: "sprint listing is not supported"}
			}
			items, err := lister.ListWorkItemsBySprint(message.SprintID)
			if err != nil {
				return workItemResponse{Operation: "list", Error: err.Error()}
			}
			return workItemResponse{Operation: "listed", WorkItems: items}
		}
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

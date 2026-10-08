package http

import (
	"errors"
	"strings"

	"github.com/RookieJoel/Chura/backend/internal/port/in"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type notificationMessage struct {
	Operation   string   `json:"operation"`
	SprintID    string   `json:"sprint_id,omitempty"`
	ReporterIDs []string `json:"reporter_ids,omitempty"`
}

type notificationResponse struct {
	Operation string `json:"operation"`
	Error     string `json:"error,omitempty"`
}

func RegisterNotificationWebSocket(app *fiber.App, service in.NotificationService) {
	app.Use("/ws/notifications", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/notifications", websocket.New(func(connection *websocket.Conn) {
		for {
			var message notificationMessage
			if err := connection.ReadJSON(&message); err != nil {
				return
			}
			response := handleNotificationMessage(service, message)
			if err := connection.WriteJSON(response); err != nil {
				return
			}
		}
	}))
}

func handleNotificationMessage(
	service in.NotificationService,
	message notificationMessage,
) notificationResponse {
	if strings.EqualFold(strings.TrimSpace(message.Operation), "sprint_finished") {
		err := service.SendSprintFinishedNotifications(message.SprintID, message.ReporterIDs)
		if err != nil {
			return notificationResponse{Operation: "sprint_finished", Error: err.Error()}
		}
		return notificationResponse{Operation: "notification_sent"}
	}
	return notificationResponse{
		Operation: message.Operation,
		Error:     errors.New("unsupported operation").Error(),
	}
}

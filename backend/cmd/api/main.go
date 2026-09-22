package main

import (
	"log"
	"net"

	workitemgrpc "github.com/RookieJoel/Chura/backend/internal/adapter/grpc/workitem"
	workitemhttp "github.com/RookieJoel/Chura/backend/internal/adapter/http"
	"github.com/RookieJoel/Chura/backend/internal/adapter/memory"
	"github.com/RookieJoel/Chura/backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

func main() {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	workItemRepository := memory.NewWorkItemRepository()
	workItemService := service.NewWorkItemService(workItemRepository)
	workitemhttp.RegisterWorkItemWebSocket(app, workItemService)

	grpcServer := grpc.NewServer()
	workitemgrpc.RegisterWorkItemServiceServer(grpcServer, workitemgrpc.NewServer(workItemService))
	grpcListener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal(err)
		}
	}()

	log.Fatal(app.Listen(":8080"))
}

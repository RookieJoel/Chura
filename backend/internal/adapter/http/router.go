package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(handler *SprintHandler, frontendURL string) *gin.Engine {
	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set(
			"Access-Control-Allow-Origin",
			frontendURL,
		)

		c.Writer.Header().Set(
			"Access-Control-Allow-Credentials",
			"true",
		)

		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization",
		)

		c.Writer.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Chura Agile Execution Service API",
			"version": "1.0.0",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	api := router.Group("/api/v1")

	sprints := api.Group("/sprints")
	{
		sprints.POST("", handler.CreateSprint)
		sprints.GET("", handler.ListSprints)
		sprints.GET("/:id", handler.GetSprint)
		sprints.PUT("/:id", handler.UpdateSprint)
		sprints.DELETE("/:id", handler.DeleteSprint)
	}

	return router
}

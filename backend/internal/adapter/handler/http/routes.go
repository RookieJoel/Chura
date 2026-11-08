package http

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterProjectRoutes mounts templates, projects, and whoami under /api/v1.
func RegisterProjectRoutes(api fiber.Router, templateHandler *TemplateHandler, projectHandler *ProjectHandler, authHandler *AuthHandler) {
	api.Get("/templates", templateHandler.List)
	api.Get("/templates/:id", templateHandler.Get)

	api.Get("/projects", projectHandler.List)
	api.Post("/projects", projectHandler.Create)
	api.Get("/projects/:id", projectHandler.Get)
	api.Patch("/projects/:id", projectHandler.Update)
	api.Delete("/projects/:id", projectHandler.Delete)
	api.Post("/projects/:id/members", projectHandler.AddMember)
	api.Put("/projects/:id/members/:userId/role", projectHandler.AssignRole)
	api.Delete("/projects/:id/members/:userId", projectHandler.RemoveMember)
	api.Post("/projects/:id/auditors/me", projectHandler.SelfAssignAuditor)
	api.Get("/projects/:id/auditors", projectHandler.ListAuditors)
	api.Post("/projects/:id/auditors", projectHandler.AssignAuditor)
	api.Get("/auditors", projectHandler.ListAuditorDirectory)
	api.Get("/projects/:id/feedback", projectHandler.ListFeedback)
	api.Post("/projects/:id/feedback", projectHandler.SubmitFeedback)
	api.Post("/projects/:id/work-items/validate", projectHandler.ValidateWorkItem)

	api.Get("/whoami", authHandler.WhoAmI)
}

// RegisterSprintRoutes mounts sprint CRUD under /api/v1.
func RegisterSprintRoutes(api fiber.Router, sprintHandler *SprintHandler) {
	api.Post("/sprints", sprintHandler.Create)
	api.Get("/sprints", sprintHandler.List)
	api.Get("/sprints/:id", sprintHandler.Get)
	api.Put("/sprints/:id", sprintHandler.Update)
	api.Delete("/sprints/:id", sprintHandler.Delete)
}

// RegisterAnalyticsRoutes mounts sprint analytics under /api/v1.
func RegisterAnalyticsRoutes(api fiber.Router, analyticsHandler *AnalyticsHandler) {
	api.Get("/sprints/:id/work-summary", analyticsHandler.WorkSummary)
	api.Get("/sprints/:id/metrics", analyticsHandler.Metrics)
	api.Get("/sprints/:id/reflections", analyticsHandler.ListReflections)
	api.Put("/sprints/:id/reflections/me", analyticsHandler.SubmitReflection)
	api.Get("/sprints/:id/review-summary", analyticsHandler.ReviewSummary)
	api.Get("/sprints/:id/review-summary/export", analyticsHandler.ExportReviewSummary)
}

// RegisterWorkItemCommentRoutes mounts Work Item comment REST under /api/v1.
func RegisterWorkItemCommentRoutes(api fiber.Router, commentHandler *WorkItemCommentHandler) {
	api.Get("/work-items/:id/comments", commentHandler.List)
	api.Post("/work-items/:id/comments", commentHandler.Create)
}

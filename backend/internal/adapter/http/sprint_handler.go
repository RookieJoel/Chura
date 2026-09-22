package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/in"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type SprintHandler struct {
	service in.SprintService
}

func NewSprintHandler(service in.SprintService) *SprintHandler {
	return &SprintHandler{
		service: service,
	}
}

type sprintRequest struct {
	Name      string              `json:"name" binding:"required,min=1,max=120"`
	Team      string              `json:"team" binding:"required,min=1,max=120"`
	StartDate *string             `json:"start_date"`
	EndDate   *string             `json:"end_date"`
	Status    domain.SprintStatus `json:"status"`
}

func parseDate(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	t, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (h *SprintHandler) CreateSprint(c *gin.Context) {
	var req sprintRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": err.Error(),
		})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": "invalid start_date, expected YYYY-MM-DD",
		})
		return
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": "invalid end_date, expected YYYY-MM-DD",
		})
		return
	}

	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": "end_date must be on or after start_date",
		})
		return
	}

	status := req.Status
	if status == "" {
		status = domain.Planned
	}

	sprint := &domain.Sprint{
		Name:      req.Name,
		Team:      req.Team,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    status,
	}

	result, err := h.service.CreateSprint(c.Request.Context(), sprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *SprintHandler) ListSprints(c *gin.Context) {
	sprints, err := h.service.ListSprints(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"detail": err.Error(),
		})
		return
	}

	if sprints == nil {
		sprints = []domain.Sprint{}
	}

	c.JSON(http.StatusOK, sprints)
}

func (h *SprintHandler) GetSprint(c *gin.Context) {
	id := c.Param("id")

	sprint, err := h.service.GetSprint(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"detail": err.Error(),
		})
		return
	}

	if sprint == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"detail": "Sprint not found",
		})
		return
	}

	c.JSON(http.StatusOK, sprint)
}

func (h *SprintHandler) UpdateSprint(c *gin.Context) {
	id := c.Param("id")

	var req sprintRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": err.Error(),
		})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": "invalid start_date, expected YYYY-MM-DD",
		})
		return
	}

	endDate, err := parseDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": "invalid end_date, expected YYYY-MM-DD",
		})
		return
	}

	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"detail": "end_date must be on or after start_date",
		})
		return
	}

	status := req.Status
	if status == "" {
		status = domain.Planned
	}

	sprint := &domain.Sprint{
		Name:      req.Name,
		Team:      req.Team,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    status,
	}

	result, err := h.service.UpdateSprint(
		c.Request.Context(),
		id,
		sprint,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"detail": err.Error(),
		})
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"detail": "Sprint not found",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *SprintHandler) DeleteSprint(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteSprint(
		c.Request.Context(),
		id,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"detail": "Sprint not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"detail": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

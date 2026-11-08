package http

import (
	"fmt"
	"strconv"

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

func (h *ProjectHandler) List(c *fiber.Ctx) error {
	query, err := projectListQuery(c)
	if err != nil {
		return writeError(c, err)
	}
	result, err := h.service.ListAccessibleProjects(c.UserContext(), actorFrom(c), query)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(result)
}

func projectListQuery(c *fiber.Ctx) (in.ProjectListQuery, error) {
	query := in.ProjectListQuery{Page: 1, PageSize: in.DefaultProjectPageSize}
	parameters := []struct {
		name   string
		target *int
	}{{"page", &query.Page}, {"page_size", &query.PageSize}}
	for _, parameter := range parameters {
		raw, exists := c.Queries()[parameter.name]
		if !exists {
			continue
		}
		value, err := strconv.Atoi(raw)
		if err != nil {
			return in.ProjectListQuery{}, fmt.Errorf("invalid %s: %w", parameter.name, domain.ErrInvalidInput)
		}
		*parameter.target = value
	}
	if query.Page < 1 || query.PageSize < 1 || query.PageSize > in.MaxProjectPageSize {
		return in.ProjectListQuery{}, fmt.Errorf("invalid project list pagination: %w", domain.ErrInvalidInput)
	}
	return query, nil
}

func (h *ProjectHandler) Get(c *fiber.Ctx) error {
	project, err := h.service.RetrieveProjectData(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(project)
}

type addMemberRequest struct {
	Email string `json:"email"`
}

func (h *ProjectHandler) AddMember(c *fiber.Ctx) error {
	var req addMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	project, err := h.service.AddProjectMember(c.UserContext(), actorFrom(c), c.Params("id"), req.Email)
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

func (h *ProjectHandler) RemoveMember(c *fiber.Ctx) error {
	project, err := h.service.RemoveProjectMember(c.UserContext(), actorFrom(c), c.Params("id"), c.Params("userId"))
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

func (h *ProjectHandler) SelfAssignAuditor(c *fiber.Ctx) error {
	assignment, err := h.service.SelfAssignAuditor(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(assignment)
}

func (h *ProjectHandler) ListAuditors(c *fiber.Ctx) error {
	assignments, err := h.service.ListProjectAuditors(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(assignments)
}

type assignAuditorRequest struct {
	AuditorUserID string `json:"auditor_user_id"`
}

// AssignAuditor assigns an Auditor to the Project on a Team Member's behalf.
func (h *ProjectHandler) AssignAuditor(c *fiber.Ctx) error {
	var req assignAuditorRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	assignment, err := h.service.AssignProjectAuditor(c.UserContext(), actorFrom(c), c.Params("id"), req.AuditorUserID)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(assignment)
}

type auditorResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ListAuditorDirectory lists every Auditor in the system.
func (h *ProjectHandler) ListAuditorDirectory(c *fiber.Ctx) error {
	auditors, err := h.service.ListAuditorDirectory(c.UserContext(), actorFrom(c))
	if err != nil {
		return writeError(c, err)
	}
	response := make([]auditorResponse, 0, len(auditors))
	for _, auditor := range auditors {
		response = append(response, auditorResponse{ID: auditor.ID, Name: auditor.Name, Email: auditor.Email})
	}
	return c.JSON(response)
}

type submitFeedbackRequest struct {
	Body string `json:"body"`
}

func (h *ProjectHandler) SubmitFeedback(c *fiber.Ctx) error {
	var req submitFeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	feedback, err := h.service.SubmitProjectFeedback(c.UserContext(), actorFrom(c), c.Params("id"), req.Body)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(feedback)
}

func (h *ProjectHandler) ListFeedback(c *fiber.Ctx) error {
	items, err := h.service.ListProjectFeedback(c.UserContext(), actorFrom(c), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(items)
}

type updateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TemplateID  *string `json:"template_id"`
}

func (h *ProjectHandler) Update(c *fiber.Ctx) error {
	var req updateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, errInvalidJSONBody)
	}
	project, err := h.service.UpdateProject(c.UserContext(), actorFrom(c), c.Params("id"), in.UpdateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		TemplateID:  req.TemplateID,
	})
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(project)
}

func (h *ProjectHandler) Delete(c *fiber.Ctx) error {
	if err := h.service.DeleteProject(c.UserContext(), actorFrom(c), c.Params("id")); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

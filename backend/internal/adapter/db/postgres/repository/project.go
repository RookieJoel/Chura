package memory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/RookieJoel/Chura/backend/internal/adapter/lockstripe"
	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"

	membershipLockPrefix = "project-membership:"
)

type projectModel struct {
	ID          string    `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	TemplateID  string    `gorm:"column:template_id"`
	Mode        string    `gorm:"column:mode"`
	CreatedBy   string    `gorm:"column:created_by"`
	GroupID     string    `gorm:"column:keycloak_group_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (projectModel) TableName() string { return "projects" }

type projectAuditorAssignmentModel struct {
	ProjectID     string    `gorm:"column:project_id;primaryKey"`
	AuditorUserID string    `gorm:"column:auditor_user_id;primaryKey"`
	AuditorName   string    `gorm:"column:auditor_name"`
	AuditorEmail  string    `gorm:"column:auditor_email"`
	AssignedAt    time.Time `gorm:"column:assigned_at"`
}

func (projectAuditorAssignmentModel) TableName() string {
	return "project_auditor_assignments"
}

type projectFeedbackModel struct {
	ID           string    `gorm:"column:id;primaryKey"`
	ProjectID    string    `gorm:"column:project_id"`
	AuthorUserID string    `gorm:"column:author_user_id"`
	AuthorName   string    `gorm:"column:author_name"`
	AuthorEmail  string    `gorm:"column:author_email"`
	Body         string    `gorm:"column:body"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (projectFeedbackModel) TableName() string { return "project_feedback" }

type ProjectRepository struct {
	db *gorm.DB
	// membershipLocks queues same-project requests in this process before they
	// take a pooled DB connection for the advisory lock.
	membershipLocks *lockstripe.Set
}

var _ out.ProjectRepository = (*ProjectRepository)(nil)

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db, membershipLocks: lockstripe.New()}
}

func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	model := projectModel{
		ID: p.ID, Name: p.Name, Description: p.Description, TemplateID: p.TemplateID,
		Mode: string(p.Mode), CreatedBy: p.CreatedBy, GroupID: p.GroupID,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("create project: %w", mapProjectError(err))
	}
	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var model projectModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project %s: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get project %s: %w", id, err)
	}
	return &domain.Project{
		ID: model.ID, Name: model.Name, Description: model.Description, TemplateID: model.TemplateID,
		Mode: domain.Mode(model.Mode), CreatedBy: model.CreatedBy, GroupID: model.GroupID,
		CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *ProjectRepository) ListSummaries(ctx context.Context, query out.ProjectListQuery) ([]domain.ProjectSummary, int, error) {
	if query.ProjectIDs != nil {
		projectIDs := make([]string, 0, len(query.ProjectIDs))
		for _, id := range query.ProjectIDs {
			if _, err := uuid.Parse(id); err == nil {
				projectIDs = append(projectIDs, id)
			}
		}
		query.ProjectIDs = projectIDs
		if len(query.ProjectIDs) == 0 {
			return []domain.ProjectSummary{}, 0, nil
		}
	}
	newQuery := func() *gorm.DB {
		projects := r.db.WithContext(ctx).Model(&projectModel{})
		if query.ProjectIDs != nil {
			projects = projects.Where("id IN ?", query.ProjectIDs)
		}
		return projects
	}
	if query.ProjectIDs != nil {
		if len(query.ProjectIDs) == 0 {
			return []domain.ProjectSummary{}, 0, nil
		}
	}

	var total int64
	if err := newQuery().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}

	var models []projectModel
	if err := newQuery().
		Select("id", "name", "description", "template_id", "mode").
		Order("name ASC").
		Order("id ASC").
		Offset(query.Offset).
		Limit(query.Limit).
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("list project summaries: %w", err)
	}

	summaries := make([]domain.ProjectSummary, 0, len(models))
	for _, model := range models {
		summaries = append(summaries, domain.ProjectSummary{
			ID: model.ID, Name: model.Name, Description: model.Description,
			TemplateID: model.TemplateID, Mode: domain.Mode(model.Mode),
		})
	}
	return summaries, int(total), nil
}

// WithMembershipLock holds a per-project lock for the duration of fn. An
// in-process stripe lock is taken first, so waiting requests hold no DB
// connection; the Postgres advisory lock then covers other instances. The
// transaction exists only to scope the advisory lock; fn does not use it.
func (r *ProjectRepository) WithMembershipLock(ctx context.Context, projectID string, fn func(ctx context.Context) error) error {
	unlock, err := r.membershipLocks.Lock(ctx, projectID)
	if err != nil {
		return fmt.Errorf("wait for project %s membership lock: %w", projectID, errors.Join(domain.ErrUnavailable, err))
	}
	defer unlock()
	var fnErr error
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", membershipLockPrefix+projectID).Error; err != nil {
			return fmt.Errorf("lock project %s membership: %w", projectID, err)
		}
		fnErr = fn(ctx)
		return fnErr
	})
	if fnErr != nil {
		return fnErr
	}
	return err
}

func (r *ProjectRepository) Update(ctx context.Context, p *domain.Project) error {
	result := r.db.WithContext(ctx).Model(&projectModel{}).
		Where("id = ?", p.ID).
		Updates(map[string]any{
			"name":        p.Name,
			"description": p.Description,
			"template_id": p.TemplateID,
			"mode":        string(p.Mode),
			"updated_at":  p.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("update project %s: %w", p.ID, mapProjectError(result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("project %s: %w", p.ID, domain.ErrNotFound)
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, projectID string) error {
	result := r.db.WithContext(ctx).Where("id = ?", projectID).Delete(&projectModel{})
	if result.Error != nil {
		return fmt.Errorf("delete project %s: %w", projectID, mapProjectError(result.Error))
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("project %s: %w", projectID, domain.ErrNotFound)
	}
	return nil
}

func (r *ProjectRepository) AssignAuditor(ctx context.Context, assignment domain.ProjectAuditorAssignment) (domain.ProjectAuditorAssignment, error) {
	model := projectAuditorAssignmentModel{
		ProjectID:     assignment.ProjectID,
		AuditorUserID: assignment.AuditorUserID,
		AuditorName:   assignment.AuditorName,
		AuditorEmail:  assignment.AuditorEmail,
		AssignedAt:    assignment.AssignedAt,
	}
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND auditor_user_id = ?", assignment.ProjectID, assignment.AuditorUserID).
		Attrs(projectAuditorAssignmentModel{
			AuditorName:  assignment.AuditorName,
			AuditorEmail: assignment.AuditorEmail,
			AssignedAt:   assignment.AssignedAt,
		}).
		FirstOrCreate(&model).Error
	if err != nil {
		return domain.ProjectAuditorAssignment{}, fmt.Errorf("assign auditor to project %s: %w", assignment.ProjectID, mapProjectError(err))
	}
	return domain.ProjectAuditorAssignment{
		ProjectID:     model.ProjectID,
		AuditorUserID: model.AuditorUserID,
		AuditorName:   model.AuditorName,
		AuditorEmail:  model.AuditorEmail,
		AssignedAt:    model.AssignedAt,
	}, nil
}

func (r *ProjectRepository) ListAuditors(ctx context.Context, projectID string) ([]domain.ProjectAuditorAssignment, error) {
	var models []projectAuditorAssignmentModel
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("assigned_at ASC").
		Order("auditor_user_id ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list auditors for project %s: %w", projectID, err)
	}
	assignments := make([]domain.ProjectAuditorAssignment, 0, len(models))
	for _, model := range models {
		assignments = append(assignments, domain.ProjectAuditorAssignment{
			ProjectID:     model.ProjectID,
			AuditorUserID: model.AuditorUserID,
			AuditorName:   model.AuditorName,
			AuditorEmail:  model.AuditorEmail,
			AssignedAt:    model.AssignedAt,
		})
	}
	return assignments, nil
}

func (r *ProjectRepository) IsAuditorAssigned(ctx context.Context, projectID, auditorUserID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&projectAuditorAssignmentModel{}).
		Where("project_id = ? AND auditor_user_id = ?", projectID, auditorUserID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check auditor assignment for project %s: %w", projectID, err)
	}
	return count > 0, nil
}

func (r *ProjectRepository) CreateFeedback(ctx context.Context, feedback *domain.ProjectFeedback) error {
	model := projectFeedbackModel{
		ID:           feedback.ID,
		ProjectID:    feedback.ProjectID,
		AuthorUserID: feedback.AuthorUserID,
		AuthorName:   feedback.AuthorName,
		AuthorEmail:  feedback.AuthorEmail,
		Body:         feedback.Body,
		CreatedAt:    feedback.CreatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("create project feedback: %w", mapProjectError(err))
	}
	return nil
}

func (r *ProjectRepository) ListFeedback(ctx context.Context, projectID string) ([]domain.ProjectFeedback, error) {
	var models []projectFeedbackModel
	if err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Order("id DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list feedback for project %s: %w", projectID, err)
	}
	items := make([]domain.ProjectFeedback, 0, len(models))
	for _, model := range models {
		items = append(items, domain.ProjectFeedback{
			ID:           model.ID,
			ProjectID:    model.ProjectID,
			AuthorUserID: model.AuthorUserID,
			AuthorName:   model.AuthorName,
			AuthorEmail:  model.AuthorEmail,
			Body:         model.Body,
			CreatedAt:    model.CreatedAt,
		})
	}
	return items, nil
}

// mapProjectError translates Postgres constraint violations to bare domain
// sentinels; the driver detail is logged server-side only.
func mapProjectError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	var sentinel error
	switch pgErr.Code {
	case pgUniqueViolation:
		sentinel = domain.ErrConflict
	case pgForeignKeyViolation:
		sentinel = domain.ErrNotFound
	default:
		return err
	}
	slog.Warn("project repository constraint violation",
		"code", pgErr.Code, "constraint", pgErr.ConstraintName, "table", pgErr.TableName, "detail", pgErr.Detail)
	return sentinel
}

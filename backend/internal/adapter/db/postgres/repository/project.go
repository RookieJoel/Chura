package memory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
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

type ProjectRepository struct {
	db *gorm.DB
}

var _ out.ProjectRepository = (*ProjectRepository)(nil)

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
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

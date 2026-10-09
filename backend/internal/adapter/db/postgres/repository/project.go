package memory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (projectModel) TableName() string { return "projects" }

type projectMemberModel struct {
	ProjectID string    `gorm:"column:project_id;primaryKey"`
	UserID    string    `gorm:"column:user_id;primaryKey"`
	Role      string    `gorm:"column:role"`
	AddedAt   time.Time `gorm:"column:added_at"`
}

func (projectMemberModel) TableName() string { return "project_members" }

func newMemberModel(projectID string, m domain.Member) projectMemberModel {
	return projectMemberModel{ProjectID: projectID, UserID: m.UserID, Role: string(m.Role), AddedAt: m.AddedAt}
}

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
		Mode: string(p.Mode), CreatedBy: p.CreatedBy, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		for _, m := range p.Members {
			member := newMemberModel(p.ID, m)
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
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
	var memberModels []projectMemberModel
	if err := r.db.WithContext(ctx).Where("project_id = ?", id).Order("added_at ASC, user_id ASC").Find(&memberModels).Error; err != nil {
		return nil, fmt.Errorf("get members of project %s: %w", id, err)
	}
	members := toDomainMembers(memberModels)
	return &domain.Project{
		ID: model.ID, Name: model.Name, Description: model.Description, TemplateID: model.TemplateID,
		Mode: domain.Mode(model.Mode), CreatedBy: model.CreatedBy, Members: members,
		CreatedAt: model.CreatedAt, UpdatedAt: model.UpdatedAt,
	}, nil
}

func (r *ProjectRepository) AddMember(ctx context.Context, projectID string, member domain.Member) error {
	model := newMemberModel(projectID, member)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("add member %s to project %s: %w", member.UserID, projectID, mapProjectError(err))
	}
	return nil
}

func (r *ProjectRepository) UpdateMemberRole(ctx context.Context, projectID, userID string, role domain.ProjectRole, guard func([]domain.Member) error) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock every member row of the project so concurrent role changes
		// serialise and guard sees the committed state.
		var locked []projectMemberModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("project_id = ?", projectID).Order("added_at ASC, user_id ASC").
			Find(&locked).Error; err != nil {
			return err
		}
		if err := guard(toDomainMembers(locked)); err != nil {
			return err
		}
		result := tx.Model(&projectMemberModel{}).
			Where("project_id = ? AND user_id = ?", projectID, userID).
			Update("role", string(role))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update role of %s in project %s: %w", userID, projectID, err)
	}
	return nil
}

func toDomainMembers(models []projectMemberModel) []domain.Member {
	members := make([]domain.Member, 0, len(models))
	for _, m := range models {
		members = append(members, domain.Member{UserID: m.UserID, Role: domain.ProjectRole(m.Role), AddedAt: m.AddedAt})
	}
	return members
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

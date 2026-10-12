package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/RookieJoel/Chura/backend/internal/domain"
	"github.com/RookieJoel/Chura/backend/internal/port/out"
)

type reflectionModel struct {
	ID          uint64    `gorm:"column:id;primaryKey"`
	SprintID    uint64    `gorm:"column:sprint_id"`
	ProjectID   string    `gorm:"column:project_id;type:uuid"`
	AuthorID    string    `gorm:"column:author_id"`
	AuthorEmail string    `gorm:"column:author_email"`
	TemplateID  string    `gorm:"column:template_id"`
	Answers     string    `gorm:"column:answers;type:jsonb"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (reflectionModel) TableName() string { return "sprint_reflections" }

func (m *reflectionModel) toDomain() (domain.SprintReflection, error) {
	answers := map[string]string{}
	if m.Answers != "" {
		if err := json.Unmarshal([]byte(m.Answers), &answers); err != nil {
			return domain.SprintReflection{}, fmt.Errorf("decode answers of reflection %d: %w", m.ID, err)
		}
	}
	return domain.SprintReflection{
		ID:          strconv.FormatUint(m.ID, 10),
		SprintID:    strconv.FormatUint(m.SprintID, 10),
		ProjectID:   m.ProjectID,
		AuthorID:    m.AuthorID,
		AuthorEmail: m.AuthorEmail,
		TemplateID:  m.TemplateID,
		Answers:     answers,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

type ReflectionRepository struct {
	db *gorm.DB
}

var _ out.ReflectionRepository = (*ReflectionRepository)(nil)

func NewReflectionRepository(db *gorm.DB) *ReflectionRepository {
	return &ReflectionRepository{db: db}
}

// Upsert keeps the original created_at when the author answers again.
func (r *ReflectionRepository) Upsert(ctx context.Context, reflection *domain.SprintReflection) (*domain.SprintReflection, error) {
	sprintID, err := parseSprintID(reflection.SprintID)
	if err != nil {
		return nil, err
	}
	answers, err := json.Marshal(reflection.Answers)
	if err != nil {
		return nil, fmt.Errorf("encode reflection answers: %w", err)
	}
	model := reflectionModel{
		SprintID:    sprintID,
		ProjectID:   reflection.ProjectID,
		AuthorID:    reflection.AuthorID,
		AuthorEmail: reflection.AuthorEmail,
		TemplateID:  reflection.TemplateID,
		Answers:     string(answers),
		CreatedAt:   reflection.CreatedAt,
		UpdatedAt:   reflection.UpdatedAt,
	}
	err = r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sprint_id"}, {Name: "author_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"author_email", "template_id", "answers", "updated_at"}),
	}).Create(&model).Error
	if err != nil {
		return nil, fmt.Errorf("upsert sprint reflection: %w", mapProjectError(err))
	}

	var stored reflectionModel
	if err := r.db.WithContext(ctx).
		Where("sprint_id = ? AND author_id = ?", sprintID, reflection.AuthorID).
		Take(&stored).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("reload sprint reflection: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("reload sprint reflection: %w", err)
	}
	result, err := stored.toDomain()
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *ReflectionRepository) ListBySprint(ctx context.Context, sprintID string) ([]domain.SprintReflection, error) {
	id, err := parseSprintID(sprintID)
	if err != nil {
		return nil, err
	}
	var models []reflectionModel
	if err := r.db.WithContext(ctx).
		Where("sprint_id = ?", id).
		Order("created_at ASC, id ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list sprint reflections: %w", err)
	}
	reflections := make([]domain.SprintReflection, 0, len(models))
	for _, m := range models {
		reflection, err := m.toDomain()
		if err != nil {
			return nil, err
		}
		reflections = append(reflections, reflection)
	}
	return reflections, nil
}

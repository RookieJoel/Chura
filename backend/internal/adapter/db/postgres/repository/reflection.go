package memory

import (
	"context"
	"fmt"
	"strconv"

	"github.com/RookieJoel/Chura/backend/internal/domain"

	"gorm.io/gorm"
)

type reflectionModel struct {
	gorm.Model
	SprintID uint   `gorm:"column:sprint_id"`
	Author   string `gorm:"column:author"`
	Content  string `gorm:"column:content"`
}

func (reflectionModel) TableName() string {
	return "sprint_reflections"
}

func (m *reflectionModel) toDomain() domain.SprintReflection {
	return domain.SprintReflection{
		ID:        strconv.FormatUint(uint64(m.ID), 10),
		SprintID:  strconv.FormatUint(uint64(m.SprintID), 10),
		Author:    m.Author,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

type ReflectionRepository struct {
	db *gorm.DB
}

func NewReflectionRepository(db *gorm.DB) *ReflectionRepository {
	return &ReflectionRepository{
		db: db,
	}
}

func (r *ReflectionRepository) ListBySprintID(
	ctx context.Context,
	sprintID string,
) ([]domain.SprintReflection, error) {

	id, err := strconv.ParseUint(sprintID, 10, 64)
	if err != nil {
		return []domain.SprintReflection{}, nil
	}

	var models []reflectionModel

	if err := r.db.WithContext(ctx).
		Where("sprint_id = ?", id).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list sprint reflections: %w", err)
	}

	reflections := make([]domain.SprintReflection, 0, len(models))
	for _, model := range models {
		reflections = append(reflections, model.toDomain())
	}

	return reflections, nil
}

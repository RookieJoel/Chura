package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/RookieJoel/Chura/backend/internal/domain"

	"gorm.io/gorm"
)

type reflectionModel struct {
	gorm.Model
	SprintID uint   `gorm:"column:sprint_id"`
	Author   string `gorm:"column:author"`
	Answers  string `gorm:"column:answers;type:jsonb"`
}

func (reflectionModel) TableName() string {
	return "sprint_reflections"
}

func (m *reflectionModel) toDomain() (domain.SprintReflection, error) {
	answers := map[string]string{}
	if m.Answers != "" {
		if err := json.Unmarshal([]byte(m.Answers), &answers); err != nil {
			return domain.SprintReflection{}, fmt.Errorf("decode reflection answers: %w", err)
		}
	}

	return domain.SprintReflection{
		ID:        strconv.FormatUint(uint64(m.ID), 10),
		SprintID:  strconv.FormatUint(uint64(m.SprintID), 10),
		Author:    m.Author,
		Answers:   answers,
		CreatedAt: m.CreatedAt,
	}, nil
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
		reflection, err := model.toDomain()
		if err != nil {
			return nil, err
		}
		reflections = append(reflections, reflection)
	}

	return reflections, nil
}

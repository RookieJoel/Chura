package memory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/RookieJoel/Chura/backend/internal/domain"

	"gorm.io/gorm"
)

type sprintModel struct {
	gorm.Model
	Name      string     `gorm:"column:name"`
	Team      string     `gorm:"column:team"`
	StartDate *time.Time `gorm:"column:start_date"`
	EndDate   *time.Time `gorm:"column:end_date"`
	Status    string     `gorm:"column:status"`
}

func (sprintModel) TableName() string {
	return "sprints"
}

const sprintIDClause = "id = ?"

func (m *sprintModel) toDomain() domain.Sprint {
	return domain.Sprint{
		ID:        strconv.FormatUint(uint64(m.ID), 10),
		Name:      m.Name,
		Team:      m.Team,
		StartDate: m.StartDate,
		EndDate:   m.EndDate,
		Status:    domain.SprintStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type SprintRepository struct {
	db *gorm.DB
}

func NewSprintRepository(db *gorm.DB) *SprintRepository {
	return &SprintRepository{
		db: db,
	}
}

func (r *SprintRepository) Create(
	ctx context.Context,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {

	model := sprintModel{
		Name:      sprint.Name,
		Team:      sprint.Team,
		StartDate: sprint.StartDate,
		EndDate:   sprint.EndDate,
		Status:    string(sprint.Status),
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	result := model.toDomain()
	return &result, nil
}

func (r *SprintRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Sprint, error) {

	sprintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, nil
	}

	var model sprintModel

	err = r.db.WithContext(ctx).First(&model, sprintIDClause, sprintID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get sprint: %w", err)
	}

	result := model.toDomain()
	return &result, nil
}

func (r *SprintRepository) List(
	ctx context.Context,
) ([]domain.Sprint, error) {

	var models []sprintModel

	if err := r.db.WithContext(ctx).
		Order("start_date ASC NULLS LAST").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}

	sprints := make([]domain.Sprint, 0, len(models))
	for _, model := range models {
		sprints = append(sprints, model.toDomain())
	}

	return sprints, nil
}

func (r *SprintRepository) Update(
	ctx context.Context,
	id string,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {

	sprintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, nil
	}

	updates := map[string]any{
		"name":       sprint.Name,
		"team":       sprint.Team,
		"start_date": sprint.StartDate,
		"end_date":   sprint.EndDate,
		"status":     string(sprint.Status),
		"updated_at": time.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&sprintModel{}).
		Where(sprintIDClause, sprintID).
		Updates(updates)

	if result.Error != nil {
		return nil, fmt.Errorf("update sprint: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	sprint.ID = strconv.FormatUint(sprintID, 10)

	return sprint, nil
}

func (r *SprintRepository) Delete(
	ctx context.Context,
	id string,
) error {

	sprintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return gorm.ErrRecordNotFound
	}

	result := r.db.WithContext(ctx).Delete(&sprintModel{}, sprintIDClause, sprintID)
	if result.Error != nil {
		return fmt.Errorf("delete sprint: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

package memory

import (
	"context"
	"errors"
	"fmt"

	"github.com/RookieJoel/Chura/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SprintRepository struct {
	db *pgxpool.Pool
}

func NewSprintRepository(db *pgxpool.Pool) *SprintRepository {
	return &SprintRepository{
		db: db,
	}
}

func (r *SprintRepository) Create(
	ctx context.Context,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {

	id := uuid.New()

	query := `
		INSERT INTO sprints (
			id,
			name,
			team,
			start_date,
			end_date,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		sprint.Name,
		sprint.Team,
		sprint.StartDate,
		sprint.EndDate,
		sprint.Status,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	sprint.ID = id.String()

	return sprint, nil
}

func (r *SprintRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Sprint, error) {

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return nil, nil
	}

	query := `
		SELECT
			id,
			name,
			team,
			start_date,
			end_date,
			status
		FROM sprints
		WHERE id = $1
	`

	var sprint domain.Sprint
	var status string

	err = r.db.QueryRow(
		ctx,
		query,
		sprintID,
	).Scan(
		&sprintID,
		&sprint.Name,
		&sprint.Team,
		&sprint.StartDate,
		&sprint.EndDate,
		&status,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get sprint: %w", err)
	}

	sprint.ID = sprintID.String()
	sprint.Status = domain.SprintStatus(status)

	return &sprint, nil
}

func (r *SprintRepository) List(
	ctx context.Context,
) ([]domain.Sprint, error) {

	query := `
		SELECT
			id,
			name,
			team,
			start_date,
			end_date,
			status
		FROM sprints
		ORDER BY start_date ASC NULLS LAST
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}
	defer rows.Close()

	var sprints []domain.Sprint

	for rows.Next() {
		var (
			sprint domain.Sprint
			id     uuid.UUID
			status string
		)

		err := rows.Scan(
			&id,
			&sprint.Name,
			&sprint.Team,
			&sprint.StartDate,
			&sprint.EndDate,
			&status,
		)

		if err != nil {
			return nil, fmt.Errorf("scan sprint: %w", err)
		}

		sprint.ID = id.String()
		sprint.Status = domain.SprintStatus(status)

		sprints = append(sprints, sprint)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sprints: %w", err)
	}

	return sprints, nil
}

func (r *SprintRepository) Update(
	ctx context.Context,
	id string,
	sprint *domain.Sprint,
) (*domain.Sprint, error) {

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return nil, nil
	}

	query := `
		UPDATE sprints
		SET
			name = $2,
			team = $3,
			start_date = $4,
			end_date = $5,
			status = $6
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		sprintID,
		sprint.Name,
		sprint.Team,
		sprint.StartDate,
		sprint.EndDate,
		sprint.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("update sprint: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, nil
	}

	sprint.ID = sprintID.String()

	return sprint, nil
}

func (r *SprintRepository) Delete(
	ctx context.Context,
	id string,
) error {

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return pgx.ErrNoRows
	}

	query := `
		DELETE FROM sprints
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, sprintID)
	if err != nil {
		return fmt.Errorf("delete sprint: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

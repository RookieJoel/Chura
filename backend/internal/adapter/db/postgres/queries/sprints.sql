-- name: CreateSprint :one
INSERT INTO sprints (
    name,
    team,
    start_date,
    end_date,
    status
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;


-- name: GetSprint :one
SELECT *
FROM sprints
WHERE id = $1;


-- name: ListSprints :many
SELECT *
FROM sprints
ORDER BY start_date ASC NULLS LAST;


-- name: UpdateSprint :one
UPDATE sprints
SET
    name = $2,
    team = $3,
    start_date = $4,
    end_date = $5,
    status = $6
WHERE id = $1
RETURNING *;


-- name: DeleteSprint :exec
DELETE FROM sprints
WHERE id = $1;
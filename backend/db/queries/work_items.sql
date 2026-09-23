-- Work item CRUD queries.
-- sqlc generates Go methods from the annotations below.

-- name: CreateWorkItem :one
INSERT INTO work_items (
    id,
    project_id,
    title,
    description,
    type,
    status,
    priority,
    assignee_id,
    story_points,
    features
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
RETURNING *;

-- name: GetWorkItem :one
SELECT *
FROM work_items
WHERE id = $1
LIMIT 1;

-- name: ListWorkItems :many
SELECT *
FROM work_items
WHERE project_id = $1
ORDER BY created_at ASC, id ASC;

-- name: UpdateWorkItem :one
UPDATE work_items
SET
    project_id = $2,
    title = $3,
    description = $4,
    type = $5,
    status = $6,
    priority = $7,
    assignee_id = $8,
    story_points = $9,
    features = $10,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteWorkItem :exec
DELETE FROM work_items
WHERE id = $1;

-- Multiple reporter associations.

-- name: AddWorkItemReporter :exec
INSERT INTO work_item_reporters (id, work_item_id, reporter_id)
VALUES ($1, $2, $3)
ON CONFLICT (work_item_id, reporter_id) DO NOTHING;

-- name: ListWorkItemReporterIDs :many
SELECT reporter_id
FROM work_item_reporters
WHERE work_item_id = $1
  AND reporter_id IS NOT NULL
ORDER BY reporter_id ASC;

-- name: RemoveWorkItemReporter :exec
DELETE FROM work_item_reporters
WHERE work_item_id = $1
  AND reporter_id = $2;

-- name: RemoveAllWorkItemReporters :exec
DELETE FROM work_item_reporters
WHERE work_item_id = $1;

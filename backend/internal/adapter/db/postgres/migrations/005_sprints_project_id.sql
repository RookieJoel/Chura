-- !!! DESTRUCTIVE MIGRATION !!!
-- This DELETES ALL EXISTING ROWS from `sprints` (including soft-deleted ones).
-- Apply ONLY after explicit confirmation, and ONLY after 004_create_projects.sql.
-- Rollback: goose down (drops project_id; deleted rows are NOT recoverable).
-- goose runs this file in a single transaction (no explicit BEGIN/COMMIT needed).

-- +goose Up
DELETE FROM sprints;
ALTER TABLE sprints
    ADD COLUMN project_id UUID NOT NULL REFERENCES projects(id);
CREATE INDEX IF NOT EXISTS idx_sprints_project_id ON sprints (project_id);

-- +goose Down
DROP INDEX IF EXISTS idx_sprints_project_id;
ALTER TABLE sprints DROP COLUMN IF EXISTS project_id;

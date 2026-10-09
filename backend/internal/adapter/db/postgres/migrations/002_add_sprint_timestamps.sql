-- +goose Up
ALTER TABLE sprints
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE sprints
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

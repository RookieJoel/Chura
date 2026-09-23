-- +goose Up

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE work_items (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects (id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    assignee_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    story_points INTEGER NOT NULL DEFAULT 0,
    features JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT work_items_type_check
        CHECK (type IN ('task', 'user_story', 'bug')),
    CONSTRAINT work_items_status_check
        CHECK (status IN ('to_do', 'in_progress', 'review', 'done', 'blocked')),
    CONSTRAINT work_items_priority_check
        CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    CONSTRAINT work_items_story_points_check
        CHECK (story_points >= 0)
);

CREATE INDEX work_items_project_id_idx ON work_items (project_id);

CREATE TABLE work_item_reporters (
    id TEXT PRIMARY KEY,
    work_item_id TEXT NOT NULL REFERENCES work_items (id) ON DELETE CASCADE,
    reporter_id TEXT REFERENCES users (id) ON DELETE SET NULL,
    UNIQUE (work_item_id, reporter_id)
);

CREATE INDEX work_item_reporters_work_item_id_idx
    ON work_item_reporters (work_item_id);

-- +goose Down

DROP TABLE work_item_reporters;
DROP TABLE work_items;
DROP TABLE users;
DROP TABLE projects;

-- +goose Up
CREATE TABLE IF NOT EXISTS project_auditor_assignments (
    project_id       UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    auditor_user_id  VARCHAR(100) NOT NULL,
    auditor_name     VARCHAR(200) NOT NULL DEFAULT '',
    auditor_email    VARCHAR(254) NOT NULL DEFAULT '',
    assigned_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, auditor_user_id)
);

CREATE TABLE IF NOT EXISTS project_feedback (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    author_user_id  VARCHAR(100) NOT NULL,
    author_name     VARCHAR(200) NOT NULL DEFAULT '',
    author_email    VARCHAR(254) NOT NULL DEFAULT '',
    body            VARCHAR(4000) NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT project_feedback_body_check CHECK (length(btrim(body)) > 0)
);

CREATE INDEX IF NOT EXISTS project_feedback_project_created_idx
    ON project_feedback (project_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS project_feedback;
DROP TABLE IF EXISTS project_auditor_assignments;

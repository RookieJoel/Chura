-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100)  NOT NULL,
    description VARCHAR(2000) NOT NULL DEFAULT '',
    template_id VARCHAR(40)   NOT NULL,
    mode        VARCHAR(20)   NOT NULL,
    created_by  VARCHAR(100)  NOT NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT project_mode_check CHECK (mode IN ('general', 'se')),
    CONSTRAINT project_name_check CHECK (length(btrim(name)) > 0)
);

CREATE TABLE IF NOT EXISTS project_members (
    project_id UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id    VARCHAR(100) NOT NULL,
    role       VARCHAR(40)  NOT NULL,
    added_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, user_id),
    CONSTRAINT member_user_check CHECK (length(btrim(user_id)) > 0)
);
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_project_members_user_id;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;

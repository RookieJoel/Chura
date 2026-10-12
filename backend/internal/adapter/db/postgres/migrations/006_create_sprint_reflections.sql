-- +goose Up
-- One reflection per team member per Sprint; answers map a Template
-- reflection topic key (went_well, stop, start) to the member's answer.
CREATE TABLE IF NOT EXISTS sprint_reflections (
    id           BIGSERIAL    PRIMARY KEY,
    sprint_id    BIGINT       NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    project_id   UUID         NOT NULL REFERENCES projects(id),
    author_id    VARCHAR(100) NOT NULL,
    author_email VARCHAR(254) NOT NULL,
    template_id  VARCHAR(40)  NOT NULL,
    answers      JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT sprint_reflections_author_unique UNIQUE (sprint_id, author_id)
);

-- +goose Down
DROP TABLE IF EXISTS sprint_reflections;

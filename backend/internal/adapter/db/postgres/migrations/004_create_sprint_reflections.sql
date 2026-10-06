CREATE TABLE IF NOT EXISTS sprint_reflections (
    id BIGSERIAL PRIMARY KEY,
    sprint_id BIGINT NOT NULL REFERENCES sprints (id),
    author VARCHAR(120) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_sprint_reflections_sprint_id ON sprint_reflections (sprint_id);
CREATE INDEX IF NOT EXISTS idx_sprint_reflections_deleted_at ON sprint_reflections (deleted_at);

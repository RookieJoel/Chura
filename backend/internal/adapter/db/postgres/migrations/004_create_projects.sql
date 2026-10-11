-- +goose Up
CREATE TABLE IF NOT EXISTS projects (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(100)  NOT NULL,
    description       VARCHAR(2000) NOT NULL DEFAULT '',
    template_id       VARCHAR(40)   NOT NULL,
    mode              VARCHAR(20)   NOT NULL,
    created_by        VARCHAR(100)  NOT NULL,
    keycloak_group_id VARCHAR(64)   NOT NULL,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT now(),
    CONSTRAINT project_mode_check CHECK (mode IN ('general', 'se')),
    CONSTRAINT project_name_check CHECK (length(btrim(name)) > 0),
    CONSTRAINT project_group_unique UNIQUE (keycloak_group_id)
);

-- +goose Down
DROP TABLE IF EXISTS projects;

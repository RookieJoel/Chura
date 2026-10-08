CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS sprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(120) NOT NULL,

    team VARCHAR(120) NOT NULL,

    start_date DATE,

    end_date DATE,

    status VARCHAR(20) NOT NULL DEFAULT 'planned',

    CONSTRAINT sprint_status_check
        CHECK (status IN ('planned', 'active', 'completed')),

    CONSTRAINT sprint_date_check
        CHECK (
            end_date IS NULL
            OR start_date IS NULL
            OR end_date >= start_date
        )
);
-- Local dev seed for Board / Sprints / project overview.
-- Applied by scripts/seed-dev-postgres.sh (make seed-dev). Idempotent.
--
-- Project id must stay in sync with:
--   frontend/src/lib/dev-project.ts (SEEDED_DEV_PROJECT_ID)
--   frontend NEXT_PUBLIC_CHURA_PROJECT_ID

-- ---------------------------------------------------------------------------
-- Project
-- ---------------------------------------------------------------------------
INSERT INTO projects (
    id,
    name,
    description,
    template_id,
    mode,
    created_by,
    keycloak_group_id,
    created_at,
    updated_at
)
VALUES (
    '11111111-1111-4111-8111-111111111111',
    'Chura (dev)',
    'Seeded for local development. Membership is wired by the seed script via Keycloak.',
    'se',
    'se',
    'dev-seed',
    'dev-seed-chura-keycloak-group',
    TIMESTAMPTZ '2026-01-01 00:00:00+00',
    TIMESTAMPTZ '2026-01-01 00:00:00+00'
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    template_id = EXCLUDED.template_id,
    mode = EXCLUDED.mode;

-- ---------------------------------------------------------------------------
-- Sprints (fixed BIGSERIAL ids so Mongo work items can reference them)
-- ---------------------------------------------------------------------------
INSERT INTO sprints (
    id,
    name,
    team,
    start_date,
    end_date,
    status,
    project_id,
    created_at,
    updated_at
)
VALUES
    (
        1,
        'Sprint 0 — completed',
        'Platform',
        DATE '2026-01-06',
        DATE '2026-01-19',
        'completed',
        '11111111-1111-4111-8111-111111111111',
        TIMESTAMPTZ '2026-01-06 00:00:00+00',
        TIMESTAMPTZ '2026-01-19 00:00:00+00'
    ),
    (
        2,
        'Sprint 1 — active',
        'Platform',
        DATE '2026-01-20',
        DATE '2026-02-02',
        'active',
        '11111111-1111-4111-8111-111111111111',
        TIMESTAMPTZ '2026-01-20 00:00:00+00',
        TIMESTAMPTZ '2026-01-20 00:00:00+00'
    ),
    (
        3,
        'Sprint 2 — planned',
        'Platform',
        DATE '2026-02-03',
        DATE '2026-02-16',
        'planned',
        '11111111-1111-4111-8111-111111111111',
        TIMESTAMPTZ '2026-01-20 00:00:00+00',
        TIMESTAMPTZ '2026-01-20 00:00:00+00'
    )
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    team = EXCLUDED.team,
    start_date = EXCLUDED.start_date,
    end_date = EXCLUDED.end_date,
    status = EXCLUDED.status,
    project_id = EXCLUDED.project_id,
    deleted_at = NULL,
    updated_at = EXCLUDED.updated_at;

SELECT setval(
    pg_get_serial_sequence('sprints', 'id'),
    GREATEST((SELECT COALESCE(MAX(id), 1) FROM sprints), 1)
);

-- Local dev default project (matches frontend SEEDED_DEV_PROJECT_ID).
-- Applied by scripts/seed-dev-postgres.sh (make seed-dev). Idempotent.
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
    'Seeded for local development until project onboarding ships in the UI.',
    'se',
    'se',
    'dev-seed',
    'dev-seed-chura-keycloak-group',
    TIMESTAMPTZ '2026-01-01 00:00:00+00',
    TIMESTAMPTZ '2026-01-01 00:00:00+00'
)
ON CONFLICT (id) DO NOTHING;

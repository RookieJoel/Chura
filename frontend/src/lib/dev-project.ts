/**
 * Seeded local Project id from `make seed-dev` / backend/scripts/seed/dev_project.sql.
 * Board/Backlog/Sprints should use ProjectProvider selection; this remains an
 * optional last-resort fallback for local tooling only.
 */
export const SEEDED_DEV_PROJECT_ID =
  "11111111-1111-4111-8111-111111111111";

export function defaultProjectId(override?: string): string {
  const explicit = override?.trim();
  if (explicit) return explicit;
  const fromEnv = process.env.NEXT_PUBLIC_CHURA_PROJECT_ID?.trim();
  if (fromEnv) return fromEnv;
  return SEEDED_DEV_PROJECT_ID;
}

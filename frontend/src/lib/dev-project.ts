/**
 * Default project for local backlog/board work items until project onboarding exists.
 * Must stay in sync with backend/scripts/seed/dev_project.sql (make seed-dev).
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

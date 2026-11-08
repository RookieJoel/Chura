/** Team Members may mutate Sprint / Work Item execution surfaces. */
export function canMutateExecution(roles: string[] | null | undefined): boolean {
  return !!roles?.includes("member");
}

/** System Auditor role (may also hold member when dual-role). */
export function hasAuditorRole(roles: string[] | null | undefined): boolean {
  return !!roles?.includes("auditor");
}

type ClassValue = string | number | null | undefined | false;

/** Joins truthy class name fragments. No dedupe/merge — keep conflicting utility classes out of the same call. */
export function cn(...values: ClassValue[]): string {
  return values.filter(Boolean).join(" ");
}

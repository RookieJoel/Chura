import { SESSION_EXPIRED_MESSAGE } from "@/api/client";

const KNOWN_ERRORS: Record<string, string> = {
  forbidden: "You don't have permission to do that.",
  "not found": "This Project was not found or is no longer available.",
  unauthenticated: SESSION_EXPIRED_MESSAGE,
  "invalid input":
    "Some of the information you entered isn't valid. Check the fields and try again.",
  "invalid id": "That Project link isn't valid.",
  "invalid json body": "The request couldn't be understood. Please try again.",
  "you are not a member of this project":
    "You're not a member of this Project.",
  "user is already a member of this project":
    "That person is already a member of this Project.",
  conflict:
    "This change conflicts with the current Project state. Refresh and try again.",
  "service unavailable":
    "The service is temporarily unavailable. Please try again shortly.",
  "internal server error":
    "Something went wrong on our side. Please try again.",
  "this user does not exist":
    "This user doesn't exist. Check the email address and try again.",
  "team member role required":
    "You need the Team Member role to create or manage Projects.",
};

/** Turn API / thrown errors into short copy users can act on. */
export function userFacingError(reason: unknown, fallback: string): string {
  const raw =
    reason instanceof Error
      ? reason.message
      : typeof reason === "string"
        ? reason
        : "";
  if (!raw.trim()) return fallback;

  const stripped = raw.replace(/\s*\(\d{3}\)\s*$/u, "").trim();
  const known = KNOWN_ERRORS[stripped.toLowerCase()];
  if (known) return known;

  if (/[A-Z]/.test(stripped[0] ?? "") || stripped.includes(" ")) {
    return stripped.charAt(0).toUpperCase() + stripped.slice(1);
  }

  return fallback;
}

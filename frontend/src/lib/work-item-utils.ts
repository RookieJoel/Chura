import type { ProjectTemplate } from "@/api/templates";
import type {
  AcceptanceCriterion,
  WorkItem,
  WorkItemStatus,
  WorkItemType,
} from "@/types/work-item";

export const STATUS_ORDER: WorkItemStatus[] = [
  "todo",
  "in_progress",
  "review",
  "blocked",
  "done",
];

export const STATUS_LABELS: Record<WorkItemStatus, string> = {
  todo: "To do",
  in_progress: "In Progress",
  review: "Review",
  blocked: "Blocked",
  done: "Done",
};

export function groupByStatus(
  items: WorkItem[],
): Record<WorkItemStatus, WorkItem[]> {
  const grouped: Record<WorkItemStatus, WorkItem[]> = {
    todo: [],
    in_progress: [],
    review: [],
    blocked: [],
    done: [],
  };
  for (const item of items) grouped[item.status].push(item);
  return grouped;
}

export function sumPoints(items: WorkItem[]): number {
  return items.reduce((total, item) => total + item.storyPoints, 0);
}

export function toFrontendWorkItemType(type: string): WorkItemType | null {
  if (type === "user_story" || type === "story") return "story";
  if (type === "task" || type === "bug") return type;
  return null;
}

export function toApiWorkItemType(type: WorkItemType): string {
  return type === "story" ? "user_story" : type;
}

export function workItemTypesFromTemplate(
  template: ProjectTemplate | null | undefined,
): WorkItemType[] {
  const allowed = (template?.work_item_types ?? ["task"])
    .map(toFrontendWorkItemType)
    .filter((type): type is WorkItemType => type !== null);
  return allowed.length > 0 ? allowed : ["task"];
}

export function templateRequiresAcceptanceCriteria(
  template: ProjectTemplate | null | undefined,
  type: WorkItemType,
): boolean {
  const apiType = toApiWorkItemType(type);
  return (
    template?.validation_rules?.type_rules.some(
      (rule) => rule.type === apiType && rule.require_acceptance_criteria,
    ) ?? false
  );
}

export function templateUsesUserStoryDescription(
  template: ProjectTemplate | null | undefined,
  type: WorkItemType,
): boolean {
  const apiType = toApiWorkItemType(type);
  return (
    template?.validation_rules?.type_rules.some(
      (rule) =>
        rule.type === apiType && rule.description_format === "user_story",
    ) ?? false
  );
}

export function acceptanceCriteriaTexts(
  criteria: AcceptanceCriterion[] | string[] | undefined,
): string[] {
  if (!criteria) return [];
  return criteria
    .map((entry) => (typeof entry === "string" ? entry : entry.text))
    .map((text) => text.trim())
    .filter(Boolean);
}

export function parseAcceptanceCriteria(
  features?: Record<string, unknown>,
): AcceptanceCriterion[] {
  const raw = features?.acceptance_criteria;
  if (!Array.isArray(raw)) return [];
  return raw.flatMap((entry) => {
    if (typeof entry === "string" && entry.trim()) {
      return [{ text: entry.trim(), done: false }];
    }
    if (typeof entry === "object" && entry !== null && "text" in entry) {
      const record = entry as Record<string, unknown>;
      if (typeof record.text === "string" && record.text.trim()) {
        return [{ text: record.text.trim(), done: Boolean(record.done) }];
      }
    }
    return [];
  });
}

export function parseBlockedReason(
  features?: Record<string, unknown>,
): string | undefined {
  const reason = features?.blocked_reason;
  return typeof reason === "string" && reason.trim() ? reason.trim() : undefined;
}

export function formatValidationViolations(
  violations: Array<{ field?: string; message?: string }>,
): string {
  const messages = violations
    .map((violation) => violation.message?.trim() || "")
    .filter(Boolean);
  return messages.length > 0
    ? messages.join("; ")
    : "Work Item failed template validation";
}

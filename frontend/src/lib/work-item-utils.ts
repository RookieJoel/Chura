import type { WorkItem, WorkItemStatus } from "@/types/work-item";

export const STATUS_ORDER: WorkItemStatus[] = [
  "todo",
  "in_progress",
  "review",
  "done",
];

export const STATUS_LABELS: Record<WorkItemStatus, string> = {
  todo: "To do",
  in_progress: "In Progress",
  review: "Review",
  done: "Done",
};

export function groupByStatus(
  items: WorkItem[],
): Record<WorkItemStatus, WorkItem[]> {
  const grouped: Record<WorkItemStatus, WorkItem[]> = {
    todo: [],
    in_progress: [],
    review: [],
    done: [],
  };
  for (const item of items) grouped[item.status].push(item);
  return grouped;
}

export function sumPoints(items: WorkItem[]): number {
  return items.reduce((total, item) => total + item.storyPoints, 0);
}

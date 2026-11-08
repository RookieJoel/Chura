import type { WorkItem } from "@/types/work-item";
import { sumPoints } from "@/lib/work-item-utils";

/** Sums story points for Work Items assigned to each Sprint id. */
export function plannedPointsBySprint(
  items: WorkItem[],
): Record<string, number> {
  const bySprint: Record<string, WorkItem[]> = {};
  for (const item of items) {
    if (!item.sprintId) continue;
    (bySprint[item.sprintId] ??= []).push(item);
  }
  const totals: Record<string, number> = {};
  for (const [sprintId, sprintItems] of Object.entries(bySprint)) {
    totals[sprintId] = sumPoints(sprintItems);
  }
  return totals;
}

export function workItemCountBySprint(
  items: WorkItem[],
): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const item of items) {
    if (!item.sprintId) continue;
    counts[item.sprintId] = (counts[item.sprintId] ?? 0) + 1;
  }
  return counts;
}

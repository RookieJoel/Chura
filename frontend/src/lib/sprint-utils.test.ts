import { describe, expect, it } from "vitest";
import {
  plannedPointsBySprint,
  workItemCountBySprint,
} from "@/lib/sprint-utils";
import type { WorkItem } from "@/types/work-item";

function item(
  overrides: Partial<WorkItem> & Pick<WorkItem, "id" | "sprintId" | "storyPoints">,
): WorkItem {
  return {
    title: overrides.id,
    type: "task",
    status: "todo",
    blocked: false,
    priority: "medium",
    epic: "",
    labels: [],
    acceptanceCriteria: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("plannedPointsBySprint", () => {
  it("sums story points per assigned Sprint and ignores backlog items", () => {
    const items = [
      item({ id: "1", sprintId: "s1", storyPoints: 3 }),
      item({ id: "2", sprintId: "s1", storyPoints: 5 }),
      item({ id: "3", sprintId: "s2", storyPoints: 2 }),
      item({ id: "4", sprintId: null, storyPoints: 8 }),
    ];

    expect(plannedPointsBySprint(items)).toEqual({ s1: 8, s2: 2 });
    expect(workItemCountBySprint(items)).toEqual({ s1: 2, s2: 1 });
  });
});

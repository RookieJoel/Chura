import { describe, expect, it } from "vitest";
import type { ProjectTemplate } from "@/api/templates";
import {
  acceptanceCriteriaTexts,
  formatValidationViolations,
  groupByStatus,
  parseAcceptanceCriteria,
  parseBlockedReason,
  sumPoints,
  templateRequiresAcceptanceCriteria,
  templateUsesUserStoryDescription,
  toApiWorkItemType,
  workItemTypesFromTemplate,
} from "@/lib/work-item-utils";
import type { WorkItem } from "@/types/work-item";

function item(
  overrides: Partial<WorkItem> & Pick<WorkItem, "id" | "status" | "storyPoints">,
): WorkItem {
  return {
    title: overrides.id,
    type: "task",
    blocked: false,
    priority: "medium",
    epic: "",
    sprintId: "s1",
    labels: [],
    acceptanceCriteria: [],
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

const seTemplate: ProjectTemplate = {
  id: "se",
  name: "Software Engineering",
  mode: "se",
  work_item_types: ["user_story", "task", "bug"],
  roles: ["product_owner", "scrum_master", "developer"],
  default_role: "developer",
  creator_role: "product_owner",
  validation_rules: {
    required_fields: ["title"],
    story_points: { greater_than: 0 },
    type_rules: [
      {
        type: "user_story",
        description_format: "user_story",
        require_acceptance_criteria: true,
      },
    ],
  },
};

describe("board work-item utils", () => {
  it("groups items by status and sums story points per column", () => {
    const items = [
      item({ id: "1", status: "todo", storyPoints: 3 }),
      item({ id: "2", status: "todo", storyPoints: 2 }),
      item({ id: "3", status: "in_progress", storyPoints: 5 }),
      item({ id: "4", status: "done", storyPoints: 1 }),
    ];

    const grouped = groupByStatus(items);
    expect(grouped.todo).toHaveLength(2);
    expect(grouped.in_progress).toHaveLength(1);
    expect(grouped.review).toHaveLength(0);
    expect(grouped.blocked).toHaveLength(0);
    expect(grouped.done).toHaveLength(1);

    expect(sumPoints(items)).toBe(11);
    expect(sumPoints(grouped.todo)).toBe(5);
    expect(sumPoints(grouped.in_progress)).toBe(5);
    expect(sumPoints(grouped.done)).toBe(1);
  });
});

describe("template and planning helpers", () => {
  it("maps template work item types and SE validation rules", () => {
    expect(workItemTypesFromTemplate(null)).toEqual(["task"]);
    expect(workItemTypesFromTemplate(seTemplate)).toEqual([
      "story",
      "task",
      "bug",
    ]);
    expect(toApiWorkItemType("story")).toBe("user_story");
    expect(templateRequiresAcceptanceCriteria(seTemplate, "story")).toBe(true);
    expect(templateRequiresAcceptanceCriteria(seTemplate, "task")).toBe(false);
    expect(templateUsesUserStoryDescription(seTemplate, "story")).toBe(true);
  });

  it("parses acceptance criteria and blocked reason from features", () => {
    expect(
      parseAcceptanceCriteria({
        acceptance_criteria: [
          "plain criterion",
          { text: "structured", done: true },
          { text: "  ", done: false },
          42,
        ],
      }),
    ).toEqual([
      { text: "plain criterion", done: false },
      { text: "structured", done: true },
    ]);
    expect(acceptanceCriteriaTexts(["  keep ", "", "me"])).toEqual([
      "keep",
      "me",
    ]);
    expect(parseBlockedReason({ blocked_reason: " waiting on API " })).toBe(
      "waiting on API",
    );
    expect(parseBlockedReason({ blocked_reason: "   " })).toBeUndefined();
  });

  it("formats validation violations for the create/edit UI", () => {
    expect(
      formatValidationViolations([
        { field: "title", message: "title is required" },
        { field: "story_points", message: " story points must be > 0 " },
      ]),
    ).toBe("title is required; story points must be > 0");
    expect(formatValidationViolations([])).toBe(
      "Work Item failed template validation",
    );
  });
});

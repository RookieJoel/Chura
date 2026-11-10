// Mongo seed for the local Board. Idempotent upserts by _id.
// Applied by scripts/seed-dev-postgres.sh after Postgres sprints exist.
//
// Sprint ids match backend/scripts/seed/dev_project.sql (stringified BIGSERIAL).

const projectId = "11111111-1111-4111-8111-111111111111";
const now = ISODate("2026-01-20T12:00:00Z");

const items = [
  {
    _id: "seed-wi-001",
    project_id: projectId,
    sprint_id: "2",
    title: "Wire project selector into Board",
    description:
      "As a team member, I want the Board to use the selected project so I stop seeing seed-env 400s.",
    type: "user_story",
    status: "in_progress",
    priority: "high",
    story_points: 5,
    features: {
      acceptance_criteria: [
        "Board loads for selected project",
        "Invalid env id no longer used",
      ],
    },
    created_at: now,
    updated_at: now,
  },
  {
    _id: "seed-wi-002",
    project_id: projectId,
    sprint_id: "2",
    title: "Fix NEXT_PUBLIC_CHURA_PROJECT_ID default",
    description: "Replace placeholder project-1 with the seeded UUID.",
    type: "task",
    status: "to_do",
    priority: "high",
    story_points: 2,
    created_at: now,
    updated_at: now,
  },
  {
    _id: "seed-wi-003",
    project_id: projectId,
    sprint_id: "2",
    title: "Board card drag drops status",
    description: "Regression: moving a card updates work item status.",
    type: "bug",
    status: "review",
    priority: "medium",
    story_points: 3,
    created_at: now,
    updated_at: now,
  },
  {
    _id: "seed-wi-004",
    project_id: projectId,
    sprint_id: "2",
    title: "Document seed users and password",
    description: "tonnam / pim / jojo — password test1234",
    type: "task",
    status: "done",
    priority: "low",
    story_points: 1,
    created_at: now,
    updated_at: now,
  },
  {
    _id: "seed-wi-005",
    project_id: projectId,
    sprint_id: "1",
    title: "Ship project discovery API",
    description: "Completed in an earlier sprint.",
    type: "user_story",
    status: "done",
    priority: "medium",
    story_points: 8,
    created_at: ISODate("2026-01-10T12:00:00Z"),
    updated_at: ISODate("2026-01-18T12:00:00Z"),
  },
  {
    _id: "seed-wi-006",
    project_id: projectId,
    title: "Backlog: Auditor feedback polish",
    description: "Not yet planned into a sprint.",
    type: "task",
    status: "to_do",
    priority: "medium",
    story_points: 3,
    created_at: now,
    updated_at: now,
  },
];

const col = db.getSiblingDB("chura").work_items;
for (const item of items) {
  col.replaceOne({ _id: item._id }, item, { upsert: true });
}

print(`✔ seeded ${items.length} work items for project ${projectId}`);

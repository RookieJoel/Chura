import type {
  Member,
  Sprint,
  WorkItem,
  WorkItemComment,
} from "@/types/work-item";

/**
 * Fixtures backing the `src/api` layer until the real Work Item API lands.
 * Shapes mirror what that API is expected to return so swapping the fetch
 * implementation later shouldn't require touching any page or component.
 */

export const members: Member[] = [
  { id: "amorn", name: "Amorn Phanturat", initials: "AP", color: "#F79009" },
  { id: "suttanop", name: "Suttanop Chanah", initials: "SC", color: "#2E90FA" },
  { id: "jirapat", name: "Jirapat Dangkowkiew", initials: "JD", color: "#12B76A" },
  { id: "akarawat", name: "Akarawat Pongchaikaikiti", initials: "AK", color: "#9E5CF7" },
  { id: "akkharachai", name: "Akkharachai Yongsuwankul", initials: "AY", color: "#0BA5A4" },
];

export const sprints: Sprint[] = [
  {
    id: "sprint-4",
    name: "Sprint 4",
    startDate: "2026-09-15",
    endDate: "2026-09-29",
    status: "active",
  },
];

export const activeSprintId = "sprint-4";

export const workItems: WorkItem[] = [
  {
    id: "CHURA-101",
    type: "story",
    title: "As a team member, I want to create a Work Item so I can plan the Backlog",
    status: "review",
    blocked: false,
    priority: "high",
    storyPoints: 5,
    epic: "Backlog",
    assigneeId: "jirapat",
    reporterId: "jirapat",
    sprintId: activeSprintId,
    labels: [],
    acceptanceCriteria: [],
    createdAt: "2026-09-10",
    updatedAt: "2026-09-19",
  }
];

export const comments: WorkItemComment[] = [
  {
    id: "c1",
    workItemId: "CHURA-109",
    authorId: "jirapat",
    body: "Should the reason be a free-text field, or a short list of common blockers (waiting on review, dependency, environment)?",
    createdAt: "2026-09-19T10:42:00Z",
  },
  {
    id: "c2",
    workItemId: "CHURA-109",
    authorId: "akarawat",
    body: "Let's start with free text — we can add presets later if patterns show up in Sprint reviews.",
    createdAt: "2026-09-20T09:15:00Z",
  },
];

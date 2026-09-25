export type WorkItemType = "story" | "task" | "bug";

export type WorkItemStatus = "todo" | "in_progress" | "review" | "done";

export type Priority = "highest" | "high" | "medium" | "low";

export interface Member {
  id: string;
  name: string;
  initials: string;
  color: string;
}

export interface AcceptanceCriterion {
  text: string;
  done: boolean;
}

export interface WorkItem {
  /** Ticket key, e.g. "CHURA-109". The real API may split this into a UUID `id` plus a human-readable `key`. */
  id: string;
  projectId?: string;
  type: WorkItemType;
  title: string;
  description?: string;
  status: WorkItemStatus;
  blocked: boolean;
  blockedReason?: string;
  priority: Priority;
  storyPoints: number;
  epic: string;
  assigneeId?: string;
  reporterId?: string;
  /** null = unplanned, still sitting in the Backlog */
  sprintId: string | null;
  labels: string[];
  features?: Record<string, unknown>;
  acceptanceCriteria: AcceptanceCriterion[];
  createdAt: string;
  updatedAt: string;
}

export interface Sprint {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  status: "active" | "planned" | "completed";
}

export interface WorkItemComment {
  id: string;
  workItemId: string;
  authorId: string;
  body: string;
  createdAt: string;
}

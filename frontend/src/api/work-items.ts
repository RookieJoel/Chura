import { comments, workItems } from "@/lib/mock-data";
import type { WorkItem, WorkItemComment } from "@/types/work-item";

// TODO: swap fixtures for real calls to the backend/ Fiber API once it exists.

export interface ListWorkItemsParams {
  /** null = unplanned Backlog items only. Omit to get every Work Item. */
  sprintId?: string | null;
}

export async function listWorkItems(
  params: ListWorkItemsParams = {},
): Promise<WorkItem[]> {
  if (params.sprintId === undefined) return workItems;
  return workItems.filter((item) => item.sprintId === params.sprintId);
}

export async function getWorkItem(id: string): Promise<WorkItem | null> {
  return workItems.find((item) => item.id === id) ?? null;
}

export async function listComments(
  workItemId: string,
): Promise<WorkItemComment[]> {
  return comments.filter((comment) => comment.workItemId === workItemId);
}

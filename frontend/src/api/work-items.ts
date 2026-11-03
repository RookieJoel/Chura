import type {
  WorkItem,
  WorkItemComment,
  WorkItemType,
  WorkItemStatus,
  Priority,
} from "@/types/work-item";
import { apiRequest } from "@/api/client";

const DEFAULT_PROJECT_ID =
  process.env.NEXT_PUBLIC_CHURA_PROJECT_ID;

interface ApiWorkItem {
  id: string;
  project_id?: string;
  sprint_id?: string;
  title: string;
  description?: string;
  type: string;
  status: string;
  priority: string;
  assignee_id?: string;
  reporter_id?: string;
  story_points?: number;
  features?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

function toWorkItem(item: ApiWorkItem): WorkItem {
  return {
    id: item.id,
    projectId: item.project_id,
    title: item.title,
    description: item.description,
    type: item.type === "user_story" ? "story" : (item.type as WorkItemType),
    status: item.status === "to_do" ? "todo" : (item.status as WorkItemStatus),
    blocked: item.status === "blocked",
    priority: item.priority === "critical" ? "highest" : (item.priority as Priority),
    storyPoints: item.story_points ?? 0,
    epic: "",
    assigneeId: item.assignee_id,
    reporterId: item.reporter_id,
    sprintId: item.sprint_id || null,
    labels: [],
    features: item.features,
    acceptanceCriteria: [],
    createdAt: item.created_at,
    updatedAt: item.updated_at,
  };
}

export interface WorkItemInput {
  projectId?: string;
  sprintId?: string | null;
  title: string;
  description?: string;
  type?: WorkItemType;
  status?: WorkItemStatus;
  priority?: Priority;
  assigneeId?: string;
  reporterId?: string;
  storyPoints?: number;
  features?: Record<string, unknown>;
}

export type CreateWorkItemInput = WorkItemInput;
export interface UpdateWorkItemInput extends WorkItemInput {
  id: string;
}

function toPayload(input: WorkItemInput, id?: string) {
  const projectId = input.projectId ?? DEFAULT_PROJECT_ID;
  if (!projectId) throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
  return {
    ...(id ? { id } : {}),
    project_id: projectId,
    sprint_id: input.sprintId === undefined ? undefined : input.sprintId,
    title: input.title,
    description: input.description ?? "",
    type: input.type === "story" ? "user_story" : input.type ?? "task",
    status: input.status === "todo" ? "to_do" : input.status ?? "to_do",
    priority: input.priority === "highest" ? "critical" : input.priority ?? "medium",
    assignee_id: input.assigneeId ?? "",
    reporter_id: input.reporterId ?? "",
    story_points: input.storyPoints ?? 0,
    features: input.features ?? {},
  };
}

export interface ListWorkItemsParams {
  sprintId?: string | null;
  projectId?: string;
}

export async function listWorkItems(
  params: ListWorkItemsParams = {},
): Promise<WorkItem[]> {
  const query = new URLSearchParams();
  if (params.sprintId !== undefined) {
    if (params.sprintId) {
      query.set("sprint_id", params.sprintId);
    } else {
      const projectId = params.projectId ?? DEFAULT_PROJECT_ID;
      if (!projectId) throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
      query.set("project_id", projectId);
    }
  } else {
    const projectId = params.projectId ?? DEFAULT_PROJECT_ID;
    if (!projectId) throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
    query.set("project_id", projectId);
  }
  const items = await apiRequest<ApiWorkItem[]>(
    `/work-items?${query.toString()}`,
  );
  return items.map(toWorkItem).filter((item) =>
    params.sprintId === null ? item.sprintId === null : true,
  );
}

export async function getWorkItem(id: string): Promise<WorkItem | null> {
  try {
    return toWorkItem(await apiRequest<ApiWorkItem>(`/work-items/${id}`));
  } catch (error) {
    if (error instanceof Error && error.message === "not found") return null;
    throw error;
  }
}

export async function createWorkItem(input: CreateWorkItemInput): Promise<WorkItem> {
  return toWorkItem(await apiRequest<ApiWorkItem>("/work-items", {
    method: "POST",
    body: JSON.stringify(toPayload(input)),
  }));
}

export async function updateWorkItem(input: UpdateWorkItemInput): Promise<WorkItem> {
  return toWorkItem(await apiRequest<ApiWorkItem>(`/work-items/${input.id}`, {
    method: "PUT",
    body: JSON.stringify(toPayload(input, input.id)),
  }));
}

export async function deleteWorkItem(id: string): Promise<void> {
  await apiRequest<void>(`/work-items/${id}`, { method: "DELETE" });
}

export async function listComments(
  workItemId: string,
): Promise<WorkItemComment[]> {
  void workItemId;
  return [];
}

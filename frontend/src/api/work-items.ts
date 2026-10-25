import { comments } from "@/lib/mock-data";
import { defaultProjectId } from "@/lib/dev-project";
import type { WorkItem, WorkItemComment, WorkItemType, WorkItemStatus, Priority } from "@/types/work-item";

const WORK_ITEMS_RPC_URL =
  process.env.NEXT_PUBLIC_WORK_ITEM_RPC_URL ??
  "ws://127.0.0.1:8083/ws/work-items";
const RPC_TIMEOUT_MS = 10_000;

type RpcWorkItem = Record<string, unknown>;

export interface WorkItemInput {
  projectId?: string;
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

interface RpcResponse {
  error?: string;
  message?: string;
  work_item?: RpcWorkItem;
  work_items?: RpcWorkItem[];
  data?: RpcWorkItem | RpcWorkItem[];
}

function pascalToSnake(name: string): string {
  if (name === "ID") return "id";
  return name
    .replace(/([A-Z]+)([A-Z][a-z])/g, "$1_$2")
    .replace(/([a-z\d])([A-Z])/g, "$1_$2")
    .toLowerCase();
}

function readField<T>(item: RpcWorkItem, name: string): T | undefined {
  const camel = name.charAt(0).toLowerCase() + name.slice(1);
  const snake = pascalToSnake(name);
  const value = item[name] ?? item[camel] ?? item[snake];
  return value as T | undefined;
}

function toWorkItem(item: RpcWorkItem): WorkItem {
  const status = readField<string>(item, "Status");
  const type = readField<string>(item, "Type");
  const priority = readField<string>(item, "Priority");

  return {
    id: readField<string>(item, "ID") ?? "",
    projectId: readField<string>(item, "ProjectID"),
    title: readField<string>(item, "Title") ?? "",
    description: readField<string>(item, "Description"),
    type: type === "user_story" ? "story" : (type as WorkItemType) ?? "task",
    status: status === "to_do" ? "todo" : (status as WorkItemStatus) ?? "todo",
    blocked: status === "blocked",
    priority: priority === "critical" ? "highest" : (priority as Priority) ?? "medium",
    storyPoints: readField<number>(item, "StoryPoints") ?? 0,
    epic: "",
    assigneeId: readField<string>(item, "AssigneeID"),
    reporterId: readField<string>(item, "ReporterID"),
    sprintId: null,
    labels: [],
    features: readField<Record<string, unknown>>(item, "Features"),
    acceptanceCriteria: [],
    createdAt: readField<string>(item, "CreatedAt") ?? new Date().toISOString(),
    updatedAt: readField<string>(item, "UpdatedAt") ?? new Date().toISOString(),
  };
}

function toRpcWorkItem(input: WorkItemInput, projectId: string, id?: string) {
  return {
    ...(id ? { id } : {}),
    project_id: projectId,
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

async function callWorkItemRpc(payload: Record<string, unknown>): Promise<RpcResponse> {
  const Socket = globalThis.WebSocket;
  if (!Socket) throw new Error("WebSocket is not available in this runtime");

  return new Promise((resolve, reject) => {
    const socket = new Socket(WORK_ITEMS_RPC_URL);
    const timeout = setTimeout(() => {
      socket.close();
      reject(new Error("Work Item RPC timed out"));
    }, RPC_TIMEOUT_MS);

    const fail = (error: Error) => {
      clearTimeout(timeout);
      reject(error);
    };

    socket.onopen = () => socket.send(JSON.stringify(payload));
    socket.onerror = () => fail(new Error("Work Item RPC connection failed"));
    socket.onclose = () => clearTimeout(timeout);
    socket.onmessage = (event) => {
      try {
        const response = JSON.parse(String(event.data)) as RpcResponse;
        clearTimeout(timeout);
        socket.close();
        if (response.error || response.message) {
          reject(new Error(response.error ?? response.message));
          return;
        }
        resolve(response);
      } catch {
        fail(new Error("Invalid Work Item RPC response"));
      }
    };
  });
}

export interface ListWorkItemsParams {
  /** null = unplanned Backlog items only. Omit to get every Work Item. */
  sprintId?: string | null;
  projectId?: string;
}

export async function listWorkItems(
  params: ListWorkItemsParams = {},
): Promise<WorkItem[]> {
  const projectId = defaultProjectId(params.projectId);
  const response = await callWorkItemRpc({
    operation: "list",
    project_id: projectId,
  });
  const items = response.work_items ?? (Array.isArray(response.data) ? response.data : []);
  const workItems = items.map(toWorkItem);
  return params.sprintId === undefined
    ? workItems
    : workItems.filter((item) => item.sprintId === params.sprintId);
}

export async function getWorkItem(id: string): Promise<WorkItem | null> {
  const response = await callWorkItemRpc({ operation: "get", id });
  const item = response.work_item ?? (response.data && !Array.isArray(response.data) ? response.data : undefined);
  return item ? toWorkItem(item) : null;
}

export async function createWorkItem(input: CreateWorkItemInput): Promise<WorkItem> {
  const projectId = defaultProjectId(input.projectId);
  const response = await callWorkItemRpc({
    operation: "create",
    work_item: toRpcWorkItem(input, projectId),
  });
  const item = response.work_item ?? (response.data && !Array.isArray(response.data) ? response.data : undefined);
  if (!item) throw new Error("Create Work Item response did not include a work item");
  return toWorkItem(item);
}

export async function updateWorkItem(input: UpdateWorkItemInput): Promise<WorkItem> {
  const projectId = defaultProjectId(input.projectId);
  const response = await callWorkItemRpc({
    operation: "update",
    work_item: toRpcWorkItem(input, projectId, input.id),
  });
  const item = response.work_item ?? (response.data && !Array.isArray(response.data) ? response.data : undefined);
  if (!item) throw new Error("Update Work Item response did not include a work item");
  return toWorkItem(item);
}

export async function deleteWorkItem(id: string): Promise<void> {
  await callWorkItemRpc({ operation: "delete", id });
}

export async function listComments(
  workItemId: string,
): Promise<WorkItemComment[]> {
  return comments.filter((comment) => comment.workItemId === workItemId);
}

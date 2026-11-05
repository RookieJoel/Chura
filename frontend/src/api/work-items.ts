import type {
  WorkItem,
  WorkItemComment,
  WorkItemType,
  WorkItemStatus,
  Priority,
} from "@/types/work-item";

const DEFAULT_PROJECT_ID = process.env.NEXT_PUBLIC_CHURA_PROJECT_ID;
const DEFAULT_WS_URL = process.env.NEXT_PUBLIC_WORK_ITEM_RPC_URL;

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

interface ApiWorkItemInput {
  id?: string;
  project_id: string;
  sprint_id?: string | null;
  title: string;
  description: string;
  type: string;
  status: string;
  priority: string;
  assignee_id: string;
  reporter_id: string;
  story_points: number;
  features: Record<string, unknown>;
}

interface WorkItemResponse {
  request_id: string;
  operation: string;
  work_item?: ApiWorkItem;
  work_items?: ApiWorkItem[];
  valid?: boolean;
  error?: string;
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

function toPayload(input: WorkItemInput, id?: string): ApiWorkItemInput {
  const projectId = input.projectId ?? DEFAULT_PROJECT_ID;
  if (!projectId) throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
  return {
    ...(id ? { id } : {}),
    project_id: projectId,
    sprint_id: input.sprintId,
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

function websocketURL(accessToken: string): string {
  const baseURL = DEFAULT_WS_URL;
  if (baseURL) {
    const url = new URL(baseURL);
    url.searchParams.set("access_token", accessToken);
    return url.toString();
  }
  if (typeof window === "undefined") {
    throw new Error("NEXT_PUBLIC_WORK_ITEM_RPC_URL is not configured");
  }
  const url = new URL(
    `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/ws/work-items`,
  );
  url.searchParams.set("access_token", accessToken);
  return url.toString();
}

async function openWorkItemSocket(): Promise<WebSocket> {
  const tokenResponse = await fetch("/api/ws-token", { cache: "no-store" });
  if (!tokenResponse.ok) {
    throw new Error(`Unable to authenticate WorkItem socket (${tokenResponse.status})`);
  }
  const token = (await tokenResponse.json()) as { access_token?: string };
  if (!token.access_token) throw new Error("WorkItem socket token is missing");
  const accessToken = token.access_token;
  return new Promise((resolve, reject) => {
    const socket = new WebSocket(websocketURL(accessToken), "chura-auth");
    socket.onopen = () => resolve(socket);
    socket.onerror = () => reject(new Error("Unable to connect to WorkItem socket"));
    socket.onclose = (event) => {
      reject(new Error(`WorkItem socket closed (${event.code}): ${event.reason || "handshake failed"}`));
    };
  });
}

function requestID(): string {
  return typeof crypto.randomUUID === "function"
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random()}`;
}

async function sendWorkItemMessage(
  message: Record<string, unknown>,
  retryRead = false,
): Promise<WorkItemResponse> {
  const request_id = requestID();
  let attempts = retryRead ? 2 : 1;
  while (attempts > 0) {
    attempts -= 1;
    try {
      const socket = await openWorkItemSocket();
      return await new Promise((resolve, reject) => {
        const request = { ...message, request_id };
        socket.onmessage = (event) => {
          const response = JSON.parse(event.data) as WorkItemResponse;
          if (response.request_id !== request_id) return;
          socket.close();
          if (response.error) reject(new Error(response.error));
          else resolve(response);
        };
        socket.onclose = () => reject(new Error("WorkItem socket closed"));
        socket.onerror = () => reject(new Error("WorkItem socket failed"));
        socket.send(JSON.stringify(request));
      });
    } catch (error) {
      if (attempts === 0) throw error;
    }
  }
  throw new Error("WorkItem request failed");
}

export interface ListWorkItemsParams {
  sprintId?: string | null;
  projectId?: string;
}

export async function listWorkItems(params: ListWorkItemsParams = {}): Promise<WorkItem[]> {
  const projectId = params.projectId ?? DEFAULT_PROJECT_ID;
  if (!params.sprintId && !projectId) {
    throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
  }
  const response = await sendWorkItemMessage({
    operation: "list",
    project_id: projectId,
    sprint_id: params.sprintId ?? undefined,
  }, true);
  return (response.work_items ?? []).map(toWorkItem).filter((item) =>
    params.sprintId === null ? item.sprintId === null : true,
  );
}

export async function getWorkItem(id: string): Promise<WorkItem | null> {
  try {
    const response = await sendWorkItemMessage({ operation: "get", id }, true);
    return response.work_item ? toWorkItem(response.work_item) : null;
  } catch (error) {
    if (error instanceof Error && error.message === "not found") return null;
    throw error;
  }
}

export async function validateWorkItem(input: CreateWorkItemInput): Promise<boolean> {
  const response = await sendWorkItemMessage({
    operation: "validate",
    work_item: toPayload(input),
  });
  return response.valid === true;
}

export async function createWorkItem(input: CreateWorkItemInput): Promise<WorkItem> {
  const response = await sendWorkItemMessage({ operation: "create", work_item: toPayload(input) });
  if (!response.work_item) throw new Error("WorkItem create response is missing an item");
  return toWorkItem(response.work_item);
}

export async function updateWorkItem(input: UpdateWorkItemInput): Promise<WorkItem> {
  const response = await sendWorkItemMessage({
    operation: "update",
    id: input.id,
    work_item: toPayload(input, input.id),
  });
  if (!response.work_item) throw new Error("WorkItem update response is missing an item");
  return toWorkItem(response.work_item);
}

export async function deleteWorkItem(id: string): Promise<void> {
  await sendWorkItemMessage({ operation: "delete", id });
}

export async function listComments(workItemId: string): Promise<WorkItemComment[]> {
  void workItemId;
  return [];
}

import { SESSION_EXPIRED_MESSAGE, sameOriginRequest } from "@/api/client";
import {
  parseAcceptanceCriteria,
  parseBlockedReason,
  toApiWorkItemType,
  toFrontendWorkItemType,
} from "@/lib/work-item-utils";
import type {
  AcceptanceCriterion,
  WorkItem,
  WorkItemComment,
  WorkItemType,
  WorkItemStatus,
  Priority,
} from "@/types/work-item";

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

export interface WorkItemValidationAttributes {
  type: string;
  title: string;
  description?: string;
  story_points?: number;
  acceptance_criteria?: string[];
}

export interface ValidationViolation {
  field: string;
  message: string;
}

export interface ValidationResult {
  valid: boolean;
  violations: ValidationViolation[];
}

function mapStatus(status: string): WorkItemStatus {
  if (status === "to_do") return "todo";
  if (status === "blocked") return "blocked";
  if (
    status === "todo" ||
    status === "in_progress" ||
    status === "review" ||
    status === "done"
  ) {
    return status;
  }
  return "todo";
}

function toApiStatus(status?: WorkItemStatus): string {
  if (status === "todo") return "to_do";
  return status ?? "to_do";
}

function toWorkItem(item: ApiWorkItem): WorkItem {
  const type = toFrontendWorkItemType(item.type) ?? "task";
  const status = mapStatus(item.status);
  return {
    id: item.id,
    projectId: item.project_id,
    title: item.title,
    description: item.description,
    type,
    status,
    blocked: status === "blocked",
    blockedReason: parseBlockedReason(item.features),
    priority: item.priority === "critical" ? "highest" : (item.priority as Priority),
    storyPoints: item.story_points ?? 0,
    epic: "",
    assigneeId: item.assignee_id,
    reporterId: item.reporter_id,
    sprintId: item.sprint_id || null,
    labels: [],
    features: item.features,
    acceptanceCriteria: parseAcceptanceCriteria(item.features),
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
  acceptanceCriteria?: AcceptanceCriterion[] | string[];
  blockedReason?: string;
}

export type CreateWorkItemInput = WorkItemInput;
export interface UpdateWorkItemInput extends WorkItemInput {
  id: string;
}

function requireProjectId(projectId: string | null | undefined): string {
  const id = projectId?.trim();
  if (!id) throw new Error("A Project must be selected");
  return id;
}

function buildFeatures(input: WorkItemInput): Record<string, unknown> {
  const features: Record<string, unknown> = { ...(input.features ?? {}) };

  if (input.acceptanceCriteria !== undefined) {
    const criteria = input.acceptanceCriteria
      .map((entry) => {
        if (typeof entry === "string") {
          const text = entry.trim();
          return text ? { text, done: false } : null;
        }
        const text = entry.text.trim();
        return text ? { text, done: Boolean(entry.done) } : null;
      })
      .filter((entry): entry is { text: string; done: boolean } => entry !== null);
    if (criteria.length > 0) features.acceptance_criteria = criteria;
    else delete features.acceptance_criteria;
  }

  const status = input.status ?? "todo";
  if (status === "blocked") {
    const reason = input.blockedReason?.trim();
    if (reason) features.blocked_reason = reason;
  } else {
    delete features.blocked_reason;
  }

  return features;
}

function toPayload(input: WorkItemInput, id?: string): ApiWorkItemInput {
  const projectId = requireProjectId(input.projectId);
  return {
    ...(id ? { id } : {}),
    project_id: projectId,
    sprint_id: input.sprintId,
    title: input.title,
    description: input.description ?? "",
    type: toApiWorkItemType(input.type ?? "task"),
    status: toApiStatus(input.status),
    priority: input.priority === "highest" ? "critical" : input.priority ?? "medium",
    assignee_id: input.assigneeId ?? "",
    reporter_id: input.reporterId ?? "",
    story_points: input.storyPoints ?? 0,
    features: buildFeatures(input),
  };
}

function websocketURL(): string {
  const baseURL = DEFAULT_WS_URL;
  if (baseURL) {
    const url = new URL(baseURL);
    // Prefer 127.0.0.1 over localhost so the browser does not attach Auth.js
    // cookies (same registrable host) and Kong returns HTTP 431.
    if (url.hostname === "localhost") {
      url.hostname = "127.0.0.1";
    }
    return url.toString();
  }
  if (typeof window === "undefined") {
    throw new Error("NEXT_PUBLIC_WORK_ITEM_RPC_URL is not configured");
  }
  return fallbackWebsocketURL(window.location);
}

// Without an explicit NEXT_PUBLIC_WORK_ITEM_RPC_URL, a local dev page talks to
// Kong on 127.0.0.1:8083 (not localhost, so Auth.js cookies are not attached
// and Kong does not answer 431); any other host is served same-origin.
export function fallbackWebsocketURL(
  location: Pick<Location, "protocol" | "hostname" | "host">,
): string {
  const scheme = location.protocol === "https:" ? "wss:" : "ws:";
  const isLocal =
    location.hostname === "localhost" || location.hostname === "127.0.0.1";
  const host = isLocal ? "127.0.0.1:8083" : location.host;
  return `${scheme}//${host}/ws/work-items`;
}

async function openWorkItemSocket(): Promise<WebSocket> {
  const tokenResponse = await fetch("/api/ws-token", { cache: "no-store" });
  if (!tokenResponse.ok) {
    if (tokenResponse.status === 401) throw new Error(SESSION_EXPIRED_MESSAGE);
    throw new Error(`Unable to authenticate WorkItem socket (${tokenResponse.status})`);
  }
  const token = (await tokenResponse.json()) as { access_token?: string };
  if (!token.access_token) throw new Error("WorkItem socket token is missing");
  const accessToken = token.access_token;
  return new Promise((resolve, reject) => {
    // Token rides in the Sec-WebSocket-Protocol header (auth.<jwt>), which the
    // backend already accepts. Keep "chura-auth" for upgrade negotiation.
    const socket = new WebSocket(websocketURL(), ["chura-auth", `auth.${accessToken}`]);
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
  const projectId = params.projectId?.trim() ?? "";
  if (!projectId) {
    throw new Error("A Project must be selected");
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

function isValidationResult(value: unknown): value is ValidationResult {
  if (typeof value !== "object" || value === null) return false;
  const result = value as Record<string, unknown>;
  return (
    typeof result.valid === "boolean" &&
    Array.isArray(result.violations) &&
    result.violations.every((violation) => {
      if (typeof violation !== "object" || violation === null) return false;
      const entry = violation as Record<string, unknown>;
      return typeof entry.field === "string" && typeof entry.message === "string";
    })
  );
}

/** Template-aware validation via REST. Persists nothing; valid=false is still 200. */
export async function validateWorkItemAttributes(
  projectId: string,
  attrs: WorkItemValidationAttributes,
): Promise<ValidationResult> {
  const result = await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/work-items/validate`,
    {
      method: "POST",
      body: JSON.stringify(attrs),
    },
  );
  if (!isValidationResult(result)) {
    throw new Error("Invalid work item validation response");
  }
  return result;
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

interface ApiWorkItemComment {
  id: string;
  work_item_id: string;
  author_user_id: string;
  author_name: string;
  author_email: string;
  body: string;
  created_at: string;
}

function isApiWorkItemComment(value: unknown): value is ApiWorkItemComment {
  if (typeof value !== "object" || value === null) return false;
  const item = value as Record<string, unknown>;
  return (
    typeof item.id === "string" &&
    typeof item.work_item_id === "string" &&
    typeof item.author_user_id === "string" &&
    typeof item.author_name === "string" &&
    typeof item.author_email === "string" &&
    typeof item.body === "string" &&
    typeof item.created_at === "string"
  );
}

function toWorkItemComment(item: ApiWorkItemComment): WorkItemComment {
  return {
    id: item.id,
    workItemId: item.work_item_id,
    authorUserId: item.author_user_id,
    authorName: item.author_name,
    authorEmail: item.author_email,
    body: item.body,
    createdAt: item.created_at,
  };
}

/** Newest-first append-only comments from Session 4 REST. */
export async function listComments(workItemId: string): Promise<WorkItemComment[]> {
  const items = await sameOriginRequest<unknown>(
    `/api/v1/work-items/${encodeURIComponent(workItemId)}/comments`,
  );
  if (!Array.isArray(items) || !items.every(isApiWorkItemComment)) {
    throw new Error("Invalid Work Item comments response");
  }
  return items.map(toWorkItemComment);
}

export async function createComment(
  workItemId: string,
  body: string,
): Promise<WorkItemComment> {
  const comment = await sameOriginRequest<unknown>(
    `/api/v1/work-items/${encodeURIComponent(workItemId)}/comments`,
    { method: "POST", body: JSON.stringify({ body }) },
  );
  if (!isApiWorkItemComment(comment)) {
    throw new Error("Invalid Work Item comment create response");
  }
  return toWorkItemComment(comment);
}

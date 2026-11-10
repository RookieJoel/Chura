import { sameOriginRequest } from "@/api/client";
import type { Member } from "@/types/work-item";

const projectId = process.env.NEXT_PUBLIC_CHURA_PROJECT_ID;

export interface CreateProjectInput {
  name: string;
  description: string;
  template_id: string;
}

interface ApiProjectMember {
  user_id: string;
  name: string;
  email: string;
  role: string;
}

interface ApiProject {
  id: string;
  name: string;
  description: string;
  template_id: string;
  mode: string;
  created_by: string;
  members: ApiProjectMember[];
}

export interface ProjectSummary {
  id: string;
  name: string;
  description: string;
  template_id: string;
  mode: string;
}

export interface ProjectMember extends Member {
  email: string;
  role: string;
}

export interface ProjectDetails extends ProjectSummary {
  created_by: string;
  members: ProjectMember[];
}

export interface CurrentProject {
  id: string;
  name: string;
  members: Member[];
}

interface ProjectListPage {
  items: ProjectSummary[];
  page: number;
  page_size: number;
  total: number;
}

const projectPageSize = 30;

async function requestProjectPage(page: number): Promise<ProjectListPage> {
  const result = await sameOriginRequest<ProjectListPage>(
    `/api/v1/projects?page=${page}&page_size=${projectPageSize}`,
  );
  if (
    !result ||
    !Array.isArray(result.items) ||
    result.page !== page ||
    result.page_size !== projectPageSize ||
    !Number.isSafeInteger(result.total) ||
    result.total < 0 ||
    result.items.length > projectPageSize
  ) {
    throw new Error("Invalid Project discovery response");
  }
  return result;
}

export async function listProjects(): Promise<ProjectSummary[]> {
  const firstPage = await requestProjectPage(1);
  const projects = [...firstPage.items];
  const pageCount = Math.ceil(firstPage.total / projectPageSize);

  for (let page = 2; page <= pageCount; page += 1) {
    const result = await requestProjectPage(page);
    projects.push(...result.items);
  }

  if (projects.length !== firstPage.total) {
    throw new Error("Project discovery returned an incomplete result");
  }
  return projects;
}

/** @deprecated Prefer getProject(selectedProjectId). Kept for local seed fallbacks only. */
export async function getCurrentProject(): Promise<CurrentProject> {
  if (!projectId) {
    throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
  }
  const project = await getProject(projectId);
  return {
    id: project.id,
    name: project.name,
    members: project.members.map((member) => ({
      id: member.id,
      name: member.name,
      initials: member.initials,
      color: member.color,
    })),
  };
}

export async function createProject(
  input: CreateProjectInput,
): Promise<ProjectSummary> {
  const project = await sameOriginRequest<ProjectSummary>("/api/v1/projects", {
    method: "POST",
    body: JSON.stringify(input),
  });
  if (
    !project ||
    typeof project.id !== "string" ||
    !project.id ||
    typeof project.name !== "string" ||
    typeof project.description !== "string" ||
    typeof project.template_id !== "string" ||
    typeof project.mode !== "string"
  ) {
    throw new Error("Invalid Project creation response");
  }
  return project;
}

export async function getProject(projectId: string): Promise<ProjectDetails> {
  const project = await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}`,
  );
  return normalizeProject(project);
}

function isApiProjectMember(value: unknown): value is ApiProjectMember {
  if (typeof value !== "object" || value === null) return false;
  const member = value as Record<string, unknown>;
  return (
    typeof member.user_id === "string" &&
    typeof member.name === "string" &&
    typeof member.email === "string" &&
    typeof member.role === "string"
  );
}

function isApiProject(value: unknown): value is ApiProject {
  if (typeof value !== "object" || value === null) return false;
  const project = value as Record<string, unknown>;
  return (
    typeof project.id === "string" &&
    typeof project.name === "string" &&
    typeof project.description === "string" &&
    typeof project.template_id === "string" &&
    typeof project.mode === "string" &&
    typeof project.created_by === "string" &&
    Array.isArray(project.members) &&
    project.members.every(isApiProjectMember)
  );
}

function normalizeProject(value: unknown): ProjectDetails {
  if (!isApiProject(value)) {
    throw new Error("Invalid Project detail response");
  }
  const project = value;
  return {
    ...project,
    members: project.members.map((member) => {
      const name = member.name || member.email || member.user_id;
      return {
        id: member.user_id,
        email: member.email,
        name,
        role: member.role,
        initials: name
          .split(/\s+/)
          .map((part) => part[0] ?? "")
          .join("")
          .slice(0, 2)
          .toUpperCase(),
        color: "#64748B",
      };
    }),
  };
}

export async function addProjectMember(
  projectId: string,
  email: string,
): Promise<ProjectDetails> {
  return sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/members`,
    { method: "POST", body: JSON.stringify({ email }) },
  ).then(normalizeProject);
}

export async function assignProjectRole(
  projectId: string,
  userId: string,
  role: string,
): Promise<ProjectDetails> {
  return sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/members/${encodeURIComponent(userId)}/role`,
    { method: "PUT", body: JSON.stringify({ role }) },
  ).then(normalizeProject);
}

export async function removeProjectMember(
  projectId: string,
  userId: string,
): Promise<ProjectDetails> {
  return sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/members/${encodeURIComponent(userId)}`,
    { method: "DELETE" },
  ).then(normalizeProject);
}

export interface ProjectAuditorAssignment {
  project_id: string;
  auditor_user_id: string;
  auditor_name: string;
  auditor_email: string;
  assigned_at: string;
}

export interface ProjectFeedback {
  id: string;
  project_id: string;
  author_user_id: string;
  author_name: string;
  author_email: string;
  body: string;
  created_at: string;
}

export interface UpdateProjectInput {
  name?: string;
  description?: string;
  template_id?: string;
}

function isAuditorAssignment(value: unknown): value is ProjectAuditorAssignment {
  if (typeof value !== "object" || value === null) return false;
  const item = value as Record<string, unknown>;
  return (
    typeof item.project_id === "string" &&
    typeof item.auditor_user_id === "string" &&
    typeof item.auditor_name === "string" &&
    typeof item.auditor_email === "string" &&
    typeof item.assigned_at === "string"
  );
}

function isProjectFeedback(value: unknown): value is ProjectFeedback {
  if (typeof value !== "object" || value === null) return false;
  const item = value as Record<string, unknown>;
  return (
    typeof item.id === "string" &&
    typeof item.project_id === "string" &&
    typeof item.author_user_id === "string" &&
    typeof item.author_name === "string" &&
    typeof item.author_email === "string" &&
    typeof item.body === "string" &&
    typeof item.created_at === "string"
  );
}

export async function selfAssignAuditor(
  projectId: string,
): Promise<ProjectAuditorAssignment> {
  const assignment = await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/auditors/me`,
    { method: "POST" },
  );
  if (!isAuditorAssignment(assignment)) {
    throw new Error("Invalid Auditor assignment response");
  }
  return assignment;
}

/** An Auditor in the system, as listed for the Work Item Auditor picker. */
export interface AuditorDirectoryEntry {
  id: string;
  name: string;
  email: string;
}

function isAuditorDirectoryEntry(value: unknown): value is AuditorDirectoryEntry {
  if (typeof value !== "object" || value === null) return false;
  const entry = value as Record<string, unknown>;
  return (
    typeof entry.id === "string" &&
    typeof entry.name === "string" &&
    typeof entry.email === "string"
  );
}

export async function listAuditorDirectory(): Promise<AuditorDirectoryEntry[]> {
  const items = await sameOriginRequest<unknown>("/api/v1/auditors");
  if (!Array.isArray(items) || !items.every(isAuditorDirectoryEntry)) {
    throw new Error("Invalid Auditor directory response");
  }
  return items;
}

/** Assigns an Auditor to the Project; assigning an already-assigned Auditor is a no-op. */
export async function assignProjectAuditor(
  projectId: string,
  auditorUserId: string,
): Promise<ProjectAuditorAssignment> {
  const assignment = await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/auditors`,
    { method: "POST", body: JSON.stringify({ auditor_user_id: auditorUserId }) },
  );
  if (!isAuditorAssignment(assignment)) {
    throw new Error("Invalid Auditor assignment response");
  }
  return assignment;
}

export async function listProjectAuditors(
  projectId: string,
): Promise<ProjectAuditorAssignment[]> {
  const items = await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/auditors`,
  );
  if (!Array.isArray(items) || !items.every(isAuditorAssignment)) {
    throw new Error("Invalid Auditor list response");
  }
  return items;
}

export async function listProjectFeedback(
  projectId: string,
): Promise<ProjectFeedback[]> {
  const items = await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/feedback`,
  );
  if (!Array.isArray(items) || !items.every(isProjectFeedback)) {
    throw new Error("Invalid Project feedback response");
  }
  return items;
}

export async function submitProjectFeedback(
  projectId: string,
  body: string,
): Promise<ProjectFeedback> {
  const feedback = await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/feedback`,
    { method: "POST", body: JSON.stringify({ body }) },
  );
  if (!isProjectFeedback(feedback)) {
    throw new Error("Invalid Project feedback create response");
  }
  return feedback;
}

export async function updateProject(
  projectId: string,
  input: UpdateProjectInput,
): Promise<ProjectDetails> {
  return sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}`,
    { method: "PATCH", body: JSON.stringify(input) },
  ).then(normalizeProject);
}

export async function deleteProject(projectId: string): Promise<void> {
  await sameOriginRequest<unknown>(
    `/api/v1/projects/${encodeURIComponent(projectId)}`,
    { method: "DELETE" },
  );
}

import { apiRequest, sameOriginRequest } from "@/api/client";
import type { Member } from "@/types/work-item";

const projectId = process.env.NEXT_PUBLIC_CHURA_PROJECT_ID;

export interface CreateProjectInput {
  name: string;
  description: string;
  template_id: string;
}

interface ApiProjectMember {
  user_id: string;
  name?: string;
  email?: string;
}

interface ApiProject {
  id: string;
  name: string;
  members?: ApiProjectMember[];
}

export interface ProjectSummary {
  id: string;
  name: string;
  description: string;
  template_id: string;
  mode: string;
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

export async function getCurrentProject(): Promise<CurrentProject> {
  if (!projectId) {
    throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
  }
  const project = await apiRequest<ApiProject>(`/projects/${projectId}`);
  return {
    id: project.id,
    name: project.name,
    members: (project.members ?? []).map((member) => {
      const name = member.name || member.email || member.user_id;
      return {
        id: member.user_id,
        name,
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

export async function getProject(projectId: string): Promise<ProjectSummary> {
  return apiRequest<ProjectSummary>(`/projects/${encodeURIComponent(projectId)}`);
}

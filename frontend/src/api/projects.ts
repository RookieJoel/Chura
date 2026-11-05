import { apiRequest } from "@/api/client";
import type { Member } from "@/types/work-item";

const projectId = process.env.NEXT_PUBLIC_CHURA_PROJECT_ID;

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
  const response = await fetch(
    `/api/v1/projects?page=${page}&page_size=${projectPageSize}`,
    { cache: "no-store" },
  );
  if (!response.ok) {
    let message = `Project discovery failed (${response.status})`;
    try {
      const body = (await response.json()) as { error?: unknown };
      if (typeof body.error === "string") message = body.error;
    } catch {
      // Keep the status-based message if the proxy did not return JSON.
    }
    throw new Error(message);
  }

  const result = (await response.json()) as ProjectListPage;
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

import { apiRequest } from "@/api/client";
import type { Sprint } from "@/types/work-item";

const projectId = process.env.NEXT_PUBLIC_CHURA_PROJECT_ID;

interface ApiSprint {
  id: string;
  project_id: string;
  name: string;
  team: string;
  start_date?: string;
  end_date?: string;
  status: Sprint["status"];
}

function toDateInput(value?: string): string {
  return value ? value.slice(0, 10) : "";
}

function toSprint(sprint: ApiSprint): Sprint {
  return {
    id: sprint.id,
    projectId: sprint.project_id,
    name: sprint.name,
    team: sprint.team,
    startDate: toDateInput(sprint.start_date),
    endDate: toDateInput(sprint.end_date),
    status: sprint.status,
  };
}

export async function listSprints(): Promise<Sprint[]> {
  if (!projectId) throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
  const sprints = await apiRequest<ApiSprint[]>(
    `/sprints?project_id=${encodeURIComponent(projectId)}`,
  );
  return sprints.map(toSprint);
}

export interface CreateSprintInput {
  name: string;
  team?: string;
  startDate?: string;
  endDate?: string;
  status?: Sprint["status"];
}

export async function createSprint(input: CreateSprintInput): Promise<Sprint> {
  if (!projectId) throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
  if (!input.name.trim()) throw new Error("Sprint name is required");
  if (!input.team?.trim()) throw new Error("Team is required");
  const created = await apiRequest<ApiSprint>("/sprints", {
    method: "POST",
    body: JSON.stringify({
      project_id: projectId,
      name: input.name,
      team: input.team ?? "",
      start_date: input.startDate || undefined,
      end_date: input.endDate || undefined,
      status: input.status ?? "planned",
    }),
  });
  return toSprint(created);
}

export async function getActiveSprint(): Promise<Sprint | null> {
  const available = await listSprints();
  return available.find((sprint) => sprint.status === "active") ?? null;
}

export async function updateSprintStatus(
  sprint: Sprint,
  status: Sprint["status"],
): Promise<Sprint> {
  return updateSprint(sprint, { status });
}

export interface UpdateSprintInput {
  name?: string;
  team?: string;
  startDate?: string;
  endDate?: string;
  status?: Sprint["status"];
}

export async function updateSprint(
  sprint: Sprint,
  input: UpdateSprintInput,
): Promise<Sprint> {
  const name = input.name ?? sprint.name;
  const team = input.team ?? sprint.team ?? "";
  if (!name.trim()) throw new Error("Sprint name is required");
  if (!team.trim()) throw new Error("Team is required");
  if (input.startDate && input.endDate && input.startDate > input.endDate) {
    throw new Error("End date must be on or after start date");
  }

  const updated = await apiRequest<ApiSprint>(`/sprints/${sprint.id}`, {
    method: "PUT",
    body: JSON.stringify({
      name,
      team,
      start_date: (input.startDate ?? sprint.startDate) || undefined,
      end_date: (input.endDate ?? sprint.endDate) || undefined,
      status: input.status ?? sprint.status,
    }),
  });
  return toSprint(updated);
}

export async function deleteSprint(id: string): Promise<void> {
  await apiRequest<void>(`/sprints/${id}`, { method: "DELETE" });
}

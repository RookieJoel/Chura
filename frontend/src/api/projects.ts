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

export interface CurrentProject {
  id: string;
  name: string;
  members: Member[];
}

export async function getCurrentProject(): Promise<CurrentProject> {
  if (!projectId) throw new Error("NEXT_PUBLIC_CHURA_PROJECT_ID is not configured");
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

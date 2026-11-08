import {
  getProject,
  type AuditorDirectoryEntry,
  type ProjectDetails,
} from "@/api/projects";
import type { Member } from "@/types/work-item";

/** Shapes Auditors like Members so pickers and avatars can show them. */
export function membersFromAuditors(auditors: AuditorDirectoryEntry[]): Member[] {
  return auditors.map((auditor) => {
    const name = auditor.name || auditor.email || auditor.id;
    return {
      id: auditor.id,
      name,
      initials: name
        .split(/\s+/)
        .map((part) => part[0] ?? "")
        .join("")
        .slice(0, 2)
        .toUpperCase(),
      color: "#7C3AED",
    };
  });
}

/** Maps already-loaded Project details to the Member shape used by Work Item UIs. */
export function membersOf(project: Pick<ProjectDetails, "members"> | null | undefined): Member[] {
  return (project?.members ?? []).map((member) => ({
    id: member.id,
    name: member.name,
    initials: member.initials,
    color: member.color,
  }));
}

/**
 * Fetches a Project's members. Prefer `membersOf(useProjects().selectedProject)`
 * when the Project is the selected one: the provider has already loaded it.
 */
export async function listMembers(projectId: string): Promise<Member[]> {
  const id = projectId.trim();
  if (!id) throw new Error("A Project must be selected");
  return membersOf(await getProject(id));
}

export async function getMember(
  projectId: string,
  id: string,
): Promise<Member | undefined> {
  return (await listMembers(projectId)).find((member) => member.id === id);
}

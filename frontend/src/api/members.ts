import { getCurrentProject } from "@/api/projects";
import type { Member } from "@/types/work-item";

export async function listMembers(): Promise<Member[]> {
  return (await getCurrentProject()).members;
}

export async function getMember(id: string): Promise<Member | undefined> {
  return (await listMembers()).find((member) => member.id === id);
}

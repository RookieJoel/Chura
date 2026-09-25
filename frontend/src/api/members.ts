import { members } from "@/lib/mock-data";
import type { Member } from "@/types/work-item";

// TODO: swap fixtures for real calls to the backend/ Fiber API once it exists.

export async function listMembers(): Promise<Member[]> {
  return members;
}

export async function getMember(id: string): Promise<Member | undefined> {
  return members.find((member) => member.id === id);
}

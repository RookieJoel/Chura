import { sprints } from "@/lib/mock-data";
import type { Sprint } from "@/types/work-item";

// TODO: swap fixtures for real calls to the backend/ Fiber API once it exists.

export async function getActiveSprint(): Promise<Sprint | null> {
  return sprints.find((sprint) => sprint.status === "active") ?? null;
}

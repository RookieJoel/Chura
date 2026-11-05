import { sameOriginRequest } from "@/api/client";

export interface ProjectTemplate {
  id: string;
  name: string;
  mode: string;
  creator_role: string;
}

function isProjectTemplate(value: unknown): value is ProjectTemplate {
  if (typeof value !== "object" || value === null) return false;
  const template = value as Record<string, unknown>;
  return (
    typeof template.id === "string" &&
    typeof template.name === "string" &&
    typeof template.mode === "string" &&
    typeof template.creator_role === "string"
  );
}

export async function listTemplates(): Promise<ProjectTemplate[]> {
  const templates = await sameOriginRequest<unknown>("/api/v1/templates");
  if (!Array.isArray(templates) || !templates.every(isProjectTemplate)) {
    throw new Error("Invalid template list response");
  }
  return templates;
}

export async function getTemplate(templateId: string): Promise<ProjectTemplate> {
  const template = await sameOriginRequest<unknown>(
    `/api/v1/templates/${encodeURIComponent(templateId)}`,
  );
  if (!isProjectTemplate(template)) {
    throw new Error("Invalid template response");
  }
  return template;
}

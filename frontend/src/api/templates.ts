import { sameOriginRequest } from "@/api/client";

export interface TemplateTypeRule {
  type: string;
  description_format?: string;
  require_acceptance_criteria: boolean;
}

export interface TemplateValidationRules {
  required_fields: string[];
  story_points?: { greater_than: number };
  type_rules: TemplateTypeRule[];
}

export interface ReflectionTopic {
  key: string;
  question: string;
  hint?: string;
  required?: boolean;
  max_runes?: number;
}

export interface ReflectionTemplate {
  topics: ReflectionTopic[];
}

export interface MetricDefinition {
  key: string;
  label: string;
  unit: string;
}

export interface SprintReviewTemplate {
  title: string;
  metrics: MetricDefinition[];
  default_metrics: string[];
}

export interface ProjectTemplate {
  id: string;
  name: string;
  mode: string;
  work_item_types: string[];
  roles: string[];
  default_role: string;
  creator_role: string;
  capabilities?: string[];
  validation_rules?: TemplateValidationRules;
  reflection?: ReflectionTemplate;
  sprint_review?: SprintReviewTemplate;
}

function isTypeRule(value: unknown): value is TemplateTypeRule {
  if (typeof value !== "object" || value === null) return false;
  const rule = value as Record<string, unknown>;
  return (
    typeof rule.type === "string" &&
    typeof rule.require_acceptance_criteria === "boolean" &&
    (rule.description_format === undefined ||
      typeof rule.description_format === "string")
  );
}

function isValidationRules(value: unknown): value is TemplateValidationRules {
  if (typeof value !== "object" || value === null) return false;
  const rules = value as Record<string, unknown>;
  if (
    !Array.isArray(rules.required_fields) ||
    !rules.required_fields.every((field) => typeof field === "string") ||
    !Array.isArray(rules.type_rules) ||
    !rules.type_rules.every(isTypeRule)
  ) {
    return false;
  }
  if (rules.story_points !== undefined) {
    if (typeof rules.story_points !== "object" || rules.story_points === null) {
      return false;
    }
    const storyPoints = rules.story_points as Record<string, unknown>;
    if (typeof storyPoints.greater_than !== "number") return false;
  }
  return true;
}

function isReflectionTopic(value: unknown): value is ReflectionTopic {
  if (typeof value !== "object" || value === null) return false;
  const topic = value as Record<string, unknown>;
  return (
    typeof topic.key === "string" &&
    typeof topic.question === "string" &&
    (topic.hint === undefined || typeof topic.hint === "string") &&
    (topic.required === undefined || typeof topic.required === "boolean") &&
    (topic.max_runes === undefined || typeof topic.max_runes === "number")
  );
}

function isReflectionTemplate(value: unknown): value is ReflectionTemplate {
  if (typeof value !== "object" || value === null) return false;
  const reflection = value as Record<string, unknown>;
  return (
    Array.isArray(reflection.topics) &&
    reflection.topics.every(isReflectionTopic)
  );
}

function isMetricDefinition(value: unknown): value is MetricDefinition {
  if (typeof value !== "object" || value === null) return false;
  const metric = value as Record<string, unknown>;
  return (
    typeof metric.key === "string" &&
    typeof metric.label === "string" &&
    typeof metric.unit === "string"
  );
}

function isSprintReviewTemplate(value: unknown): value is SprintReviewTemplate {
  if (typeof value !== "object" || value === null) return false;
  const review = value as Record<string, unknown>;
  return (
    typeof review.title === "string" &&
    Array.isArray(review.metrics) &&
    review.metrics.every(isMetricDefinition) &&
    Array.isArray(review.default_metrics) &&
    review.default_metrics.every((key) => typeof key === "string")
  );
}

function isProjectTemplate(value: unknown): value is ProjectTemplate {
  if (typeof value !== "object" || value === null) return false;
  const template = value as Record<string, unknown>;
  if (
    typeof template.id !== "string" ||
    typeof template.name !== "string" ||
    typeof template.mode !== "string" ||
    !Array.isArray(template.work_item_types) ||
    !template.work_item_types.every((type) => typeof type === "string") ||
    !Array.isArray(template.roles) ||
    !template.roles.every((role) => typeof role === "string") ||
    typeof template.default_role !== "string" ||
    typeof template.creator_role !== "string"
  ) {
    return false;
  }
  if (
    template.capabilities !== undefined &&
    (!Array.isArray(template.capabilities) ||
      !template.capabilities.every((capability) => typeof capability === "string"))
  ) {
    return false;
  }
  if (
    template.validation_rules !== undefined &&
    !isValidationRules(template.validation_rules)
  ) {
    return false;
  }
  if (
    template.reflection !== undefined &&
    !isReflectionTemplate(template.reflection)
  ) {
    return false;
  }
  if (
    template.sprint_review !== undefined &&
    !isSprintReviewTemplate(template.sprint_review)
  ) {
    return false;
  }
  return true;
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

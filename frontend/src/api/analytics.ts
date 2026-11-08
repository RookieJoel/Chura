import { apiRequest } from "@/api/client";
import type {
  MetricValue,
  ReflectionTopicSummary,
  ReviewExportFormat,
  SprintMetrics,
  SprintProgress,
  SprintReflection,
  SprintReviewSummary,
  SprintWorkSummary,
  WorkBucket,
  WorkCategory,
} from "@/types/analytics";
import type { Priority, Sprint, WorkItemStatus, WorkItemType } from "@/types/work-item";

interface ApiWorkBucket {
  items?: number;
  points?: number;
}

interface ApiSprintWorkSummary {
  planned?: ApiWorkBucket;
  completed?: ApiWorkBucket;
  carry_over?: ApiWorkBucket;
  blocked?: ApiWorkBucket;
}

interface ApiSprint {
  id: string;
  project_id?: string;
  name: string;
  team?: string;
  start_date?: string;
  end_date?: string;
  status: Sprint["status"];
}

interface ApiSprintProgress {
  total_days?: number;
  elapsed_days?: number;
  remaining_days?: number;
  time_elapsed_rate?: number;
}

interface ApiSprintMetricWorkItem {
  id: string;
  title: string;
  type: WorkItemType;
  status: WorkItemStatus;
  priority: Priority;
  assignee_id?: string;
  story_points?: number;
  category: WorkCategory;
}

interface ApiSprintMetrics {
  sprint: ApiSprint;
  work_summary?: ApiSprintWorkSummary;
  completion_rate?: number;
  points_completion_rate?: number;
  status_breakdown?: Record<string, number>;
  type_breakdown?: Record<string, number>;
  progress?: ApiSprintProgress | null;
  work_items?: ApiSprintMetricWorkItem[];
  calculated_at?: string;
}

function requireSprintId(sprintId: string | null | undefined): string {
  const id = sprintId?.trim();
  if (!id) throw new Error("A Sprint is required");
  return id;
}

function toBucket(bucket?: ApiWorkBucket): WorkBucket {
  return {
    items: bucket?.items ?? 0,
    points: bucket?.points ?? 0,
  };
}

function toWorkSummary(summary?: ApiSprintWorkSummary): SprintWorkSummary {
  return {
    planned: toBucket(summary?.planned),
    completed: toBucket(summary?.completed),
    carryOver: toBucket(summary?.carry_over),
    blocked: toBucket(summary?.blocked),
  };
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

function toProgress(progress?: ApiSprintProgress | null): SprintProgress | null {
  if (!progress) return null;
  return {
    totalDays: progress.total_days ?? 0,
    elapsedDays: progress.elapsed_days ?? 0,
    remainingDays: progress.remaining_days ?? 0,
    timeElapsedRate: progress.time_elapsed_rate ?? 0,
  };
}

function toSprintMetrics(metrics: ApiSprintMetrics): SprintMetrics {
  return {
    sprint: toSprint(metrics.sprint),
    workSummary: toWorkSummary(metrics.work_summary),
    completionRate: metrics.completion_rate ?? 0,
    pointsCompletionRate: metrics.points_completion_rate ?? 0,
    statusBreakdown: (metrics.status_breakdown ?? {}) as SprintMetrics["statusBreakdown"],
    typeBreakdown: (metrics.type_breakdown ?? {}) as SprintMetrics["typeBreakdown"],
    progress: toProgress(metrics.progress),
    workItems: (metrics.work_items ?? []).map((item) => ({
      id: item.id,
      title: item.title,
      type: item.type,
      status: item.status,
      priority: item.priority,
      assigneeId: item.assignee_id,
      storyPoints: item.story_points ?? 0,
      category: item.category,
    })),
    calculatedAt: metrics.calculated_at ?? "",
  };
}

interface ApiSprintReflection {
  id?: string;
  sprint_id?: string;
  project_id?: string;
  author_id?: string;
  author_email?: string;
  template_id?: string;
  answers?: Record<string, string>;
  created_at?: string;
  updated_at?: string;
}

interface ApiMetricValue {
  key?: string;
  label?: string;
  unit?: string;
  value?: number;
}

interface ApiReflectionTopicResponse {
  author_id?: string;
  author_email?: string;
  answer?: string;
}

interface ApiReflectionTopicSummary {
  key?: string;
  question?: string;
  responses?: ApiReflectionTopicResponse[];
}

interface ApiSprintReviewSummary {
  title?: string;
  project?: { id?: string; name?: string };
  sprint?: ApiSprint;
  template_id?: string;
  quantitative?: {
    metrics?: ApiMetricValue[];
    work_summary?: ApiSprintWorkSummary;
    progress?: ApiSprintProgress | null;
  };
  qualitative?: {
    topics?: ApiReflectionTopicSummary[];
    respondents?: number;
    team_size?: number;
    response_rate?: number;
  };
  generated_at?: string;
}

function toReflection(reflection: ApiSprintReflection): SprintReflection {
  return {
    id: reflection.id ?? "",
    sprintId: reflection.sprint_id ?? "",
    projectId: reflection.project_id ?? "",
    authorId: reflection.author_id ?? "",
    authorEmail: reflection.author_email ?? "",
    templateId: reflection.template_id ?? "",
    answers: reflection.answers ?? {},
    createdAt: reflection.created_at ?? "",
    updatedAt: reflection.updated_at ?? "",
  };
}

function toMetricValue(metric: ApiMetricValue): MetricValue {
  return {
    key: metric.key ?? "",
    label: metric.label ?? metric.key ?? "Metric",
    unit: metric.unit ?? "items",
    value: metric.value ?? 0,
  };
}

function toTopicSummary(topic: ApiReflectionTopicSummary): ReflectionTopicSummary {
  return {
    key: topic.key ?? "",
    question: topic.question ?? topic.key ?? "Topic",
    responses: (topic.responses ?? []).map((response) => ({
      authorId: response.author_id ?? "",
      authorEmail: response.author_email ?? "",
      answer: response.answer ?? "",
    })),
  };
}

function toReviewSummary(summary: ApiSprintReviewSummary): SprintReviewSummary {
  if (!summary.sprint) {
    throw new Error("Review summary is missing sprint data");
  }
  return {
    title: summary.title ?? "Sprint Review Summary",
    project: {
      id: summary.project?.id ?? "",
      name: summary.project?.name ?? "",
    },
    sprint: toSprint(summary.sprint),
    templateId: summary.template_id ?? "",
    quantitative: {
      metrics: (summary.quantitative?.metrics ?? []).map(toMetricValue),
      workSummary: summary.quantitative?.work_summary
        ? toWorkSummary(summary.quantitative.work_summary)
        : undefined,
      progress: toProgress(summary.quantitative?.progress),
    },
    qualitative: {
      topics: (summary.qualitative?.topics ?? []).map(toTopicSummary),
      respondents: summary.qualitative?.respondents ?? 0,
      teamSize: summary.qualitative?.team_size ?? 0,
      responseRate: summary.qualitative?.response_rate ?? 0,
    },
    generatedAt: summary.generated_at ?? "",
  };
}

function metricsQuery(metrics?: string[]): string {
  const keys = (metrics ?? []).map((key) => key.trim()).filter(Boolean);
  if (keys.length === 0) return "";
  return `?metrics=${encodeURIComponent(keys.join(","))}`;
}

function parseFilename(disposition: string | null): string | null {
  if (!disposition) return null;
  const utfMatch = /filename\*=UTF-8''([^;]+)/i.exec(disposition);
  if (utfMatch?.[1]) {
    try {
      return decodeURIComponent(utfMatch[1].trim());
    } catch {
      return utfMatch[1].trim();
    }
  }
  const plainMatch = /filename="?([^";]+)"?/i.exec(disposition);
  return plainMatch?.[1]?.trim() || null;
}

function extensionForFormat(format: ReviewExportFormat): string {
  return format === "csv" ? "csv" : "md";
}

/** GET /sprints/:id/work-summary — planned / completed / carry-over / blocked buckets. */
export async function getSprintWorkSummary(
  sprintId: string,
): Promise<SprintWorkSummary> {
  const id = requireSprintId(sprintId);
  const summary = await apiRequest<ApiSprintWorkSummary>(
    `/sprints/${encodeURIComponent(id)}/work-summary`,
  );
  return toWorkSummary(summary);
}

/** GET /sprints/:id/metrics — work summary, rates, breakdowns, and calendar progress. */
export async function getSprintMetrics(sprintId: string): Promise<SprintMetrics> {
  const id = requireSprintId(sprintId);
  const metrics = await apiRequest<ApiSprintMetrics>(
    `/sprints/${encodeURIComponent(id)}/metrics`,
  );
  return toSprintMetrics(metrics);
}

/** GET /sprints/:id/reflections — team reflections, oldest first. */
export async function listReflections(
  sprintId: string,
): Promise<SprintReflection[]> {
  const id = requireSprintId(sprintId);
  const reflections = await apiRequest<ApiSprintReflection[]>(
    `/sprints/${encodeURIComponent(id)}/reflections`,
  );
  return reflections.map(toReflection);
}

/** PUT /sprints/:id/reflections/me — create or replace the caller's reflection. */
export async function submitReflection(
  sprintId: string,
  answers: Record<string, string>,
): Promise<SprintReflection> {
  const id = requireSprintId(sprintId);
  const reflection = await apiRequest<ApiSprintReflection>(
    `/sprints/${encodeURIComponent(id)}/reflections/me`,
    {
      method: "PUT",
      body: JSON.stringify({ answers }),
    },
  );
  return toReflection(reflection);
}

/** GET /sprints/:id/review-summary — quantitative metrics + qualitative topics. */
export async function getReviewSummary(
  sprintId: string,
  metrics?: string[],
): Promise<SprintReviewSummary> {
  const id = requireSprintId(sprintId);
  const summary = await apiRequest<ApiSprintReviewSummary>(
    `/sprints/${encodeURIComponent(id)}/review-summary${metricsQuery(metrics)}`,
  );
  return toReviewSummary(summary);
}

/**
 * GET /sprints/:id/review-summary/export — download CSV or Markdown.
 * Triggers a browser download; returns the resolved filename.
 */
export async function exportReviewSummary(
  sprintId: string,
  options?: { metrics?: string[]; format?: ReviewExportFormat },
): Promise<string> {
  const id = requireSprintId(sprintId);
  const format: ReviewExportFormat = options?.format ?? "csv";
  const params = new URLSearchParams();
  const keys = (options?.metrics ?? []).map((key) => key.trim()).filter(Boolean);
  if (keys.length > 0) params.set("metrics", keys.join(","));
  params.set("format", format);
  const path = `/api/backend/sprints/${encodeURIComponent(id)}/review-summary/export?${params.toString()}`;

  let response: Response;
  try {
    response = await fetch(path, { cache: "no-store" });
  } catch (error) {
    throw new Error(
      `Unable to reach the backend: ${
        error instanceof Error ? error.message : "network request failed"
      }`,
    );
  }

  if (!response.ok) {
    const body = await response.text();
    let message = `Backend request failed (${response.status})`;
    try {
      const data = body ? JSON.parse(body) : undefined;
      if (
        typeof data === "object" &&
        data !== null &&
        "error" in data &&
        typeof data.error === "string"
      ) {
        message = `${data.error} (${response.status})`;
      }
    } catch {
      // Keep the status message when the error body is not JSON.
    }
    throw new Error(message);
  }

  const blob = await response.blob();
  const filename =
    parseFilename(response.headers.get("content-disposition")) ??
    `sprint-${id}-review.${extensionForFormat(format)}`;

  const objectUrl = URL.createObjectURL(blob);
  try {
    const anchor = document.createElement("a");
    anchor.href = objectUrl;
    anchor.download = filename;
    anchor.rel = "noopener";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
  } finally {
    URL.revokeObjectURL(objectUrl);
  }

  return filename;
}

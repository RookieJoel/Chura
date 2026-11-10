import type { Sprint, WorkItemStatus, WorkItemType, Priority } from "@/types/work-item";

export type WorkCategory = "completed" | "carry_over" | "blocked";

export interface WorkBucket {
  items: number;
  points: number;
}

export interface SprintWorkSummary {
  planned: WorkBucket;
  completed: WorkBucket;
  carryOver: WorkBucket;
  blocked: WorkBucket;
}

export interface SprintProgress {
  totalDays: number;
  elapsedDays: number;
  remainingDays: number;
  timeElapsedRate: number;
}

export interface SprintMetricWorkItem {
  id: string;
  title: string;
  type: WorkItemType;
  status: WorkItemStatus;
  priority: Priority;
  assigneeId?: string;
  storyPoints: number;
  category: WorkCategory;
}

export interface SprintMetrics {
  sprint: Sprint;
  workSummary: SprintWorkSummary;
  completionRate: number;
  pointsCompletionRate: number;
  statusBreakdown: Partial<Record<WorkItemStatus, number>>;
  typeBreakdown: Partial<Record<WorkItemType, number>>;
  /** Null when the Sprint has no start or end date. */
  progress: SprintProgress | null;
  workItems: SprintMetricWorkItem[];
  calculatedAt: string;
}

export type ReflectionTopicKey = "went_well" | "stop" | "start" | (string & {});

export type MetricKey =
  | "planned_items"
  | "completed_items"
  | "carry_over_items"
  | "blocked_items"
  | "completion_rate"
  | "planned_points"
  | "completed_points"
  | "carry_over_points"
  | "blocked_points"
  | "points_completion_rate";

export type MetricUnit = "items" | "points" | "percent";

export interface MetricDefinition {
  key: MetricKey | string;
  label: string;
  unit: MetricUnit | string;
}

export interface MetricValue extends MetricDefinition {
  value: number;
}

export interface SprintReflection {
  id: string;
  sprintId: string;
  projectId: string;
  authorId: string;
  authorEmail: string;
  templateId: string;
  answers: Record<string, string>;
  createdAt: string;
  updatedAt: string;
}

export interface ReflectionTopicResponse {
  authorId: string;
  authorEmail: string;
  answer: string;
}

export interface ReflectionTopicSummary {
  key: string;
  question: string;
  responses: ReflectionTopicResponse[];
}

export interface SprintReviewSummary {
  title: string;
  project: { id: string; name: string };
  sprint: Sprint;
  templateId: string;
  quantitative: {
    metrics: MetricValue[];
    workSummary?: SprintWorkSummary;
    progress: SprintProgress | null;
  };
  qualitative: {
    topics: ReflectionTopicSummary[];
    respondents: number;
    teamSize: number;
    responseRate: number;
  };
  generatedAt: string;
}

export type ReviewExportFormat = "csv" | "markdown" | "md";

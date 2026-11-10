import type { ReflectionTopic } from "@/api/templates";

/** Matches backend StandardReflectionTemplate when the Project Template omits topics. */
export const DEFAULT_REFLECTION_TOPICS: ReflectionTopic[] = [
  {
    key: "went_well",
    question: "What went well?",
    hint: "Practices, tools or habits that helped the team this sprint.",
    required: true,
    max_runes: 2000,
  },
  {
    key: "stop",
    question: "What should we stop doing?",
    hint: "Things that slowed the team down or caused problems.",
    required: true,
    max_runes: 2000,
  },
  {
    key: "start",
    question: "What should we start doing?",
    hint: "New ideas to try in the next sprint.",
    required: true,
    max_runes: 2000,
  },
];

export function reflectionTopicsFromTemplate(
  topics?: ReflectionTopic[],
): ReflectionTopic[] {
  return topics && topics.length > 0 ? topics : DEFAULT_REFLECTION_TOPICS;
}

export function formatReflectionTime(value: string): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function formatMetricValue(value: number, unit: string): string {
  if (unit === "percent") {
    return `${Number.isInteger(value) ? value.toFixed(0) : value.toFixed(1)}%`;
  }
  return Number.isInteger(value) ? String(value) : value.toFixed(1);
}

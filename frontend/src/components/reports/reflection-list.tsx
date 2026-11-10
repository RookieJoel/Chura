"use client";

import type { ReflectionTopic } from "@/api/templates";
import type { SprintReflection } from "@/types/analytics";
import { formatReflectionTime } from "@/lib/reflection-utils";

export function ReflectionList({
  reflections,
  topics,
  loading,
  error,
}: {
  reflections: SprintReflection[];
  topics: ReflectionTopic[];
  loading?: boolean;
  error?: string | null;
}) {
  if (loading) {
    return (
      <p role="status" className="m-0 text-sm text-text-tertiary">
        Loading reflections…
      </p>
    );
  }

  if (error) {
    return (
      <p role="alert" className="m-0 text-sm text-red-600">
        {error}
      </p>
    );
  }

  if (reflections.length === 0) {
    return (
      <p className="m-0 text-sm text-text-tertiary">
        No reflections submitted for this Sprint yet.
      </p>
    );
  }

  return (
    <ul className="m-0 list-none space-y-4 p-0">
      {reflections.map((reflection) => (
        <li
          key={reflection.id || `${reflection.authorId}-${reflection.updatedAt}`}
          className="rounded-lg border border-border bg-background p-4"
        >
          <div className="mb-3 flex flex-wrap items-baseline justify-between gap-2">
            <p className="m-0 text-[13px] font-bold text-text">
              {reflection.authorEmail || reflection.authorId || "Team member"}
            </p>
            <p className="m-0 text-[11px] text-text-tertiary">
              {formatReflectionTime(reflection.updatedAt || reflection.createdAt)}
            </p>
          </div>
          <dl className="m-0 space-y-3">
            {topics.map((topic) => {
              const answer = reflection.answers[topic.key]?.trim();
              if (!answer) return null;
              return (
                <div key={topic.key}>
                  <dt className="text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
                    {topic.question}
                  </dt>
                  <dd className="m-0 mt-1 whitespace-pre-wrap text-sm text-text-secondary">
                    {answer}
                  </dd>
                </div>
              );
            })}
          </dl>
        </li>
      ))}
    </ul>
  );
}

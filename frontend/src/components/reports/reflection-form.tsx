"use client";

import { useEffect, useState, type FormEvent } from "react";
import type { ReflectionTopic } from "@/api/templates";
import type { SprintReflection } from "@/types/analytics";

export function ReflectionForm({
  topics,
  existing,
  submitting,
  error,
  onSubmit,
}: {
  topics: ReflectionTopic[];
  existing?: SprintReflection | null;
  submitting?: boolean;
  error?: string | null;
  onSubmit: (answers: Record<string, string>) => Promise<void> | void;
}) {
  const [answers, setAnswers] = useState<Record<string, string>>({});
  // Depend on content signatures so a new topics array identity does not wipe in-progress typing.
  const topicsSignature = topics.map((topic) => topic.key).join("|");
  const existingSignature = `${existing?.id ?? ""}:${existing?.updatedAt ?? ""}`;

  /* eslint-disable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps -- seed from stored reflection / topic keys only */
  useEffect(() => {
    const next: Record<string, string> = {};
    for (const topic of topics) {
      next[topic.key] = existing?.answers[topic.key] ?? "";
    }
    setAnswers(next);
  }, [topicsSignature, existingSignature]);
  /* eslint-enable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (submitting) return;
    await onSubmit(answers);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {topics.map((topic) => {
        const maxRunes = topic.max_runes ?? 2000;
        const value = answers[topic.key] ?? "";
        const fieldId = `reflection-${topic.key}`;
        return (
          <div key={topic.key} className="block">
            <label htmlFor={fieldId} className="text-[13px] font-bold text-text">
              {topic.question}
              {topic.required !== false ? (
                <span className="text-danger"> *</span>
              ) : null}
            </label>
            {topic.hint ? (
              <p className="m-0 mt-0.5 text-[12px] text-text-tertiary">
                {topic.hint}
              </p>
            ) : null}
            <textarea
              id={fieldId}
              name={topic.key}
              value={value}
              onChange={(event) =>
                setAnswers((current) => ({
                  ...current,
                  [topic.key]: event.target.value,
                }))
              }
              rows={3}
              maxLength={maxRunes}
              required={topic.required !== false}
              className="mt-1.5 w-full resize-y rounded-lg border border-border bg-background px-3 py-2 text-sm text-text outline-none focus:border-brand"
            />
            <span className="mt-1 block text-right text-[11px] text-text-tertiary">
              {value.length}/{maxRunes}
            </span>
          </div>
        );
      })}

      {error ? (
        <p role="alert" className="m-0 text-sm text-red-600">
          {error}
        </p>
      ) : null}

      <button
        type="submit"
        disabled={submitting}
        className="fl-cta px-3 py-2 text-xs disabled:opacity-60"
      >
        {submitting
          ? "Saving…"
          : existing
            ? "Update my reflection"
            : "Submit my reflection"}
      </button>
    </form>
  );
}

"use client";

import type { MetricDefinition } from "@/api/templates";
import type { SprintReviewSummary } from "@/types/analytics";
import { formatMetricValue, formatReflectionTime } from "@/lib/reflection-utils";

export function ReviewSummaryPanel({
  availableMetrics,
  selectedMetrics,
  onToggleMetric,
  summary,
  loading,
  error,
  exporting,
  exportError,
  onGenerate,
  onExport,
}: {
  availableMetrics: MetricDefinition[];
  selectedMetrics: string[];
  onToggleMetric: (key: string) => void;
  summary: SprintReviewSummary | null;
  loading?: boolean;
  error?: string | null;
  exporting?: boolean;
  exportError?: string | null;
  onGenerate: () => void;
  onExport: (format: "csv" | "md") => void;
}) {
  return (
    <section aria-labelledby="review-summary-heading" className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2
            id="review-summary-heading"
            className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary"
          >
            Review summary
          </h2>
          <p className="m-0 mt-1 text-sm text-text-secondary">
            Generate a Sprint Review preview, then export CSV or Markdown.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            onClick={onGenerate}
            disabled={loading}
            className="fl-cta px-3 py-2 text-xs disabled:opacity-60"
          >
            {loading ? "Generating…" : summary ? "Refresh summary" : "Generate summary"}
          </button>
          <button
            type="button"
            onClick={() => onExport("csv")}
            disabled={exporting || !summary}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-xs font-bold text-text hover:bg-background disabled:opacity-60"
          >
            {exporting ? "Exporting…" : "Export CSV"}
          </button>
          <button
            type="button"
            onClick={() => onExport("md")}
            disabled={exporting || !summary}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-xs font-bold text-text hover:bg-background disabled:opacity-60"
          >
            Export Markdown
          </button>
        </div>
      </div>

      {availableMetrics.length > 0 ? (
        <fieldset className="m-0 rounded-lg border border-border bg-background p-3">
          <legend className="px-1 text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
            Metrics (optional)
          </legend>
          <p className="m-0 mb-2 text-[12px] text-text-tertiary">
            Leave unchecked to use the Project Template defaults.
          </p>
          <div className="flex flex-wrap gap-x-4 gap-y-2">
            {availableMetrics.map((metric) => {
              const checked = selectedMetrics.includes(metric.key);
              return (
                <label
                  key={metric.key}
                  className="flex items-center gap-2 text-[13px] text-text-secondary"
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => onToggleMetric(metric.key)}
                  />
                  {metric.label}
                </label>
              );
            })}
          </div>
        </fieldset>
      ) : null}

      {error ? (
        <p role="alert" className="m-0 text-sm text-red-600">
          {error}
        </p>
      ) : null}
      {exportError ? (
        <p role="alert" className="m-0 text-sm text-red-600">
          {exportError}
        </p>
      ) : null}

      {loading && !summary ? (
        <p role="status" className="m-0 text-sm text-text-tertiary">
          Generating review summary…
        </p>
      ) : null}

      {summary ? (
        <div className="space-y-5 rounded-lg border border-border bg-background p-4">
          <div>
            <h3 className="m-0 text-base font-bold text-text">{summary.title}</h3>
            <p className="m-0 mt-1 text-[12px] text-text-tertiary">
              {summary.project.name} · {summary.sprint.name}
              {summary.generatedAt
                ? ` · Generated ${formatReflectionTime(summary.generatedAt)}`
                : ""}
            </p>
          </div>

          {summary.quantitative.metrics.length > 0 ? (
            <div>
              <h4 className="m-0 text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
                Quantitative
              </h4>
              <dl className="mt-2 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {summary.quantitative.metrics.map((metric) => (
                  <div key={metric.key}>
                    <dt className="text-[11px] font-semibold text-text-tertiary">
                      {metric.label}
                    </dt>
                    <dd className="m-0 mt-0.5 text-sm font-bold text-text">
                      {formatMetricValue(metric.value, metric.unit)}
                    </dd>
                  </div>
                ))}
              </dl>
            </div>
          ) : (
            <p className="m-0 text-sm text-text-tertiary">
              No quantitative metrics available for this preview.
            </p>
          )}

          <div>
            <h4 className="m-0 text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
              Reflections
            </h4>
            <p className="m-0 mt-1 text-[12px] text-text-secondary">
              {summary.qualitative.respondents} of {summary.qualitative.teamSize}{" "}
              members responded
              {summary.qualitative.teamSize > 0
                ? ` (${formatMetricValue(summary.qualitative.responseRate, "percent")})`
                : ""}
            </p>
            {summary.qualitative.topics.length === 0 ? (
              <p className="m-0 mt-2 text-sm text-text-tertiary">
                No qualitative responses yet.
              </p>
            ) : (
              <div className="mt-3 space-y-4">
                {summary.qualitative.topics.map((topic) => (
                  <div key={topic.key}>
                    <p className="m-0 text-[13px] font-bold text-text">
                      {topic.question}
                    </p>
                    {topic.responses.length === 0 ? (
                      <p className="m-0 mt-1 text-sm text-text-tertiary">
                        No answers for this topic.
                      </p>
                    ) : (
                      <ul className="mt-1 list-none space-y-2 p-0">
                        {topic.responses.map((response, index) => (
                          <li
                            key={`${topic.key}-${response.authorId}-${index}`}
                            className="text-sm text-text-secondary"
                          >
                            <span className="font-semibold text-text">
                              {response.authorEmail || response.authorId}:
                            </span>{" "}
                            <span className="whitespace-pre-wrap">
                              {response.answer}
                            </span>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      ) : null}
    </section>
  );
}

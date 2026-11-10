"use client";

import type { SprintMetrics, SprintWorkSummary, WorkBucket } from "@/types/analytics";

const BUCKETS: {
  key: keyof SprintWorkSummary;
  label: string;
}[] = [
  { key: "planned", label: "Planned" },
  { key: "completed", label: "Completed" },
  { key: "carryOver", label: "Carry-over" },
  { key: "blocked", label: "Blocked" },
];

function formatRate(value: number) {
  return `${Number.isInteger(value) ? value.toFixed(0) : value.toFixed(1)}%`;
}

function BucketStat({ label, bucket }: { label: string; bucket: WorkBucket }) {
  return (
    <div className="min-w-[4.5rem]">
      <p className="m-0 text-[10px] font-bold uppercase tracking-wider text-text-tertiary">
        {label}
      </p>
      <p className="m-0 mt-0.5 text-sm font-bold text-text">
        {bucket.items}
        <span className="ml-1 text-[11px] font-semibold text-text-tertiary">
          ({bucket.points} pts)
        </span>
      </p>
    </div>
  );
}

export function SprintMetricsSummary({
  metrics,
  loading,
  error,
}: {
  metrics: SprintMetrics | null;
  loading?: boolean;
  error?: string;
}) {
  if (loading && !metrics) {
    return (
      <div className="border-b border-border bg-surface px-8 py-3 text-sm text-text-tertiary">
        Loading sprint metrics...
      </div>
    );
  }

  if (error && !metrics) {
    return (
      <div className="border-b border-border bg-surface px-8 py-3 text-sm text-text-secondary">
        Sprint metrics unavailable. Showing available Project and Board
        information only.
        <span className="mt-1 block text-xs text-text-tertiary">{error}</span>
      </div>
    );
  }

  if (!metrics) return null;

  const { workSummary, completionRate, pointsCompletionRate, progress } = metrics;

  return (
    <div className="border-b border-border bg-surface px-8 py-3">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-wrap gap-5">
          {BUCKETS.map(({ key, label }) => (
            <BucketStat key={key} label={label} bucket={workSummary[key]} />
          ))}
        </div>
        <div className="flex flex-wrap items-end gap-5 text-right">
          <div>
            <p className="m-0 text-[10px] font-bold uppercase tracking-wider text-text-tertiary">
              Item completion
            </p>
            <p className="m-0 mt-0.5 text-sm font-bold text-text">
              {formatRate(completionRate)}
            </p>
          </div>
          <div>
            <p className="m-0 text-[10px] font-bold uppercase tracking-wider text-text-tertiary">
              Points completion
            </p>
            <p className="m-0 mt-0.5 text-sm font-bold text-text">
              {formatRate(pointsCompletionRate)}
            </p>
          </div>
        </div>
      </div>

      {progress ? (
        <div className="mt-3">
          <div className="mb-1 flex items-center justify-between text-[11px] font-semibold text-text-secondary">
            <span>
              Day {progress.elapsedDays} of {progress.totalDays}
            </span>
            <span>{formatRate(progress.timeElapsedRate)} elapsed</span>
          </div>
          <div
            className="h-1.5 overflow-hidden rounded-sm bg-border"
            role="progressbar"
            aria-valuenow={progress.timeElapsedRate}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label="Sprint time elapsed"
          >
            <div
              className="h-full rounded-sm bg-brand"
              style={{ width: `${Math.min(100, Math.max(0, progress.timeElapsedRate))}%` }}
            />
          </div>
        </div>
      ) : (
        <p className="m-0 mt-3 text-[11px] text-text-tertiary">
          Sprint calendar progress unavailable until start and end dates are set.
        </p>
      )}

      {error && (
        <p className="m-0 mt-2 text-xs text-red-600">{error}</p>
      )}
    </div>
  );
}

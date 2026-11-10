"use client";

import { useState } from "react";
import { updateSprintStatus } from "@/api/sprints";
import type { Sprint } from "@/types/work-item";

export function SprintActions({
  sprint,
  workItemCount,
  onUpdated,
}: {
  sprint: Sprint;
  /** Assigned Work Items; used to disable Start when the Sprint is empty. */
  workItemCount?: number;
  onUpdated?: (sprint: Sprint) => void;
}) {
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string>();

  const starting = sprint.status === "planned";
  const nextStatus = starting ? "active" : "completed";
  const actionLabel = starting ? "Start sprint" : "End sprint";
  const savingLabel = starting ? "Starting..." : "Ending...";
  const emptyStart = starting && workItemCount !== undefined && workItemCount === 0;

  const changeStatus = async () => {
    if (emptyStart) {
      setError("Add work items to this Sprint before starting it");
      return;
    }
    if (!starting) {
      const confirmed = window.confirm(
        `End "${sprint.name}"? Unfinished work items stay on this Sprint (no automatic carry-over to the Backlog).`,
      );
      if (!confirmed) return;
    }
    setSaving(true);
    setError(undefined);
    try {
      const updated = await updateSprintStatus(sprint, nextStatus);
      onUpdated?.(updated);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to update sprint");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      {error && <span className="text-xs text-red-600">{error}</span>}
      <button
        type="button"
        disabled={saving || emptyStart}
        title={
          emptyStart
            ? "Assign work items to this Sprint before starting it"
            : undefined
        }
        onClick={() => void changeStatus()}
        className="fl-cta px-4 py-2.5 text-[13px]"
      >
        {saving ? savingLabel : actionLabel}
      </button>
    </div>
  );
}

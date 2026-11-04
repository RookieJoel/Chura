"use client";

import { useState } from "react";
import { updateSprintStatus } from "@/api/sprints";
import type { Sprint } from "@/types/work-item";

export function SprintActions({
  sprint,
  onUpdated,
}: {
  sprint: Sprint;
  onUpdated?: (sprint: Sprint) => void;
}) {
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string>();

  const nextStatus = sprint.status === "planned" ? "active" : "completed";
  const actionLabel = sprint.status === "planned" ? "Start sprint" : "End sprint";
  const savingLabel = sprint.status === "planned" ? "Starting..." : "Ending...";

  const changeStatus = async () => {
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
        disabled={saving}
        onClick={() => void changeStatus()}
        className="rounded-lg bg-brand px-4 py-2.5 text-[13px] font-bold text-white disabled:opacity-50"
      >
        {saving ? savingLabel : actionLabel}
      </button>
    </div>
  );
}

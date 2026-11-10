"use client";

import { useEffect, useMemo, useState } from "react";
import { getSprintMetrics } from "@/api/analytics";
import { getActiveSprint } from "@/api/sprints";
import { listWorkItems, updateWorkItem } from "@/api/work-items";
import { membersOf } from "@/api/members";
import { SprintMetricsSummary } from "@/components/board/sprint-metrics-summary";
import { AppShell } from "@/components/layout/app-shell";
import { NoProjectSelected } from "@/components/layout/no-project-selected";
import { useProjects } from "@/components/layout/project-provider";
import { WorkItemCard } from "@/components/work-items/work-item-card";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import Link from "next/link";
import type { SprintMetrics } from "@/types/analytics";
import type { Member, WorkItem } from "@/types/work-item";
import { canMutateExecution } from "@/lib/roles";
import {
  STATUS_LABELS,
  STATUS_ORDER,
  groupByStatus,
  sumPoints,
} from "@/lib/work-item-utils";

function formatDateRange(startDate: string, endDate: string) {
  const format = (value: string) =>
    new Date(value).toLocaleDateString("en-US", { month: "short", day: "numeric" });
  return `${format(startDate)} – ${format(endDate)}`;
}

function BoardView({
  selectedProjectId,
  projectsLoadState,
  canMutate,
  members,
}: {
  selectedProjectId: string | null;
  projectsLoadState: "loading" | "error" | "ready";
  canMutate: boolean;
  members: Member[];
}) {
  const [sprint, setSprint] = useState<Awaited<ReturnType<typeof getActiveSprint>>>(null);
  const [items, setItems] = useState<Awaited<ReturnType<typeof listWorkItems>>>([]);
  const [metrics, setMetrics] = useState<SprintMetrics | null>(null);
  const [metricsLoading, setMetricsLoading] = useState(false);
  const [metricsError, setMetricsError] = useState<string>();
  const [loading, setLoading] = useState(Boolean(selectedProjectId));
  const [error, setError] = useState<string>();
  const [savingItemId, setSavingItemId] = useState<string>();

  const refreshMetrics = async (sprintId: string) => {
    setMetricsLoading(true);
    setMetricsError(undefined);
    try {
      const nextMetrics = await getSprintMetrics(sprintId);
      setMetrics(nextMetrics);
    } catch (cause) {
      setMetrics(null);
      setMetricsError(
        cause instanceof Error ? cause.message : "Unable to load sprint metrics",
      );
    } finally {
      setMetricsLoading(false);
    }
  };

  useEffect(() => {
    let active = true;

    if (!selectedProjectId) {
      return () => {
        active = false;
      };
    }

    getActiveSprint(selectedProjectId)
      .then(async (nextSprint) => {
        const nextItems = nextSprint
          ? await listWorkItems({
              sprintId: nextSprint.id,
              projectId: selectedProjectId,
            })
          : [];
        if (!active) return;
        setSprint(nextSprint);
        setItems(nextItems);
        setMetrics(null);
        setMetricsError(undefined);
        if (nextSprint) {
          setMetricsLoading(true);
          try {
            const nextMetrics = await getSprintMetrics(nextSprint.id);
            if (active) setMetrics(nextMetrics);
          } catch (cause) {
            if (active) {
              setMetrics(null);
              setMetricsError(
                cause instanceof Error
                  ? cause.message
                  : "Unable to load sprint metrics",
              );
            }
          } finally {
            if (active) setMetricsLoading(false);
          }
        }
      })
      .catch((cause) => {
        if (active) {
          setError(cause instanceof Error ? cause.message : "Unable to load board");
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [selectedProjectId]);

  const handleStatusChange = async (item: WorkItem, status: WorkItem["status"]) => {
    if (!canMutate) return;
    if (status === item.status) return;

    let blockedReason = item.blockedReason;
    if (status === "blocked") {
      const reason = window.prompt(
        "Why is this work item blocked?",
        item.blockedReason ?? "",
      );
      if (!reason?.trim()) {
        setError("A blocked reason is required");
        return;
      }
      blockedReason = reason.trim();
    } else {
      blockedReason = undefined;
    }

    setSavingItemId(item.id);
    setError(undefined);
    try {
      const updated = await updateWorkItem({
        ...item,
        status,
        id: item.id,
        projectId: item.projectId ?? selectedProjectId ?? undefined,
        acceptanceCriteria: item.acceptanceCriteria,
        blockedReason,
      });
      setItems((current) =>
        current.map((currentItem) =>
          currentItem.id === updated.id ? updated : currentItem,
        ),
      );
      if (sprint) {
        void refreshMetrics(sprint.id);
      }
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to update work item");
    } finally {
      setSavingItemId(undefined);
    }
  };

  const membersById = new Map(members.map((member) => [member.id, member]));
  const columns = groupByStatus(items);
  const totalPoints = sumPoints(items);

  return (
    <>
      <div className="flex items-center justify-between border-b border-border bg-surface px-8 py-5">
        <div className="flex items-center gap-3.5">
          <div>
            <p className="m-0 text-[11px] font-bold uppercase tracking-wider text-text-tertiary">
              Active sprint
            </p>
            <h1 className="m-0 mt-1 text-2xl font-bold tracking-tight">Board</h1>
          </div>
          {sprint && (
            <div className="flex items-center gap-1.5 bg-brand-tint px-3 py-1.5 text-[12.5px] font-semibold text-brand-dark">
              {sprint.name}
              <span className="font-medium opacity-75">
                &middot; {formatDateRange(sprint.startDate, sprint.endDate)}
              </span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-3.5">
          <span className="text-[12.5px] font-semibold text-text-secondary">
            {totalPoints} story points
          </span>
        </div>
      </div>

      {sprint && (
        <SprintMetricsSummary
          metrics={metrics}
          loading={metricsLoading}
          error={metricsError}
        />
      )}

      {members.length > 0 && (
        <div className="flex items-center gap-3 border-b border-border bg-surface px-7 py-3">
          <div className="flex items-center">
            {members.map((member, index) => (
              <WorkItemAvatar
                key={member.id}
                member={member}
                className={`border-2 border-surface ${index < members.length - 1 ? "-mr-1.5" : ""}`}
              />
            ))}
          </div>
        </div>
      )}

      {error && (
        <p className="m-0 border-b border-border px-7 py-3 text-sm text-red-600">
          {error}
        </p>
      )}
      {projectsLoadState === "loading" || (selectedProjectId && loading) ? (
        <div className="flex flex-1 items-center justify-center text-sm text-text-tertiary">
          Loading board...
        </div>
      ) : !selectedProjectId ? (
        <NoProjectSelected title="Select a Project to open the Board" />
      ) : sprint ? (
        <div className="flex flex-1 gap-4 overflow-x-auto bg-background p-8">
          {STATUS_ORDER.map((status) => {
            const statusItems = columns[status];
            return (
              <div
                key={status}
                className="flex min-w-[260px] flex-1 flex-col border border-border bg-surface-muted p-3.5"
              >
                <div className="flex items-center justify-between px-1 pb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold uppercase tracking-wide text-text-secondary">
                      {STATUS_LABELS[status]}
                    </span>
                    <span className="rounded-full border border-border bg-surface px-1.5 py-px text-[11px] font-bold text-text-tertiary">
                      {statusItems.length}
                    </span>
                  </div>
                  <span className="text-[11px] font-semibold text-text-tertiary">
                    {sumPoints(statusItems)} pts
                  </span>
                </div>
                <div className="flex flex-col gap-2.5">
                  {statusItems.map((item) => (
                    <WorkItemCard
                      key={item.id}
                      item={item}
                      assignee={
                        item.assigneeId
                          ? membersById.get(item.assigneeId)
                          : undefined
                      }
                      saving={savingItemId === item.id}
                      onStatusChange={
                        canMutate
                          ? (nextStatus) => void handleStatusChange(item, nextStatus)
                          : undefined
                      }
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="flex flex-1 flex-col items-center justify-center gap-4 px-8 text-sm text-text-tertiary">
          <p className="m-0 text-base font-semibold text-text">No active sprint</p>
          <p className="m-0 max-w-md text-center">
            Sprint metrics are unavailable until a Sprint is active. Review the
            Backlog and Sprint list for this Project in the meantime.
          </p>
          {members.length > 0 && (
            <div className="flex flex-wrap items-center justify-center gap-2">
              {members.map((member) => (
                <span
                  key={member.id}
                  className="rounded-full border border-border bg-surface px-2.5 py-1 text-[11px] font-semibold text-text-secondary"
                >
                  {member.name}
                </span>
              ))}
            </div>
          )}
          <div className="flex flex-wrap items-center justify-center gap-2">
            <Link
              href="/backlog"
              className="border border-border bg-surface px-3 py-2 text-xs font-semibold text-text-secondary hover:border-accent hover:text-accent"
            >
              Open Backlog
            </Link>
            <Link
              href="/sprints"
              className="fl-cta px-4 py-2 text-xs"
            >
              {canMutate ? "Manage sprints" : "View sprints"}
            </Link>
          </div>
        </div>
      )}
    </>
  );
}

export default function BoardPage() {
  const {
    selectedProjectId,
    selectedProject,
    loadState: projectsLoadState,
    currentUser,
  } = useProjects();
  const canMutate = canMutateExecution(currentUser?.roles);
  // Members come from the Project the provider already loaded.
  const members = useMemo(() => membersOf(selectedProject), [selectedProject]);

  return (
    <AppShell>
      <BoardView
        key={selectedProjectId ?? "none"}
        selectedProjectId={selectedProjectId}
        projectsLoadState={projectsLoadState}
        canMutate={canMutate}
        members={members}
      />
    </AppShell>
  );
}

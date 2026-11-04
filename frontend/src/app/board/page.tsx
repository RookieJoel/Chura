"use client";

import { useEffect, useState } from "react";
import { getActiveSprint } from "@/api/sprints";
import { listWorkItems, updateWorkItem } from "@/api/work-items";
import { listMembers } from "@/api/members";
import { AppShell } from "@/components/layout/app-shell";
import { WorkItemCard } from "@/components/work-items/work-item-card";
import { SearchIcon } from "@/components/icons";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import Link from "next/link";
import type { WorkItem } from "@/types/work-item";
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

export default function BoardPage() {
  const [sprint, setSprint] = useState<Awaited<ReturnType<typeof getActiveSprint>>>(null);
  const [members, setMembers] = useState<Awaited<ReturnType<typeof listMembers>>>([]);
  const [items, setItems] = useState<Awaited<ReturnType<typeof listWorkItems>>>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const [savingItemId, setSavingItemId] = useState<string>();

  useEffect(() => {
    let active = true;
    Promise.all([getActiveSprint(), listMembers()])
      .then(async ([nextSprint, nextMembers]) => {
        const nextItems = nextSprint
          ? await listWorkItems({ sprintId: nextSprint.id })
          : [];
        if (!active) return;
        setSprint(nextSprint);
        setMembers(nextMembers);
        setItems(nextItems);
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
  }, []);

  const handleStatusChange = async (item: WorkItem, status: WorkItem["status"]) => {
    if (status === item.status) return;
    setSavingItemId(item.id);
    setError(undefined);
    try {
      const updated = await updateWorkItem({ ...item, status, id: item.id });
      setItems((current) =>
        current.map((currentItem) =>
          currentItem.id === updated.id ? updated : currentItem,
        ),
      );
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
    <AppShell>
      <div className="flex items-center justify-between border-b border-border bg-surface px-8 py-5">
        <div className="flex items-center gap-3.5">
          <div>
            <p className="m-0 text-[11px] font-bold uppercase tracking-wider text-text-tertiary">Active sprint</p>
            <h1 className="m-0 mt-1 text-2xl font-bold tracking-tight">Board</h1>
          </div>
          {sprint && (
            <div className="flex items-center gap-1.5 rounded-md bg-brand-tint px-3 py-1.5 text-[12.5px] font-bold text-brand-dark">
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

      <div className="flex items-center gap-3 border-b border-border bg-surface px-7 py-3">
        <div className="flex w-[260px] items-center gap-2 rounded-md border border-border bg-background px-3 py-2 text-text-tertiary">
          <SearchIcon className="size-3.5" />
          <span className="text-[12.5px]">Filter work items</span>
        </div>
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

      {error && (
        <p className="m-0 border-b border-border px-7 py-3 text-sm text-red-600">
          {error}
        </p>
      )}
      {loading ? (
        <div className="flex flex-1 items-center justify-center text-sm text-text-tertiary">
          Loading board...
        </div>
      ) : sprint ? (
        <div className="flex flex-1 gap-4 overflow-x-auto bg-background p-8">
          {STATUS_ORDER.map((status) => {
            const statusItems = columns[status];
            return (
              <div
                key={status}
                className="flex min-w-[260px] flex-1 flex-col rounded-lg border border-border bg-[#ebecf0] p-3.5"
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
                      onStatusChange={(status) =>
                        void handleStatusChange(item, status)
                      }
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 text-sm text-text-tertiary">
          <p className="m-0">No active sprint.</p>
          <Link
            href="/sprints"
            className="rounded-lg bg-brand px-3 py-2 text-xs font-bold text-white hover:bg-brand-dark"
          >
            Manage sprints
          </Link>
        </div>
      )}
    </AppShell>
  );
}

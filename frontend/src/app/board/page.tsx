import type { Metadata } from "next";
import { getActiveSprint } from "@/api/sprints";
import { listWorkItems } from "@/api/work-items";
import { listMembers } from "@/api/members";
import { AppShell } from "@/components/layout/app-shell";
import { WorkItemCard } from "@/components/work-items/work-item-card";
import { SearchIcon } from "@/components/icons";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import {
  STATUS_LABELS,
  STATUS_ORDER,
  groupByStatus,
  sumPoints,
} from "@/lib/work-item-utils";

export const metadata: Metadata = { title: "Board · Chura" };

function formatDateRange(startDate: string, endDate: string) {
  const format = (value: string) =>
    new Date(value).toLocaleDateString("en-US", { month: "short", day: "numeric" });
  return `${format(startDate)} – ${format(endDate)}`;
}

export default async function BoardPage() {
  const [sprint, members] = await Promise.all([
    getActiveSprint(),
    listMembers(),
  ]);
  const items = sprint ? await listWorkItems({ sprintId: sprint.id }) : [];
  const membersById = new Map(members.map((member) => [member.id, member]));
  const columns = groupByStatus(items);
  const totalPoints = sumPoints(items);

  return (
    <AppShell>
      <div className="flex items-center justify-between border-b border-border bg-surface px-7 py-[18px]">
        <div className="flex items-center gap-3.5">
          <h1 className="m-0 text-xl font-extrabold tracking-tight">Board</h1>
          {sprint && (
            <div className="flex items-center gap-1.5 rounded-full bg-brand-tint px-3 py-1.5 text-[12.5px] font-bold text-brand-dark">
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
          <button
            type="button"
            className="rounded-lg bg-brand px-4 py-2.5 text-[13px] font-bold text-white"
          >
            Complete sprint
          </button>
        </div>
      </div>

      <div className="flex items-center gap-3 border-b border-border bg-surface px-7 py-3">
        <div className="flex w-[230px] items-center gap-2 rounded-lg border border-border bg-background px-3 py-1.5 text-text-tertiary">
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

      {sprint ? (
        <div className="flex flex-1 gap-4 overflow-x-auto p-7">
          {STATUS_ORDER.map((status) => {
            const statusItems = columns[status];
            return (
              <div
                key={status}
                className="flex min-w-[260px] flex-1 flex-col rounded-xl bg-[#EEF0F3] p-3.5"
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
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="flex flex-1 items-center justify-center text-sm text-text-tertiary">
          No active sprint.
        </div>
      )}
    </AppShell>
  );
}

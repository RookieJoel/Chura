import type { Metadata } from "next";
import type { ReactNode } from "react";
import { getActiveSprint } from "@/api/sprints";
import { listWorkItems } from "@/api/work-items";
import { listMembers } from "@/api/members";
import { AppShell } from "@/components/layout/app-shell";
import { WorkItemRow } from "@/components/work-items/work-item-row";
import { ChevronDownIcon, PlusIcon, SearchIcon } from "@/components/icons";
import { sumPoints } from "@/lib/work-item-utils";
import type { WorkItem } from "@/types/work-item";

export const metadata: Metadata = { title: "Backlog · Chura" };

function SectionCard({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-surface">
      {children}
    </div>
  );
}

function SectionHeader({
  title,
  meta,
  badge,
  action,
}: {
  title: string;
  meta: string;
  badge?: ReactNode;
  action: ReactNode;
}) {
  return (
    <div className="flex items-center justify-between border-b border-border px-3.5 pb-2.5 pt-4">
      <div className="flex items-center gap-2.5">
        <ChevronDownIcon className="size-3.5" />
        <span className="text-sm font-bold">{title}</span>
        <span className="text-xs font-semibold text-text-tertiary">{meta}</span>
        {badge}
      </div>
      {action}
    </div>
  );
}

export default async function BacklogPage() {
  const [sprint, members] = await Promise.all([
    getActiveSprint(),
    listMembers(),
  ]);
  const [sprintItems, backlogItems] = await Promise.all([
    sprint ? listWorkItems({ sprintId: sprint.id }) : Promise.resolve<WorkItem[]>([]),
    listWorkItems({ sprintId: null }),
  ]);
  const membersById = new Map(members.map((member) => [member.id, member]));

  const renderRow = (item: WorkItem) => (
    <WorkItemRow
      key={item.id}
      item={item}
      assignee={item.assigneeId ? membersById.get(item.assigneeId) : undefined}
    />
  );

  return (
    <AppShell>
      <div className="flex items-center justify-between border-b border-border bg-surface px-7 py-[18px]">
        <h1 className="m-0 text-xl font-extrabold tracking-tight">Backlog</h1>
        <button
          type="button"
          className="rounded-lg bg-brand px-4 py-2.5 text-[13px] font-bold text-white"
        >
          Create work item
        </button>
      </div>

      <div className="flex items-center gap-3 border-b border-border bg-surface px-7 py-3">
        <div className="flex w-[230px] items-center gap-2 rounded-lg border border-border bg-background px-3 py-1.5 text-text-tertiary">
          <SearchIcon className="size-3.5" />
          <span className="text-[12.5px]">Filter work items</span>
        </div>
        <div className="flex items-center gap-1.5 rounded-full border border-border bg-surface px-3 py-1.5 text-[12.5px] font-semibold text-text-secondary">
          Epic: All
          <ChevronDownIcon className="size-3.5" />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-7 py-3.5 pb-6">
        {sprint && (
          <SectionCard>
            <SectionHeader
              title={sprint.name}
              meta={`Sep 15 – Sep 29 · ${sprintItems.length} issues · ${sumPoints(sprintItems)} pts`}
              badge={
                <span className="rounded-full bg-story-tint px-2 py-0.5 text-[10.5px] font-bold text-story">
                  Active
                </span>
              }
              action={
                <button
                  type="button"
                  className="rounded-lg border border-border bg-surface px-3.5 py-1.5 text-[12.5px] font-bold text-text-secondary"
                >
                  Complete sprint
                </button>
              }
            />
            {sprintItems.map(renderRow)}
          </SectionCard>
        )}

        <div className="mt-4">
          <SectionCard>
            <SectionHeader
              title="Backlog"
              meta={`${backlogItems.length} issues · ${sumPoints(backlogItems)} pts`}
              action={
                <button
                  type="button"
                  className="rounded-lg bg-brand-tint px-3.5 py-1.5 text-[12.5px] font-bold text-brand-dark"
                >
                  Create sprint
                </button>
              }
            />
            {backlogItems.map(renderRow)}
          </SectionCard>
        </div>

        <div className="mt-3.5 flex items-center gap-2 rounded-[10px] border border-dashed border-border-strong px-3.5 py-2.5 text-text-tertiary">
          <PlusIcon className="size-3.5" />
          <span className="text-[13px] font-semibold">Create work item</span>
        </div>
      </div>
    </AppShell>
  );
}

"use client"

import type { ReactNode } from "react";
import { createWorkItem, listWorkItems } from "@/api/work-items";
import { AppShell } from "@/components/layout/app-shell";
import { WorkItemInsertRow, WorkItemRow } from "@/components/work-items/work-item-row";
import { ChevronDownIcon, PlusIcon } from "@/components/icons";
import * as React from "react";

import type { WorkItem } from "@/types/work-item";

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

export default function BacklogPage() {
  type ItemState = {
    id: string;
    status: 'pending' | 'resolved' | 'error';
    item: WorkItem;
    error?: string;
  };

  const [additionalWorkItems, setAdditionalWorkItems] = React.useState<Map<string,ItemState>>(new Map());
  const [workItems, setWorkItems] = React.useState<WorkItem[]>([]);
  const [listError, setListError] = React.useState<string>();

  React.useEffect(() => {
    let active = true;
    listWorkItems({ sprintId: null })
      .then((items) => {
        if (active) setWorkItems(items);
      })
      .catch((error) => {
        if (active) {
          setListError(
            error instanceof Error ? error.message : "Unable to load work items",
          );
        }
      });

    return () => {
      active = false;
    };
  }, []);

  const handleAddWorkItem = async (title: string) => {
    const trimmedTitle = title.trim();
    if (!trimmedTitle) return;

    const tempId = `temp-${Date.now()}`;
    const draftItem: WorkItem = {
      id: tempId,
      type: "task",
      title: trimmedTitle,
      status: "todo",
      blocked: false,
      priority: "medium",
      storyPoints: 0,
      epic: "",
      sprintId: null,
      labels: [],
      acceptanceCriteria: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    const newItem: ItemState = { id: tempId, status: 'pending', item: draftItem };
    setAdditionalWorkItems(prev => new Map(prev).set(tempId, newItem));

    try {
      const item = await createWorkItem({ title: trimmedTitle });
      setAdditionalWorkItems(prev => {
        const next = new Map(prev);
        next.set(tempId, { id: tempId, status: "resolved", item });
        return next;
      });
    } catch (error) {
      setAdditionalWorkItems(prev => {
        const next = new Map(prev);
        next.set(tempId, {
          id: tempId,
          status: "error",
          item: draftItem,
          error: error instanceof Error ? error.message : "Unable to create work item",
        });
        return next;
      });
    }
  };

  return (
    <AppShell>
      <div className="flex items-center justify-between border-b border-border bg-surface px-7 py-[18px]">
        <h1 className="m-0 text-xl font-extrabold tracking-tight">Backlog</h1>
      </div>

      <div className="flex-1 overflow-y-auto px-7 py-3.5 pb-6">
        <SectionCard>
          <SectionHeader
            title={"Project Backlog"}
            meta={`Unassigned to sprint`}
            action={
              <></>
            }
          />
          {listError && (
            <p className="m-0 border-b border-border px-3.5 py-3 text-[13px] text-red-600">
              {listError}
            </p>
          )}
          {workItems.map((item) => (
            <WorkItemRow key={item.id} item={item} />
          ))}
          {Array.from(additionalWorkItems.values()).map((item) => (
            <WorkItemRow
              key={item.id}
              item={item.item}
              assignee={undefined}
              state={item.status}
              error={item.error}
            />
          ))}
          <WorkItemInsertRow onSubmit={handleAddWorkItem} />
        </SectionCard>

        <div className="mt-3.5 flex items-center gap-2 rounded-[10px] border border-dashed border-border-strong px-3.5 py-2.5 text-text-tertiary">
          <PlusIcon className="size-3.5" />
          <span className="text-[13px] font-semibold">Create Sprint</span>
        </div>
      </div>
    </AppShell>
  );
}

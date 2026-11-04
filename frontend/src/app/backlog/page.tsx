"use client"

import type { ReactNode } from "react";
import { createWorkItem, listWorkItems, updateWorkItem } from "@/api/work-items";
import { listMembers } from "@/api/members";
import { listSprints } from "@/api/sprints";
import { AppShell } from "@/components/layout/app-shell";
import { Modal } from "@/components/modal";
import { WorkItemRow } from "@/components/work-items/work-item-row";
import { ChevronDownIcon } from "@/components/icons";
import * as React from "react";

import type { WorkItem } from "@/types/work-item";

function SectionCard({ children }: { children: ReactNode }) {
  return <div className="overflow-hidden rounded-lg border border-border bg-surface shadow-[0_1px_2px_rgba(9,30,66,0.08)]">{children}</div>;
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
    <div className="flex items-center justify-between border-b border-border bg-surface px-4 py-3.5">
      <div className="flex items-center gap-2.5">
        <ChevronDownIcon className="size-3.5" />
        <span className="text-sm font-bold text-text">{title}</span>
        <span className="text-xs text-text-tertiary">{meta}</span>
        {badge}
      </div>
      {action}
    </div>
  );
}

function CreateWorkItemForm({
  sprints,
  initialSprintId,
  onCreated,
  onCancel,
  onError,
}: {
  sprints: Awaited<ReturnType<typeof listSprints>>;
  initialSprintId: string;
  onCreated: (item: WorkItem) => void;
  onCancel: () => void;
  onError: (message: string) => void;
}) {
  const [title, setTitle] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [type, setType] = React.useState<WorkItem["type"]>("task");
  const [priority, setPriority] = React.useState<WorkItem["priority"]>("medium");
  const [storyPoints, setStoryPoints] = React.useState(0);
  const [sprintId, setSprintId] = React.useState(initialSprintId);
  const [creating, setCreating] = React.useState(false);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      onError("Work Item title is required");
      return;
    }

    setCreating(true);
    onError("");
    try {
      const item = await createWorkItem({
        title: trimmedTitle,
        description: description.trim(),
        type,
        priority,
        storyPoints,
        sprintId: sprintId || null,
      });
      onCreated(item);
    } catch (error) {
      onError(error instanceof Error ? error.message : "Unable to create work item");
    } finally {
      setCreating(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="bg-background px-5 py-5">
      <div className="grid gap-3 sm:grid-cols-2">
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary sm:col-span-2">
          Title *
          <input
            required
            autoFocus
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            placeholder="What needs to be done?"
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary sm:col-span-2">
          Description
          <textarea
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            placeholder="Add context for the team"
            rows={2}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
          Type
          <select
            value={type}
            onChange={(event) => setType(event.target.value as WorkItem["type"])}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          >
            <option value="task">Task</option>
            <option value="story">Story</option>
            <option value="bug">Bug</option>
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
          Priority
          <select
            value={priority}
            onChange={(event) => setPriority(event.target.value as WorkItem["priority"])}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          >
            <option value="highest">Highest</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
          Story points
          <input
            type="number"
            min={0}
            value={storyPoints}
            onChange={(event) => setStoryPoints(Number(event.target.value))}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
          Sprint
          <select
            value={sprintId}
            onChange={(event) => setSprintId(event.target.value)}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          >
            <option value="">Backlog</option>
            {sprints.map((sprint) => (
              <option key={sprint.id} value={sprint.id}>{sprint.name}</option>
            ))}
          </select>
        </label>
      </div>
      <div className="mt-3 flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-lg border border-border px-4 py-2 text-xs font-bold text-text-secondary"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={creating}
          className="rounded-lg bg-brand px-4 py-2 text-xs font-bold text-white disabled:opacity-50"
        >
          {creating ? "Creating..." : "Create work item"}
        </button>
      </div>
    </form>
  );
}

export default function BacklogPage() {
  const [workItems, setWorkItems] = React.useState<WorkItem[]>([]);
  const [members, setMembers] = React.useState<Awaited<ReturnType<typeof listMembers>>>([]);
  const [sprints, setSprints] = React.useState<Awaited<ReturnType<typeof listSprints>>>([]);
  const [listError, setListError] = React.useState<string>();
  const [openCreateFor, setOpenCreateFor] = React.useState<string | null>(null);

  React.useEffect(() => {
    let active = true;
    void listSprints()
      .then((availableSprints) => {
        if (active) setSprints(availableSprints);
      })
      .catch((error) => {
        if (active) setListError(
          error instanceof Error ? error.message : "Unable to load sprints",
        );
      });
    void listWorkItems()
      .then((items) => {
        if (active) setWorkItems(items);
      })
      .catch((error) => {
        if (active) setListError(
          error instanceof Error ? error.message : "Unable to load work items",
        );
      });
    void listMembers()
      .then((projectMembers) => {
        if (active) setMembers(projectMembers);
      })
      .catch((error) => {
        if (active) setListError(
          error instanceof Error ? error.message : "Unable to load members",
        );
      });

    return () => {
      active = false;
    };
  }, []);

  const membersById = new Map(members.map((member) => [member.id, member]));

  const handleSprintChange = async (item: WorkItem, nextSprintId: string | null) => {
    setListError(undefined);
    try {
      await updateWorkItem({
        ...item,
        id: item.id,
        sprintId: nextSprintId,
      });
      setWorkItems((current) =>
        current.map((currentItem) =>
          currentItem.id === item.id
            ? { ...currentItem, sprintId: nextSprintId }
            : currentItem,
        ),
      );
    } catch (error) {
      setListError(error instanceof Error ? error.message : "Unable to attach work item to sprint");
    }
  };

  return (
    <AppShell>
      <div className="flex items-center justify-between border-b border-border bg-surface px-8 py-5">
        <div>
          <p className="m-0 text-[11px] font-bold uppercase tracking-wider text-text-tertiary">Chura Demo App</p>
          <h1 className="m-0 mt-1 text-2xl font-bold tracking-tight">Backlog</h1>
        </div>
        <button type="button" className="rounded-md border border-border px-3 py-2 text-xs font-bold text-text-secondary hover:bg-background">
          View settings
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-8 py-5 pb-8">
        <div className="flex flex-col gap-5">
          {[
            { id: "", name: "Project Backlog", meta: "Unassigned to sprint" },
            ...sprints.map((sprint) => ({
              id: sprint.id,
              name: sprint.name,
              meta: `${sprint.team ?? "No team"} · ${sprint.status}`,
            })),
          ].map((section) => {
            const sectionItems = workItems.filter(
              (item) => (item.sprintId ?? "") === section.id,
            );
            const isCreating = openCreateFor === section.id;
            return (
              <SectionCard key={section.id || "backlog"}>
                <SectionHeader
                  title={section.name}
                  meta={`${section.meta} · ${sectionItems.length} item${sectionItems.length === 1 ? "" : "s"}`}
                  action={
                    <button
                      type="button"
                      onClick={() => {
                        setOpenCreateFor(isCreating ? null : section.id);
                        setListError(undefined);
                      }}
                      className="rounded-lg bg-brand px-3 py-1.5 text-xs font-bold text-white"
                    >
                      {isCreating ? "Cancel" : "Add work item"}
                    </button>
                  }
                />
                {sectionItems.length === 0 && !isCreating ? (
                  <p className="m-0 px-4 py-5 text-sm text-text-tertiary">
                    No work items in this section.
                  </p>
                ) : (
                  sectionItems.map((item) => (
                    <WorkItemRow
                      key={item.id}
                      item={item}
                      assignee={item.assigneeId ? membersById.get(item.assigneeId) : undefined}
                      sprints={sprints}
                      onSprintChange={(nextSprintId) =>
                        void handleSprintChange(item, nextSprintId)
                      }
                    />
                  ))
                )}
              </SectionCard>
            );
          })}
        </div>
      </div>
      {openCreateFor !== null && (
        <Modal
          title="Create work item"
          onClose={() => setOpenCreateFor(null)}
        >
          <CreateWorkItemForm
            sprints={sprints}
            initialSprintId={openCreateFor}
            onCancel={() => setOpenCreateFor(null)}
            onError={(message) => setListError(message || undefined)}
            onCreated={(item) => {
              setWorkItems((current) => [item, ...current]);
              setOpenCreateFor(null);
              setListError(undefined);
            }}
          />
          {listError && (
            <p className="m-0 border-t border-border px-5 py-3 text-[13px] text-red-600">
              {listError}
            </p>
          )}
        </Modal>
      )}
    </AppShell>
  );
}

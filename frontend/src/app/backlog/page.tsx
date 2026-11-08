"use client"

import type { ReactNode } from "react";
import {
  createWorkItem,
  listWorkItems,
  updateWorkItem,
  validateWorkItemAttributes,
} from "@/api/work-items";
import { membersOf } from "@/api/members";
import {
  assignProjectAuditor,
  listAuditorDirectory,
  type AuditorDirectoryEntry,
} from "@/api/projects";
import { listSprints } from "@/api/sprints";
import type { ProjectTemplate } from "@/api/templates";
import { AppShell } from "@/components/layout/app-shell";
import { NoProjectSelected } from "@/components/layout/no-project-selected";
import { useProjects } from "@/components/layout/project-provider";
import { Modal } from "@/components/modal";
import { WorkItemRow } from "@/components/work-items/work-item-row";
import { ChevronDownIcon } from "@/components/icons";
import { canMutateExecution } from "@/lib/roles";
import {
  acceptanceCriteriaTexts,
  formatValidationViolations,
  templateRequiresAcceptanceCriteria,
  templateUsesUserStoryDescription,
  toApiWorkItemType,
  workItemTypesFromTemplate,
} from "@/lib/work-item-utils";
import * as React from "react";

import type { WorkItem, WorkItemType } from "@/types/work-item";

function SectionCard({ children }: { children: ReactNode }) {
  return <div className="overflow-hidden rounded-sm border border-border bg-surface shadow-[0_1px_2px_rgba(26,21,48,0.08)]">{children}</div>;
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

function typeLabel(type: WorkItemType): string {
  if (type === "story") return "User story";
  if (type === "bug") return "Bug";
  return "Task";
}

function CreateWorkItemForm({
  projectId,
  template,
  templateReady,
  members,
  auditors,
  sprints,
  initialSprintId,
  onCreated,
  onCancel,
  onError,
}: {
  projectId: string;
  template: ProjectTemplate | null;
  templateReady: boolean;
  members: ReturnType<typeof membersOf>;
  auditors: AuditorDirectoryEntry[];
  sprints: Awaited<ReturnType<typeof listSprints>>;
  initialSprintId: string;
  onCreated: (item: WorkItem) => void;
  onCancel: () => void;
  onError: (message: string) => void;
}) {
  const allowedTypes = workItemTypesFromTemplate(template);
  const [title, setTitle] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [type, setType] = React.useState<WorkItemType>(allowedTypes[0] ?? "task");
  const [priority, setPriority] = React.useState<WorkItem["priority"]>("medium");
  const [storyPoints, setStoryPoints] = React.useState(0);
  const [sprintId, setSprintId] = React.useState(initialSprintId);
  const [assigneeId, setAssigneeId] = React.useState("");
  // Stored as the Work Item's reporter_id: the Auditor following this item.
  const [auditorId, setAuditorId] = React.useState("");
  const [acceptanceCriteria, setAcceptanceCriteria] = React.useState<string[]>([""]);
  const [creating, setCreating] = React.useState(false);

  const effectiveType = allowedTypes.includes(type)
    ? type
    : (allowedTypes[0] ?? "task");

  const showAcceptanceCriteria =
    effectiveType === "story" ||
    templateRequiresAcceptanceCriteria(template, effectiveType);
  const usesStoryDescription = templateUsesUserStoryDescription(
    template,
    effectiveType,
  );

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      onError("Work Item title is required");
      return;
    }
    if (!templateReady || !template) {
      onError("Project template is still loading. Try again in a moment.");
      return;
    }

    const criteria = acceptanceCriteriaTexts(acceptanceCriteria);
    setCreating(true);
    onError("");
    try {
      if (template.mode === "se") {
        const validation = await validateWorkItemAttributes(projectId, {
          type: toApiWorkItemType(effectiveType),
          title: trimmedTitle,
          description: description.trim(),
          story_points: storyPoints,
          acceptance_criteria: criteria,
        });
        if (!validation.valid) {
          onError(formatValidationViolations(validation.violations));
          return;
        }
      }

      // Picking an Auditor also assigns them to the Project (a no-op if they
      // already are), so they can see and review it.
      if (auditorId) await assignProjectAuditor(projectId, auditorId);
      const item = await createWorkItem({
        projectId,
        title: trimmedTitle,
        description: description.trim(),
        type: effectiveType,
        priority,
        storyPoints,
        sprintId: sprintId || null,
        assigneeId: assigneeId || undefined,
        reporterId: auditorId || undefined,
        acceptanceCriteria: criteria,
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
            placeholder={
              usesStoryDescription
                ? "As a … I want … so that …"
                : "Add context for the team"
            }
            rows={usesStoryDescription ? 3 : 2}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
          Type
          <select
            value={effectiveType}
            onChange={(event) => setType(event.target.value as WorkItemType)}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          >
            {allowedTypes.map((allowedType) => (
              <option key={allowedType} value={allowedType}>
                {typeLabel(allowedType)}
              </option>
            ))}
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
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
          Assignee
          <select
            value={assigneeId}
            onChange={(event) => setAssigneeId(event.target.value)}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          >
            <option value="">Unassigned</option>
            {members.map((member) => (
              <option key={member.id} value={member.id}>{member.name}</option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
          Auditor
          <select
            value={auditorId}
            onChange={(event) => setAuditorId(event.target.value)}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
          >
            <option value="">No Auditor</option>
            {auditors.map((auditor) => (
              <option key={auditor.id} value={auditor.id}>{auditor.name || auditor.email}</option>
            ))}
          </select>
          <span className="font-normal text-text-tertiary">
            Also assigns them to this Project. They are emailed about status changes.
          </span>
        </label>
        {showAcceptanceCriteria && (
          <div className="flex flex-col gap-2 sm:col-span-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-text-secondary">
                Acceptance criteria{templateRequiresAcceptanceCriteria(template, type) ? " *" : ""}
              </span>
              <button
                type="button"
                onClick={() => setAcceptanceCriteria((current) => [...current, ""])}
                className="text-xs font-bold text-brand"
              >
                Add criterion
              </button>
            </div>
            {acceptanceCriteria.map((criterion, index) => (
              <div key={`ac-${index}`} className="flex gap-2">
                <input
                  value={criterion}
                  onChange={(event) => {
                    const next = [...acceptanceCriteria];
                    next[index] = event.target.value;
                    setAcceptanceCriteria(next);
                  }}
                  placeholder="Given / when / then…"
                  className="min-w-0 flex-1 rounded-lg border border-border bg-surface px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
                />
                <button
                  type="button"
                  onClick={() =>
                    setAcceptanceCriteria((current) =>
                      current.length <= 1
                        ? [""]
                        : current.filter((_, currentIndex) => currentIndex !== index),
                    )
                  }
                  className="rounded-lg border border-border px-2 text-xs font-bold text-text-secondary"
                  aria-label={`Remove criterion ${index + 1}`}
                >
                  Remove
                </button>
              </div>
            ))}
          </div>
        )}
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
          disabled={creating || !templateReady}
          className="fl-cta px-4 py-2 text-xs"
        >
          {creating ? "Creating..." : "Create work item"}
        </button>
      </div>
    </form>
  );
}

export default function BacklogPage() {
  const {
    selectedProjectId,
    selectedProject,
    selectedTemplate,
    selectedProjectLoadState,
    loadState: projectsLoadState,
    currentUser,
  } = useProjects();
  const canMutate = canMutateExecution(currentUser?.roles);
  const templateReady = selectedProjectLoadState === "ready" && selectedTemplate !== null;
  const [workItems, setWorkItems] = React.useState<WorkItem[]>([]);
  // Members come from the Project the provider already loaded.
  const members = React.useMemo(() => membersOf(selectedProject), [selectedProject]);
  const [auditors, setAuditors] = React.useState<AuditorDirectoryEntry[]>([]);
  React.useEffect(() => {
    if (!canMutate) return;
    let active = true;
    // The Auditor picker is optional: if the list fails, the form still works.
    listAuditorDirectory()
      .then((items) => {
        if (active) setAuditors(items);
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, [canMutate]);
  const [sprints, setSprints] = React.useState<Awaited<ReturnType<typeof listSprints>>>([]);
  const [listError, setListError] = React.useState<string>();
  const [loading, setLoading] = React.useState(true);
  const [openCreateFor, setOpenCreateFor] = React.useState<string | null>(null);

  /* Clear/refetch on Project change; sync resets avoid showing the previous Project. */
  /* eslint-disable react-hooks/set-state-in-effect -- project-scoped list reset before fetch */
  React.useEffect(() => {
    let active = true;
    setWorkItems([]);
    setSprints([]);
    setListError(undefined);
    setOpenCreateFor(null);

    if (!selectedProjectId) {
      setLoading(false);
      return () => {
        active = false;
      };
    }

    setLoading(true);
    void Promise.all([
      listSprints(selectedProjectId),
      listWorkItems({ projectId: selectedProjectId }),
    ])
      .then(([availableSprints, items]) => {
        if (!active) return;
        setSprints(availableSprints);
        setWorkItems(items);
      })
      .catch((error) => {
        if (active) {
          setListError(
            error instanceof Error ? error.message : "Unable to load backlog",
          );
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [selectedProjectId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const membersById = new Map(members.map((member) => [member.id, member]));

  const handleSprintChange = async (item: WorkItem, nextSprintId: string | null) => {
    if (!canMutate) return;
    setListError(undefined);
    try {
      await updateWorkItem({
        ...item,
        id: item.id,
        projectId: item.projectId ?? selectedProjectId ?? undefined,
        sprintId: nextSprintId,
        acceptanceCriteria: item.acceptanceCriteria,
        blockedReason: item.blockedReason,
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
          <p className="m-0 text-[11px] font-bold uppercase tracking-wider text-text-tertiary">
            {selectedProject?.name ?? "Project"}
          </p>
          <h1 className="m-0 mt-1 text-2xl font-bold tracking-tight">Backlog</h1>
        </div>
      </div>

      {listError && (
        <p className="m-0 border-b border-border px-8 py-3 text-sm text-red-600">
          {listError}
        </p>
      )}

      {projectsLoadState === "loading" || (selectedProjectId && loading) ? (
        <div className="flex flex-1 items-center justify-center text-sm text-text-tertiary">
          Loading backlog...
        </div>
      ) : !selectedProjectId ? (
        <NoProjectSelected title="Select a Project to open the Backlog" />
      ) : (
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
                      canMutate ? (
                        <button
                          type="button"
                          disabled={!isCreating && !templateReady}
                          onClick={() => {
                            setOpenCreateFor(isCreating ? null : section.id);
                            setListError(undefined);
                          }}
                          className="fl-cta px-3 py-1.5 text-xs"
                          title={
                            !isCreating && !templateReady
                              ? "Waiting for project template…"
                              : undefined
                          }
                        >
                          {isCreating ? "Cancel" : "Add work item"}
                        </button>
                      ) : null
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
                        onSprintChange={
                          canMutate
                            ? (nextSprintId) =>
                                void handleSprintChange(item, nextSprintId)
                            : undefined
                        }
                      />
                    ))
                  )}
                </SectionCard>
              );
            })}
          </div>
        </div>
      )}
      {canMutate && selectedProjectId && openCreateFor !== null && (
        <Modal
          title="Create work item"
          onClose={() => setOpenCreateFor(null)}
        >
          <CreateWorkItemForm
            projectId={selectedProjectId}
            template={selectedTemplate}
            templateReady={templateReady}
            members={members}
            auditors={auditors}
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

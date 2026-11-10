"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import ReactMarkdown from "react-markdown";
import { assignProjectAuditor, getProject } from "@/api/projects";
import {
  deleteWorkItem,
  updateWorkItem,
  validateWorkItemAttributes,
} from "@/api/work-items";
import type {
  AcceptanceCriterion,
  Member,
  Priority,
  Sprint,
  WorkItem,
  WorkItemStatus,
  WorkItemType,
} from "@/types/work-item";
import { TypeBadge, typeLabel } from "@/components/work-items/type-badge";
import { PriorityIcon } from "@/components/work-items/priority-icon";
import { EpicPill } from "@/components/work-items/epic-pill";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { WorkItemComments } from "@/components/work-items/work-item-comments";
import { ChevronRightIcon, PlusIcon, XIcon } from "@/components/icons";
import { useProjects } from "@/components/layout/project-provider";
import { canMutateExecution } from "@/lib/roles";
import {
  STATUS_LABELS,
  STATUS_ORDER,
  acceptanceCriteriaTexts,
  formatValidationViolations,
  templateRequiresAcceptanceCriteria,
  templateUsesUserStoryDescription,
  toApiWorkItemType,
  workItemTypesFromTemplate,
} from "@/lib/work-item-utils";
import { cn } from "@/lib/cn";

interface WorkItemDetailEditorProps {
  initialItem: WorkItem;
  members: Member[];
  /** Every Auditor in the system; the Work Item's Auditor is stored as reporterId. */
  auditors?: Member[];
  sprints: Sprint[];
}

type EditableWorkItem = Pick<
  WorkItem,
  | "projectId"
  | "sprintId"
  | "title"
  | "description"
  | "type"
  | "status"
  | "priority"
  | "assigneeId"
  | "reporterId"
  | "storyPoints"
  | "features"
  | "acceptanceCriteria"
  | "blockedReason"
>;

/**
 * One custom field row. `id` keeps the row (and its inputs) stable while the
 * field name is typed; keying rows by the name remounted them on every
 * keystroke, so only one character could be typed at a time.
 */
interface FeatureEntry {
  id: string;
  key: string;
  value: string;
}

const MANAGED_FEATURE_KEYS = new Set(["acceptance_criteria", "blocked_reason"]);

let featureRowCounter = 0;
const nextFeatureRowId = () => `feature-${(featureRowCounter += 1)}`;

function featureEntriesFromValue(features?: Record<string, unknown>): FeatureEntry[] {
  return Object.entries(features ?? {})
    .filter(([key]) => !MANAGED_FEATURE_KEYS.has(key))
    .map(([key, value]) => ({
      id: nextFeatureRowId(),
      key,
      value: typeof value === "string" ? value : JSON.stringify(value),
    }));
}

function featureValueFromInput(value: string): unknown {
  try {
    return JSON.parse(value);
  } catch {
    return value;
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function WorkItemDetailEditor({
  initialItem,
  members,
  auditors = [],
  sprints,
}: WorkItemDetailEditorProps) {
  const router = useRouter();
  const {
    projects,
    selectedProject,
    selectedTemplate,
    selectedProjectLoadState,
    currentUser,
  } = useProjects();
  const canMutate = canMutateExecution(currentUser?.roles);
  const [item, setItem] = useState(initialItem);
  const knownProjectName =
    projects.find((project) => project.id === item.projectId)?.name ??
    (selectedProject && selectedProject.id === item.projectId
      ? selectedProject.name
      : undefined);
  const [fetchedProjectName, setFetchedProjectName] = useState<string>();
  const projectName = knownProjectName ?? fetchedProjectName ?? "Unknown project";
  const allowedTypes = workItemTypesFromTemplate(selectedTemplate);

  useEffect(() => {
    if (knownProjectName || !item.projectId) return;
    let active = true;
    getProject(item.projectId)
      .then((project) => {
        if (active) setFetchedProjectName(project.name);
      })
      .catch(() => {
        if (active) setFetchedProjectName(undefined);
      });
    return () => {
      active = false;
    };
  }, [item.projectId, knownProjectName]);

  const [descriptionDraft, setDescriptionDraft] = useState(initialItem.description ?? "");
  const [featureEntries, setFeatureEntries] = useState(() =>
    featureEntriesFromValue(initialItem.features),
  );
  const [acceptanceDraft, setAcceptanceDraft] = useState<AcceptanceCriterion[]>(
    () =>
      initialItem.acceptanceCriteria.length > 0
        ? initialItem.acceptanceCriteria
        : [{ text: "", done: false }],
  );
  const [blockedReasonDraft, setBlockedReasonDraft] = useState(
    initialItem.blockedReason ?? "",
  );
  const [editingDescription, setEditingDescription] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string>();
  const [featuresError, setFeaturesError] = useState<string>();

  const showAcceptanceCriteria =
    item.type === "story" ||
    templateRequiresAcceptanceCriteria(selectedTemplate, item.type);
  const usesStoryDescription = templateUsesUserStoryDescription(
    selectedTemplate,
    item.type,
  );

  const saveChanges = async (patch: Partial<EditableWorkItem>) => {
    if (!canMutate) return;
    const nextItem = { ...item, ...patch };
    const nextStatus = nextItem.status;
    const nextBlockedReason =
      nextStatus === "blocked"
        ? (patch.blockedReason ?? nextItem.blockedReason ?? blockedReasonDraft).trim()
        : undefined;

    if (nextStatus === "blocked" && !nextBlockedReason) {
      setError("A blocked reason is required");
      return;
    }

    setItem({
      ...nextItem,
      blocked: nextStatus === "blocked",
      blockedReason: nextBlockedReason,
    });
    setSaving(true);
    setError(undefined);

    try {
      const projectId = nextItem.projectId;
      if (selectedProjectLoadState === "loading") {
        setItem(item);
        setError("Project template is still loading. Try again in a moment.");
        return;
      }
      if (selectedTemplate?.mode === "se" && projectId) {
        const validation = await validateWorkItemAttributes(projectId, {
          type: toApiWorkItemType(nextItem.type),
          title: nextItem.title,
          description: nextItem.description ?? "",
          story_points: nextItem.storyPoints,
          acceptance_criteria: acceptanceCriteriaTexts(nextItem.acceptanceCriteria),
        });
        if (!validation.valid) {
          setItem(item);
          setError(formatValidationViolations(validation.violations));
          return;
        }
      }

      const updatedItem = await updateWorkItem({
        id: nextItem.id,
        projectId: nextItem.projectId,
        sprintId: nextItem.sprintId,
        title: nextItem.title,
        description: nextItem.description,
        type: nextItem.type,
        status: nextItem.status,
        priority: nextItem.priority,
        assigneeId: nextItem.assigneeId,
        reporterId: nextItem.reporterId,
        storyPoints: nextItem.storyPoints,
        features: nextItem.features,
        acceptanceCriteria: nextItem.acceptanceCriteria,
        blockedReason: nextBlockedReason,
      });
      setItem(updatedItem);
      setDescriptionDraft(updatedItem.description ?? "");
      setFeatureEntries(featureEntriesFromValue(updatedItem.features));
      setAcceptanceDraft(
        updatedItem.acceptanceCriteria.length > 0
          ? updatedItem.acceptanceCriteria
          : [{ text: "", done: false }],
      );
      setBlockedReasonDraft(updatedItem.blockedReason ?? "");
    } catch (saveError) {
      setItem(item);
      setError(saveError instanceof Error ? saveError.message : "Unable to update work item");
    } finally {
      setSaving(false);
    }
  };

  const updateDescription = async () => {
    if (descriptionDraft === (item.description ?? "")) {
      setEditingDescription(false);
      return;
    }
    await saveChanges({ description: descriptionDraft });
    setEditingDescription(false);
  };

  const updateFeatures = async (entries: FeatureEntry[]) => {
    const normalizedEntries = entries.map((entry) => ({
      ...entry,
      key: entry.key.trim(),
    }));
    if (normalizedEntries.some((entry) => !entry.key)) {
      setFeaturesError("Every custom field needs a field name");
      return;
    }
    if (new Set(normalizedEntries.map((entry) => entry.key)).size !== normalizedEntries.length) {
      setFeaturesError("Field names must be unique");
      return;
    }
    if (normalizedEntries.some((entry) => MANAGED_FEATURE_KEYS.has(entry.key))) {
      setFeaturesError(
        "acceptance_criteria and blocked_reason are managed by dedicated fields",
      );
      return;
    }

    const features = Object.fromEntries(
      normalizedEntries.map((entry) => [entry.key, featureValueFromInput(entry.value)]),
    );
    setFeaturesError(undefined);
    await saveChanges({ features });
  };

  const updateAcceptanceCriteria = async (criteria: AcceptanceCriterion[]) => {
    setAcceptanceDraft(criteria);
    await saveChanges({ acceptanceCriteria: criteria });
  };

  const handleStatusChange = async (status: WorkItemStatus) => {
    if (status === "blocked") {
      const reason = blockedReasonDraft.trim() || window.prompt("Why is this work item blocked?");
      if (!reason?.trim()) {
        setError("A blocked reason is required");
        return;
      }
      setBlockedReasonDraft(reason.trim());
      await saveChanges({ status, blockedReason: reason.trim() });
      return;
    }
    await saveChanges({ status, blockedReason: undefined });
  };

  const handleDelete = async () => {
    if (!canMutate) return;
    if (!window.confirm(`Delete ${item.id}? This cannot be undone.`)) return;

    setDeleting(true);
    setError(undefined);
    try {
      await deleteWorkItem(item.id);
      router.push("/backlog");
      router.refresh();
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : "Unable to delete work item");
      setDeleting(false);
    }
  };

  const assignee = members.find((member) => member.id === item.assigneeId);
  const auditor = auditors.find((entry) => entry.id === item.reporterId);

  /** Picking an Auditor first assigns them to the Project, then saves the item. */
  const pickAuditor = async (auditorId: string) => {
    if (auditorId && item.projectId) {
      try {
        await assignProjectAuditor(item.projectId, auditorId);
      } catch (assignError) {
        setError(
          assignError instanceof Error
            ? assignError.message
            : "Unable to assign the Auditor to this Project",
        );
        return;
      }
    }
    await saveChanges({ reporterId: auditorId || undefined });
  };
  const typeOptions = allowedTypes.includes(item.type)
    ? allowedTypes
    : [item.type, ...allowedTypes];

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <div className="flex items-center justify-between border-b border-border bg-surface px-8 py-3.5">
        <div className="flex items-center gap-2 text-[13px] font-semibold text-text-tertiary">
          <span>Chura Web App</span>
          <ChevronRightIcon className="size-3" />
          <Link href="/backlog" className="hover:text-text-secondary">
            Backlog
          </Link>
          <ChevronRightIcon className="size-3" />
          <span className="truncate text-text">{item.id}</span>
        </div>
        <div className="flex items-center gap-2.5">
          {saving && <span className="text-xs text-text-tertiary">Saving...</span>}
          <Link href="/backlog" aria-label="Close">
            <XIcon className="size-4 text-text-tertiary" />
          </Link>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex min-w-0 flex-1 flex-col gap-[22px] overflow-y-auto p-8">
          {error && (
            <p className="m-0 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {error}
            </p>
          )}

          <div>
            <div className="mb-3 flex items-center gap-2">
              <TypeBadge type={item.type} />
              <span className="font-mono text-[12.5px] font-bold text-text-tertiary">
                {item.id}
              </span>
              <span className="text-[12.5px] font-semibold text-text-tertiary">
                &middot; {typeLabel(item.type)}
              </span>
            </div>
            {canMutate ? (
              <input
                value={item.title}
                onChange={(event) => setItem({ ...item, title: event.target.value })}
                onBlur={() => saveChanges({ title: item.title })}
                className="w-full bg-transparent text-[22px] font-extrabold leading-snug tracking-tight text-text outline-none ring-0"
                aria-label="Work Item title"
              />
            ) : (
              <h1 className="m-0 text-[22px] font-extrabold leading-snug tracking-tight text-text">
                {item.title}
              </h1>
            )}
          </div>

          <section>
            <div className="mb-2.5 flex items-center justify-between">
              <p className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary">
                Description
              </p>
              {canMutate && (
                <button
                  type="button"
                  onClick={() => setEditingDescription(!editingDescription)}
                  className="text-xs font-bold text-brand hover:text-brand-dark"
                >
                  {editingDescription ? "Preview" : "Edit Markdown"}
                </button>
              )}
            </div>
            {canMutate && editingDescription ? (
              <div className="flex flex-col gap-2">
                <textarea
                  value={descriptionDraft}
                  onChange={(event) => setDescriptionDraft(event.target.value)}
                  placeholder={
                    usesStoryDescription
                      ? "As a … I want … so that …"
                      : "Add a description"
                  }
                  className="min-h-32 w-full resize-y rounded-lg border border-border bg-background p-3 text-[13.5px] leading-relaxed text-text outline-none focus:border-brand"
                  aria-label="Work Item description Markdown"
                />
                <button
                  type="button"
                  onClick={updateDescription}
                  className="fl-cta self-start px-3 py-1.5 text-xs"
                >
                  Save description
                </button>
              </div>
            ) : (
              <div className="work-item-markdown max-w-2xl text-[13.5px] leading-relaxed text-text">
                {descriptionDraft ? (
                  <ReactMarkdown>{descriptionDraft}</ReactMarkdown>
                ) : (
                  <p className="text-text-tertiary">No description</p>
                )}
              </div>
            )}
          </section>

          {showAcceptanceCriteria && (
            <div>
              <div className="mb-2.5 flex items-center justify-between">
                <p className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary">
                  Acceptance Criteria
                </p>
                {canMutate && (
                  <button
                    type="button"
                    onClick={() =>
                      setAcceptanceDraft((current) => [
                        ...current,
                        { text: "", done: false },
                      ])
                    }
                    className="flex items-center gap-1 text-xs font-bold text-brand hover:text-brand-dark"
                  >
                    <PlusIcon className="size-3.5" />
                    Add criterion
                  </button>
                )}
              </div>
              <div className="flex flex-col gap-2.5">
                {acceptanceDraft.map((criterion, index) => (
                  <div key={`ac-${index}`} className="flex items-start gap-2.5">
                    <input
                      type="checkbox"
                      checked={criterion.done}
                      disabled={!canMutate}
                      onChange={(event) => {
                        const next = acceptanceDraft.map((entry, currentIndex) =>
                          currentIndex === index
                            ? { ...entry, done: event.target.checked }
                            : entry,
                        );
                        void updateAcceptanceCriteria(next);
                      }}
                      className="mt-2"
                      aria-label={`Mark criterion ${index + 1} done`}
                    />
                    <input
                      value={criterion.text}
                      readOnly={!canMutate}
                      onChange={(event) => {
                        const next = acceptanceDraft.map((entry, currentIndex) =>
                          currentIndex === index
                            ? { ...entry, text: event.target.value }
                            : entry,
                        );
                        setAcceptanceDraft(next);
                      }}
                      onBlur={() => void updateAcceptanceCriteria(acceptanceDraft)}
                      placeholder="Given / when / then…"
                      className="min-w-0 flex-1 rounded-lg border border-border bg-surface px-3 py-2 text-[13px] text-text outline-none focus:border-brand read-only:bg-background"
                      aria-label={`Acceptance criterion ${index + 1}`}
                    />
                    {canMutate && (
                      <button
                        type="button"
                        onClick={() => {
                          const next =
                            acceptanceDraft.length <= 1
                              ? [{ text: "", done: false }]
                              : acceptanceDraft.filter(
                                  (_, currentIndex) => currentIndex !== index,
                                );
                          void updateAcceptanceCriteria(next);
                        }}
                        className="mt-2 text-text-tertiary hover:text-red-600"
                        aria-label={`Remove criterion ${index + 1}`}
                      >
                        <XIcon className="size-4" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="mb-2 flex items-center justify-between">
              <p className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary">
                Custom fields
              </p>
              {canMutate && (
                <button
                  type="button"
                  onClick={() =>
                    setFeatureEntries([
                      ...featureEntries,
                      { id: nextFeatureRowId(), key: "", value: "" },
                    ])
                  }
                  className="flex items-center gap-1 text-xs font-bold text-brand hover:text-brand-dark"
                >
                  <PlusIcon className="size-3.5" />
                  Add field
                </button>
              )}
            </div>
            <div className="mb-[22px] flex flex-col gap-2">
              {featureEntries.length === 0 && !canMutate ? (
                <p className="m-0 text-sm text-text-tertiary">No custom fields</p>
              ) : (
                featureEntries.map((entry, index) => (
                  // Keyed by a stable id, not the field name: keying by the
                  // name rebuilt the row on every keystroke and dropped focus.
                  <div key={entry.id} className="flex items-center gap-2">
                    <input
                      value={entry.key}
                      readOnly={!canMutate}
                      onChange={(event) => {
                        const next = featureEntries.map((current) =>
                          current.id === entry.id ? { ...current, key: event.target.value } : current,
                        );
                        setFeatureEntries(next);
                      }}
                      onBlur={() => updateFeatures(featureEntries)}
                      className="w-2/5 rounded-lg border border-border bg-surface px-3 py-2 text-[13px] font-semibold text-text outline-none focus:border-brand read-only:bg-background"
                      aria-label={`Custom field ${index + 1} name`}
                      placeholder="Field name, e.g. Environment"
                    />
                    <input
                      value={entry.value}
                      readOnly={!canMutate}
                      onChange={(event) => {
                        const next = featureEntries.map((current) =>
                          current.id === entry.id ? { ...current, value: event.target.value } : current,
                        );
                        setFeatureEntries(next);
                      }}
                      onBlur={() => updateFeatures(featureEntries)}
                      className="min-w-0 flex-1 rounded-lg border border-border bg-surface px-3 py-2 text-[13px] text-text outline-none focus:border-brand read-only:bg-background"
                      aria-label={`Custom field ${index + 1} details`}
                      placeholder="Details, e.g. Staging"
                    />
                    {canMutate && (
                      <button
                        type="button"
                        onClick={() => {
                          const next = featureEntries.filter((current) => current.id !== entry.id);
                          setFeatureEntries(next);
                          void updateFeatures(next);
                        }}
                        className="shrink-0 text-text-tertiary hover:text-red-600"
                        aria-label={`Remove custom field ${index + 1}`}
                        title="Remove field"
                      >
                        <XIcon className="size-4" />
                      </button>
                    )}
                  </div>
                ))
              )}
              {featuresError && <p className="m-0 text-xs text-red-600">{featuresError}</p>}
            </div>
          </div>

          <WorkItemComments workItemId={item.id} />
        </div>

        <div className="w-[340px] shrink-0 overflow-y-auto border-l border-border bg-surface p-[26px]">
          {!canMutate && (
            <p className="m-0 mb-3 rounded-md border border-border bg-background px-2.5 py-2 text-[11px] font-semibold text-text-tertiary">
              Auditor view — Work Item fields are read-only. Comments remain available after you self-assign in Settings.
            </p>
          )}
          <p className="m-0 mb-3 text-xs font-bold uppercase tracking-wide text-text-tertiary">
            Status
          </p>
          <select
            value={item.status}
            disabled={!canMutate}
            onChange={(event) =>
              void handleStatusChange(event.target.value as WorkItemStatus)
            }
            className="mb-3 w-full rounded-lg border border-border bg-background px-3 py-2 text-xs font-bold text-text outline-none focus:border-brand disabled:opacity-70"
            aria-label="Work Item status"
          >
            {STATUS_ORDER.map((status) => (
              <option key={status} value={status}>
                {STATUS_LABELS[status]}
              </option>
            ))}
          </select>
          {item.status === "blocked" && (
            <label className="mb-6 flex flex-col gap-1 text-xs font-semibold text-text-secondary">
              Blocked reason *
              <textarea
                value={blockedReasonDraft}
                readOnly={!canMutate}
                onChange={(event) => setBlockedReasonDraft(event.target.value)}
                onBlur={() => {
                  if (blockedReasonDraft.trim() !== (item.blockedReason ?? "")) {
                    void saveChanges({
                      status: "blocked",
                      blockedReason: blockedReasonDraft.trim(),
                    });
                  }
                }}
                rows={3}
                className="rounded-lg border border-border bg-background px-3 py-2 text-[13px] font-normal text-text outline-none focus:border-brand read-only:opacity-80"
                aria-label="Blocked reason"
              />
            </label>
          )}

          <p className="m-0 mb-1 text-xs font-bold uppercase tracking-wide text-text-tertiary">
            Details
          </p>
          <dl className="m-0">
            <DetailRow label="Project">
              <span className="min-w-0 flex-1 truncate text-[13px] font-semibold text-text">
                {projectName}
              </span>
            </DetailRow>
            <DetailRow label="Assignee">
              <WorkItemAvatar member={assignee} size="sm" />
              <select
                value={item.assigneeId ?? ""}
                disabled={!canMutate}
                onChange={(event) => saveChanges({ assigneeId: event.target.value || undefined })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold outline-none disabled:opacity-70"
                aria-label="Assignee"
              >
                <option value="">Unassigned</option>
                {members.map((member) => (
                  <option key={member.id} value={member.id}>{member.name}</option>
                ))}
              </select>
            </DetailRow>
            <DetailRow label="Auditor">
              <WorkItemAvatar member={auditor} size="sm" />
              <select
                value={auditor ? auditor.id : ""}
                disabled={!canMutate || saving}
                onChange={(event) => void pickAuditor(event.target.value)}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold outline-none disabled:opacity-70"
                aria-label="Auditor"
                title="Picking an Auditor also assigns them to this Project"
              >
                <option value="">No Auditor</option>
                {auditors.map((entry) => (
                  <option key={entry.id} value={entry.id}>{entry.name}</option>
                ))}
              </select>
            </DetailRow>
            <DetailRow label="Sprint">
              <select
                value={item.sprintId ?? ""}
                disabled={!canMutate}
                onChange={(event) => saveChanges({ sprintId: event.target.value || null })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold outline-none disabled:opacity-70"
                aria-label="Sprint"
              >
                <option value="">Backlog</option>
                {sprints.map((sprint) => (
                  <option key={sprint.id} value={sprint.id}>{sprint.name}</option>
                ))}
              </select>
            </DetailRow>
            <DetailRow label="Priority">
              <PriorityIcon priority={item.priority} />
              <select
                value={item.priority}
                disabled={!canMutate}
                onChange={(event) => saveChanges({ priority: event.target.value as Priority })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold capitalize outline-none disabled:opacity-70"
                aria-label="Priority"
              >
                <option value="highest">Highest</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </DetailRow>
            <DetailRow label="Story points">
              <input
                type="number"
                min={0}
                value={item.storyPoints}
                readOnly={!canMutate}
                onChange={(event) => setItem({ ...item, storyPoints: Number(event.target.value) })}
                onBlur={() => saveChanges({ storyPoints: item.storyPoints })}
                className="w-16 rounded border border-border bg-background px-2 py-1 text-xs font-bold outline-none focus:border-brand read-only:opacity-70"
                aria-label="Story points"
              />
            </DetailRow>
            <DetailRow label="Type">
              <TypeBadge type={item.type} />
              <select
                value={item.type}
                disabled={!canMutate}
                onChange={(event) => saveChanges({ type: event.target.value as WorkItemType })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold outline-none disabled:opacity-70"
                aria-label="Work Item type"
              >
                {typeOptions.map((type) => (
                  <option key={type} value={type}>
                    {type === "story" ? "Story" : type === "bug" ? "Bug" : "Task"}
                  </option>
                ))}
              </select>
            </DetailRow>
            <DetailRow label="Epic">
              <EpicPill epic={item.epic} />
            </DetailRow>
            {item.labels.length > 0 && (
              <DetailRow label="Labels" noBorder>
                {item.labels.map((label) => (
                  <span
                    key={label}
                    className="rounded-full border border-border bg-background px-2.5 py-0.5 text-[11px] font-bold text-text-secondary"
                  >
                    {label}
                  </span>
                ))}
              </DetailRow>
            )}
          </dl>

          <div className="mt-5 flex flex-col gap-1.5 border-border pt-4">
            <div className="flex justify-between text-xs text-text-tertiary">
              <span>Created</span>
              <span className="font-semibold text-text-secondary">{formatDate(item.createdAt)}</span>
            </div>
            <div className="flex justify-between text-xs text-text-tertiary">
              <span>Updated</span>
              <span className="font-semibold text-text-secondary">{formatDate(item.updatedAt)}</span>
            </div>
            {canMutate && (
              <div className="mt-5 flex w-full justify-center text-xs text-danger">
                <button
                  type="button"
                  onClick={handleDelete}
                  disabled={deleting || saving}
                  className="w-full rounded-md bg-background/70 px-2.5 py-1.5 text-[12.5px] font-bold text-danger hover:bg-background hover:text-danger disabled:cursor-not-allowed disabled:opacity-50"
                  title="Delete work item"
                >
                  {deleting ? "Deleting..." : "Delete work item permanently"}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function DetailRow({
  label,
  children,
  noBorder,
}: {
  label: string;
  children: React.ReactNode;
  noBorder?: boolean;
}) {
  return (
    <div
      className={cn(
        "flex items-center gap-1.5 py-2.5",
        !noBorder && "border-b border-border",
      )}
    >
      <dt className="w-[104px] shrink-0 text-xs font-semibold text-text-tertiary">{label}</dt>
      <dd className="m-0 flex min-w-0 flex-1 items-center gap-2 text-[13px] font-semibold text-text">
        {children}
      </dd>
    </div>
  );
}

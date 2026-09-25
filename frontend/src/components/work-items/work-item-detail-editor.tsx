"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import ReactMarkdown from "react-markdown";
import { deleteWorkItem, updateWorkItem } from "@/api/work-items";
import type { Member, WorkItem, WorkItemComment, WorkItemStatus, WorkItemType, Priority } from "@/types/work-item";
import { TypeBadge, typeLabel } from "@/components/work-items/type-badge";
import { PriorityIcon } from "@/components/work-items/priority-icon";
import { EpicPill } from "@/components/work-items/epic-pill";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { ChevronRightIcon, LinkIcon, PlusIcon, XIcon } from "@/components/icons";
import { STATUS_LABELS, STATUS_ORDER } from "@/lib/work-item-utils";
import { cn } from "@/lib/cn";

interface WorkItemDetailEditorProps {
  initialItem: WorkItem;
  comments: WorkItemComment[];
  members: Member[];
}

type EditableWorkItem = Pick<
  WorkItem,
  | "projectId"
  | "title"
  | "description"
  | "type"
  | "status"
  | "priority"
  | "assigneeId"
  | "reporterId"
  | "storyPoints"
  | "features"
>;

interface FeatureEntry {
  key: string;
  value: string;
}

function featureEntriesFromValue(features?: Record<string, unknown>): FeatureEntry[] {
  return Object.entries(features ?? {}).map(([key, value]) => ({
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

function formatDateTime(value: string) {
  return new Date(value).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function WorkItemDetailEditor({
  initialItem,
  comments,
  members,
}: WorkItemDetailEditorProps) {
  const router = useRouter();
  const [item, setItem] = useState(initialItem);
  const [descriptionDraft, setDescriptionDraft] = useState(initialItem.description ?? "");
  const [featureEntries, setFeatureEntries] = useState(() =>
    featureEntriesFromValue(initialItem.features),
  );
  const [editingDescription, setEditingDescription] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string>();
  const [featuresError, setFeaturesError] = useState<string>();

  const saveChanges = async (patch: Partial<EditableWorkItem>) => {
    const nextItem = { ...item, ...patch };
    setItem(nextItem);
    setSaving(true);
    setError(undefined);

    try {
      const updatedItem = await updateWorkItem({
        id: nextItem.id,
        projectId: nextItem.projectId,
        title: nextItem.title,
        description: nextItem.description,
        type: nextItem.type,
        status: nextItem.status,
        priority: nextItem.priority,
        assigneeId: nextItem.assigneeId,
        reporterId: nextItem.reporterId,
        storyPoints: nextItem.storyPoints,
        features: nextItem.features,
      });
      setItem(updatedItem);
      setDescriptionDraft(updatedItem.description ?? "");
      setFeatureEntries(featureEntriesFromValue(updatedItem.features));
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
      setFeaturesError("Feature keys cannot be empty");
      return;
    }
    if (new Set(normalizedEntries.map((entry) => entry.key)).size !== normalizedEntries.length) {
      setFeaturesError("Feature keys must be unique");
      return;
    }

    const features = Object.fromEntries(
      normalizedEntries.map((entry) => [entry.key, featureValueFromInput(entry.value)]),
    );
    setFeaturesError(undefined);
    await saveChanges({ features });
  };

  const handleDelete = async () => {
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
  const reporter = members.find((member) => member.id === item.reporterId);

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
            <input
              value={item.title}
              onChange={(event) => setItem({ ...item, title: event.target.value })}
              onBlur={() => saveChanges({ title: item.title })}
              className="w-full bg-transparent text-[22px] font-extrabold leading-snug tracking-tight text-text outline-none ring-0"
              aria-label="Work Item title"
            />
          </div>

          <section>
            <div className="mb-2.5 flex items-center justify-between">
              <p className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary">
                Description
              </p>
              <button
                type="button"
                onClick={() => setEditingDescription(!editingDescription)}
                className="text-xs font-bold text-brand hover:text-brand-dark"
              >
                {editingDescription ? "Preview" : "Edit Markdown"}
              </button>
            </div>
            {editingDescription ? (
              <div className="flex flex-col gap-2">
                <textarea
                  value={descriptionDraft}
                  onChange={(event) => setDescriptionDraft(event.target.value)}
                  className="min-h-32 w-full resize-y rounded-lg border border-border bg-background p-3 text-[13.5px] leading-relaxed text-text outline-none focus:border-brand"
                  aria-label="Work Item description Markdown"
                />
                <button
                  type="button"
                  onClick={updateDescription}
                  className="self-start rounded-lg bg-brand px-3 py-1.5 text-xs font-bold text-white"
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

          {item.acceptanceCriteria.length > 0 && (
            <div>
              <p className="m-0 mb-2.5 text-xs font-bold uppercase tracking-wide text-text-tertiary">
                Acceptance Criteria
              </p>
              <div className="flex flex-col gap-2.5">
                {item.acceptanceCriteria.map((criterion) => (
                  <div key={criterion.text} className="flex items-start gap-2.5">
                    <div
                      className={cn(
                        "mt-px flex size-4 shrink-0 items-center justify-center rounded",
                        criterion.done
                          ? "bg-story text-white"
                          : "border-[1.5px] border-border-strong",
                      )}
                    >
                      {criterion.done && (
                        <svg
                          viewBox="0 0 24 24"
                          className="size-[11px]"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth={3}
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        >
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                      )}
                    </div>
                    <span
                      className={cn(
                        "text-[13px]",
                        criterion.done
                          ? "text-text-secondary line-through decoration-border-strong opacity-60"
                          : "text-text",
                      )}
                    >
                      {criterion.text}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          <div>
            <div className="mb-2 flex items-center justify-between">
              <p className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary">
                Features
              </p>
              <button
                type="button"
                onClick={() => setFeatureEntries([...featureEntries, { key: "", value: "" }])}
                className="flex items-center gap-1 text-xs font-bold text-brand hover:text-brand-dark"
              >
                <PlusIcon className="size-3.5" />
                Add feature
              </button>
            </div>
            <div className="mb-[22px] flex flex-col gap-2">
              {featureEntries.map((entry, index) => (
                <div key={`${entry.key}-${index}`} className="flex items-center gap-2">
                  <input
                    value={entry.key}
                    onChange={(event) => {
                      const next = featureEntries.map((current, currentIndex) =>
                        currentIndex === index ? { ...current, key: event.target.value } : current,
                      );
                      setFeatureEntries(next);
                    }}
                    onBlur={() => updateFeatures(featureEntries)}
                    className="w-2/5 rounded-lg border border-border bg-surface px-3 py-2 font-mono text-[12px] text-text outline-none focus:border-brand"
                    aria-label={`Feature ${index + 1} key`}
                    placeholder="Key"
                  />
                  <input
                    value={entry.value}
                    onChange={(event) => {
                      const next = featureEntries.map((current, currentIndex) =>
                        currentIndex === index ? { ...current, value: event.target.value } : current,
                      );
                      setFeatureEntries(next);
                    }}
                    onBlur={() => updateFeatures(featureEntries)}
                    className="min-w-0 flex-1 rounded-lg border border-border bg-surface px-3 py-2 font-mono text-[12px] text-text outline-none focus:border-brand"
                    aria-label={`Feature ${index + 1} value`}
                    placeholder="Value"
                  />
                  <button
                    type="button"
                    onClick={() => {
                      const next = featureEntries.filter((_, currentIndex) => currentIndex !== index);
                      setFeatureEntries(next);
                      void updateFeatures(next);
                    }}
                    className="shrink-0 text-text-tertiary hover:text-red-600"
                    aria-label={`Remove feature ${index + 1}`}
                    title="Remove feature"
                  >
                    <XIcon className="size-4" />
                  </button>
                </div>
              ))}
              {featuresError && <p className="m-0 text-xs text-red-600">{featuresError}</p>}
            </div>
          </div>

          <div>
            <p className="m-0 mb-[18px] text-xs font-bold uppercase tracking-wide text-text-tertiary">
              Activity
            </p>
            <div className="mb-[18px] flex gap-2.5">
              <WorkItemAvatar member={assignee} size="lg" />
              <div className="flex-1 rounded-lg border border-border bg-surface px-3 py-2.5 text-[13px] text-text-tertiary">
                Add a comment&hellip;
              </div>
            </div>
            <div className="flex flex-col gap-4">
              {comments.map((comment) => {
                const author = members.find((member) => member.id === comment.authorId);
                return (
                  <div key={comment.id} className="flex gap-2.5">
                    <WorkItemAvatar member={author} size="lg" />
                    <div className="flex-1">
                      <div className="mb-0.5 flex items-center gap-2">
                        <span className="text-[13px] font-bold">{author?.name ?? "Unknown"}</span>
                        <span className="text-[11.5px] text-text-tertiary">
                          {formatDateTime(comment.createdAt)}
                        </span>
                      </div>
                      <p className="m-0 text-[13px] leading-relaxed text-text-secondary">
                        {comment.body}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        <div className="w-[340px] shrink-0 overflow-y-auto border-l border-border bg-surface p-[26px]">
          <p className="m-0 mb-3 text-xs font-bold uppercase tracking-wide text-text-tertiary">
            Status
          </p>
          <select
            value={item.status}
            onChange={(event) => saveChanges({ status: event.target.value as WorkItemStatus })}
            className="mb-6 w-full rounded-lg border border-border bg-background px-3 py-2 text-xs font-bold text-text outline-none focus:border-brand"
            aria-label="Work Item status"
          >
            {STATUS_ORDER.map((status) => (
              <option key={status} value={status}>
                {STATUS_LABELS[status]}
              </option>
            ))}
          </select>

          <p className="m-0 mb-1 text-xs font-bold uppercase tracking-wide text-text-tertiary">
            Details
          </p>
          <dl className="m-0">
            <DetailRow label="Project ID">
              <input
                value={item.projectId ?? ""}
                onChange={(event) => setItem({ ...item, projectId: event.target.value })}
                onBlur={() => saveChanges({ projectId: item.projectId })}
                className="min-w-0 flex-1 bg-transparent font-mono text-xs outline-none"
                aria-label="Project ID"
              />
            </DetailRow>
            <DetailRow label="Assignee">
              <WorkItemAvatar member={assignee} size="sm" />
              <select
                value={item.assigneeId ?? ""}
                onChange={(event) => saveChanges({ assigneeId: event.target.value || undefined })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold outline-none"
                aria-label="Assignee"
              >
                <option value="">Unassigned</option>
                {members.map((member) => (
                  <option key={member.id} value={member.id}>{member.name}</option>
                ))}
              </select>
            </DetailRow>
            <DetailRow label="Reporter">
              <WorkItemAvatar member={reporter} size="sm" />
              <select
                value={item.reporterId ?? ""}
                onChange={(event) => saveChanges({ reporterId: event.target.value || undefined })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold outline-none"
                aria-label="Reporter"
              >
                <option value="">Unknown</option>
                {members.map((member) => (
                  <option key={member.id} value={member.id}>{member.name}</option>
                ))}
              </select>
            </DetailRow>
            <DetailRow label="Priority">
              <PriorityIcon priority={item.priority} />
              <select
                value={item.priority}
                onChange={(event) => saveChanges({ priority: event.target.value as Priority })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold capitalize outline-none"
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
                onChange={(event) => setItem({ ...item, storyPoints: Number(event.target.value) })}
                onBlur={() => saveChanges({ storyPoints: item.storyPoints })}
                className="w-16 rounded border border-border bg-background px-2 py-1 text-xs font-bold outline-none focus:border-brand"
                aria-label="Story points"
              />
            </DetailRow>
            <DetailRow label="Type">
              <TypeBadge type={item.type} />
              <select
                value={item.type}
                onChange={(event) => saveChanges({ type: event.target.value as WorkItemType })}
                className="min-w-0 flex-1 bg-transparent text-[13px] font-semibold outline-none"
                aria-label="Work Item type"
              >
                <option value="story">Story</option>
                <option value="task">Task</option>
                <option value="bug">Bug</option>
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
            <div className="flex justify-center text-xs mt-5 w-full text-danger">                
                <button
                type="button"
                onClick={handleDelete}
                disabled={deleting || saving}
                className="rounded-md w-full px-2.5 py-1.5 text-[12.5px] font-bold text-danger bg-background/70 hover:bg-background hover:text-danger disabled:cursor-not-allowed disabled:opacity-50"
                title="Delete work item"
            >
                {deleting ? "Deleting..." : "Delete work item permanently"}
            </button>
            </div>
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

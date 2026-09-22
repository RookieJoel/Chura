import type { Metadata } from "next";
import type { ReactNode } from "react";
import Link from "next/link";
import { notFound } from "next/navigation";
import { getWorkItem, listComments } from "@/api/work-items";
import { getMember, listMembers } from "@/api/members";
import { TopNav } from "@/components/layout/top-nav";
import { TypeBadge, typeLabel } from "@/components/work-items/type-badge";
import { PriorityIcon } from "@/components/work-items/priority-icon";
import { EpicPill } from "@/components/work-items/epic-pill";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { ChevronRightIcon, LinkIcon, XIcon } from "@/components/icons";
import { STATUS_LABELS, STATUS_ORDER } from "@/lib/work-item-utils";
import { cn } from "@/lib/cn";

export async function generateMetadata(
  props: PageProps<"/work-items/[id]">,
): Promise<Metadata> {
  const { id } = await props.params;
  const item = await getWorkItem(id);
  return { title: item ? `${item.id} · Chura` : "Work Item · Chura" };
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

export default async function WorkItemDetailPage(
  props: PageProps<"/work-items/[id]">,
) {
  const { id } = await props.params;
  const item = await getWorkItem(id);
  if (!item) notFound();

  const [assignee, reporter, comments] = await Promise.all([
    item.assigneeId ? getMember(item.assigneeId) : undefined,
    item.reporterId ? getMember(item.reporterId) : undefined,
    listComments(item.id),
  ]);
  const membersById = new Map((await listMembers()).map((m) => [m.id, m]));

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <TopNav />

      <div className="flex items-center justify-between border-b border-border bg-surface px-8 py-3.5">
        <div className="flex items-center gap-2 text-[13px] font-semibold text-text-tertiary">
          <span>Chura Web App</span>
          <ChevronRightIcon className="size-3" />
          <Link href="/backlog" className="hover:text-text-secondary">
            Backlog
          </Link>
          <ChevronRightIcon className="size-3" />
          <span className="text-text">{item.id}</span>
        </div>
        <div className="flex items-center gap-2.5">
          <button
            type="button"
            className="flex items-center gap-1.5 rounded-lg border border-border bg-surface px-3 py-1.5 text-[12.5px] font-bold text-text-secondary"
          >
            <LinkIcon className="size-3.5" />
            Copy link
          </button>
          <Link href="/board" aria-label="Close">
            <XIcon className="size-4 text-text-tertiary" />
          </Link>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex min-w-0 flex-1 flex-col gap-[22px] overflow-y-auto p-8">
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
            <h1 className="m-0 text-[22px] font-extrabold leading-snug tracking-tight">
              {item.title}
            </h1>
          </div>

          {item.description && (
            <div>
              <p className="m-0 mb-2.5 text-xs font-bold uppercase tracking-wide text-text-tertiary">
                Description
              </p>
              <p className="m-0 max-w-2xl text-[13.5px] leading-relaxed text-text">
                {item.description}
              </p>
            </div>
          )}

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
                const author = membersById.get(comment.authorId);
                return (
                  <div key={comment.id} className="flex gap-2.5">
                    <WorkItemAvatar member={author} size="lg" />
                    <div className="flex-1">
                      <div className="mb-0.5 flex items-center gap-2">
                        <span className="text-[13px] font-bold">
                          {author?.name ?? "Unknown"}
                        </span>
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
          <div className="mb-6 flex flex-wrap gap-1.5">
            {STATUS_ORDER.map((status) => (
              <span
                key={status}
                className={cn(
                  "rounded-full border border-border px-3.5 py-1.5 text-xs font-bold",
                  status === item.status
                    ? "border-text bg-text text-white"
                    : "bg-surface text-text-secondary",
                )}
              >
                {STATUS_LABELS[status]}
              </span>
            ))}
          </div>

          <p className="m-0 mb-1 text-xs font-bold uppercase tracking-wide text-text-tertiary">
            Details
          </p>
          <dl className="m-0">
            <DetailRow label="Assignee">
              <WorkItemAvatar member={assignee} size="sm" />
              {assignee?.name ?? "Unassigned"}
            </DetailRow>
            <DetailRow label="Reporter">
              <WorkItemAvatar member={reporter} size="sm" />
              {reporter?.name ?? "Unknown"}
            </DetailRow>
            <DetailRow label="Priority">
              <PriorityIcon priority={item.priority} />
              <span className="capitalize">{item.priority}</span>
            </DetailRow>
            <DetailRow label="Story points">
              <span className="flex size-[22px] items-center justify-center rounded-md border border-border-strong text-[10.5px] font-bold text-text-secondary">
                {item.storyPoints}
              </span>
            </DetailRow>
            <DetailRow label="Epic">
              <EpicPill epic={item.epic} />
            </DetailRow>
            <DetailRow label="Type">
              <TypeBadge type={item.type} />
              {typeLabel(item.type)}
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

          <div className="mt-5 flex flex-col gap-1.5 border-t border-border pt-4">
            <div className="flex justify-between text-xs text-text-tertiary">
              <span>Created</span>
              <span className="font-semibold text-text-secondary">
                {formatDate(item.createdAt)}
              </span>
            </div>
            <div className="flex justify-between text-xs text-text-tertiary">
              <span>Updated</span>
              <span className="font-semibold text-text-secondary">
                {formatDate(item.updatedAt)}
              </span>
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
  children: ReactNode;
  noBorder?: boolean;
}) {
  return (
    <div
      className={cn(
        "flex items-center gap-1.5 py-2.5",
        !noBorder && "border-b border-border",
      )}
    >
      <dt className="w-[104px] shrink-0 text-xs font-semibold text-text-tertiary">
        {label}
      </dt>
      <dd className="m-0 flex flex-1 items-center gap-2 text-[13px] font-semibold text-text">
        {children}
      </dd>
    </div>
  );
}

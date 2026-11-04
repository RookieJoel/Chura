import Link from "next/link";
import type { ChangeEvent } from "react";
import type { Member, WorkItem } from "@/types/work-item";
import { TypeBadge } from "@/components/work-items/type-badge";
import { PriorityIcon } from "@/components/work-items/priority-icon";
import { PointsBadge } from "@/components/work-items/points-badge";
import { EpicPill } from "@/components/work-items/epic-pill";
import { BlockedBadge } from "@/components/work-items/blocked-badge";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { cn } from "@/lib/cn";

interface WorkItemCardProps {
  item: WorkItem;
  assignee?: Member;
  className?: string;
  onStatusChange?: (status: WorkItem["status"]) => void;
  saving?: boolean;
}

export function WorkItemCard({
  item,
  assignee,
  className,
  onStatusChange,
  saving = false,
}: WorkItemCardProps) {
  const handleStatusChange = (event: ChangeEvent<HTMLSelectElement>) => {
    event.preventDefault();
    onStatusChange?.(event.target.value as WorkItem["status"]);
  };

  return (
    <Link
      href={`/work-items/${item.id}`}
      className={cn(
        "flex flex-col gap-2 rounded-md border border-border bg-surface p-3 shadow-[0_1px_2px_rgba(9,30,66,0.12)] transition-all hover:border-brand hover:shadow-[0_3px_8px_rgba(9,30,66,0.16)]",
        item.status === "done" && "opacity-90",
        className,
      )}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5">
          <TypeBadge type={item.type} />
          <span className="font-mono text-[11.5px] font-bold text-text-tertiary">
            {item.id}
          </span>
        </div>
        <PriorityIcon priority={item.priority} />
      </div>

      <p
        className={cn(
          "m-0 text-[13px] font-semibold leading-snug",
          item.status === "done"
            ? "text-text-secondary line-through decoration-border-strong"
            : "text-text",
        )}
      >
        {item.title}
      </p>

      {item.blocked && <BlockedBadge />}

      <div className="flex items-center justify-between">
        <EpicPill epic={item.epic} />
        <div className="flex items-center gap-1.5">
          {onStatusChange && (
            <select
              aria-label={`Status for ${item.title}`}
              value={item.status}
              disabled={saving}
              onClick={(event) => event.preventDefault()}
              onChange={handleStatusChange}
              className="max-w-[108px] rounded-md border border-border bg-background px-1.5 py-1 text-[10px] font-semibold text-text-secondary outline-none focus:border-brand disabled:opacity-50"
            >
              <option value="todo">To do</option>
              <option value="in_progress">In Progress</option>
              <option value="review">Review</option>
              <option value="done">Done</option>
            </select>
          )}
          <PointsBadge points={item.storyPoints} />
          <WorkItemAvatar member={assignee} />
        </div>
      </div>
    </Link>
  );
}

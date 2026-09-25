import Link from "next/link";
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
}

export function WorkItemCard({ item, assignee, className }: WorkItemCardProps) {
  return (
    <Link
      href={`/work-items/${item.id}`}
      className={cn(
        "flex flex-col gap-2 rounded-[10px] border border-border bg-surface p-3 shadow-[0_1px_2px_rgba(16,24,40,0.06)] transition-colors hover:border-border-strong",
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
          <PointsBadge points={item.storyPoints} />
          <WorkItemAvatar member={assignee} />
        </div>
      </div>
    </Link>
  );
}

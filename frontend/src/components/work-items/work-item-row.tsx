import Link from "next/link";
import type { Member, WorkItem } from "@/types/work-item";
import { TypeBadge } from "@/components/work-items/type-badge";
import { PriorityIcon } from "@/components/work-items/priority-icon";
import { PointsBadge } from "@/components/work-items/points-badge";
import { EpicPill } from "@/components/work-items/epic-pill";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { cn } from "@/lib/cn";

interface WorkItemRowProps {
  item: WorkItem;
  assignee?: Member;
  className?: string;
}

export function WorkItemRow({ item, assignee, className }: WorkItemRowProps) {
  return (
    <Link
      href={`/work-items/${item.id}`}
      className={cn(
        "flex items-center gap-3 border-b border-border px-3.5 py-[9px] last:border-b-0 hover:bg-background",
        className,
      )}
    >
      <TypeBadge type={item.type} size="md" />
      <span className="w-[76px] shrink-0 font-mono text-[11.5px] font-bold text-text-tertiary">
        {item.id}
      </span>
      <span className="flex-1 truncate text-[13px] font-medium text-text">
        {item.title}
      </span>
      <EpicPill epic={item.epic} />
      <PriorityIcon priority={item.priority} />
      <PointsBadge points={item.storyPoints} className="size-[22px]" />
      <WorkItemAvatar member={assignee} size="md" />
    </Link>
  );
}

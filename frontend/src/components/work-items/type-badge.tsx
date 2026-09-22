import type { WorkItemType } from "@/types/work-item";
import { BookmarkIcon, BugIcon, CheckSquareIcon } from "@/components/icons";
import { cn } from "@/lib/cn";

const TYPE_CONFIG: Record<
  WorkItemType,
  { label: string; icon: typeof BookmarkIcon; className: string }
> = {
  story: {
    label: "User Story",
    icon: BookmarkIcon,
    className: "bg-story-tint text-story",
  },
  task: {
    label: "Task",
    icon: CheckSquareIcon,
    className: "bg-task-tint text-task",
  },
  bug: {
    label: "Bug",
    icon: BugIcon,
    className: "bg-bug-tint text-bug",
  },
};

interface TypeBadgeProps {
  type: WorkItemType;
  size?: "sm" | "md";
  className?: string;
}

export function TypeBadge({ type, size = "sm", className }: TypeBadgeProps) {
  const { icon: TypeIcon, label } = TYPE_CONFIG[type];
  const dimension = size === "sm" ? "size-[18px]" : "size-5";
  const iconSize = size === "sm" ? "size-3.5" : "size-4";

  return (
    <div
      className={cn(
        "flex shrink-0 items-center justify-center rounded-[5px]",
        dimension,
        TYPE_CONFIG[type].className,
        className,
      )}
      title={label}
    >
      <TypeIcon className={iconSize} />
    </div>
  );
}

export function typeLabel(type: WorkItemType): string {
  return TYPE_CONFIG[type].label;
}

import type { Member } from "@/types/work-item";
import { cn } from "@/lib/cn";

const SIZE_CLASSES = {
  sm: "size-5 text-[9px]",
  md: "size-[22px] text-[9.5px]",
  lg: "size-7 text-[11px]",
} as const;

interface WorkItemAvatarProps {
  member?: Member;
  size?: keyof typeof SIZE_CLASSES;
  className?: string;
}

export function WorkItemAvatar({
  member,
  size = "md",
  className,
}: WorkItemAvatarProps) {
  if (!member) {
    return (
      <div
        className={cn(
          "flex shrink-0 items-center justify-center rounded-full border border-dashed border-border-strong bg-background text-text-tertiary",
          SIZE_CLASSES[size],
          className,
        )}
        title="Unassigned"
      >
        &ndash;
      </div>
    );
  }

  return (
    <div
      className={cn(
        "flex shrink-0 items-center justify-center rounded-full font-bold text-white",
        SIZE_CLASSES[size],
        className,
      )}
      style={{ backgroundColor: member.color }}
      title={member.name}
    >
      {member.initials}
    </div>
  );
}

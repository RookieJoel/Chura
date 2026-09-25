import { AlertCircleIcon } from "@/components/icons";
import { cn } from "@/lib/cn";

export function BlockedBadge({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex w-fit items-center gap-1 rounded-full bg-bug-tint px-2 py-0.5 text-[10.5px] font-bold text-bug",
        className,
      )}
    >
      <AlertCircleIcon className="size-[11px]" />
      Blocked
    </span>
  );
}

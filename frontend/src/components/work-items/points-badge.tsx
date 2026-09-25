import { cn } from "@/lib/cn";

interface PointsBadgeProps {
  points: number;
  className?: string;
}

export function PointsBadge({ points, className }: PointsBadgeProps) {
  return (
    <span
      className={cn(
        "flex size-5 shrink-0 items-center justify-center rounded-md border border-border-strong bg-surface text-[10px] font-bold text-text-secondary",
        className,
      )}
      title={`${points} story point${points === 1 ? "" : "s"}`}
    >
      {points}
    </span>
  );
}

import { cn } from "@/lib/cn";

const EPIC_STYLES: Record<string, string> = {
  Backlog: "bg-brand-tint text-brand-dark",
  "Sprint Board": "bg-task-tint text-task",
  Reporting: "bg-teal-tint text-teal",
};

const FALLBACK_STYLE = "bg-background text-text-secondary";

interface EpicPillProps {
  epic: string;
  className?: string;
}

export function EpicPill({ epic, className }: EpicPillProps) {
  return (
    <span
      className={cn(
        "w-fit shrink-0 rounded-full px-2 py-0.5 text-[10.5px] font-bold",
        EPIC_STYLES[epic] ?? FALLBACK_STYLE,
        className,
      )}
    >
      {epic}
    </span>
  );
}

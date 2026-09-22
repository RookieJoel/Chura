import type { Priority } from "@/types/work-item";
import { ArrowDownIcon, ArrowUpIcon, MinusIcon } from "@/components/icons";
import { cn } from "@/lib/cn";

const PRIORITY_CONFIG: Record<
  Priority,
  { label: string; icon: typeof ArrowUpIcon; className: string }
> = {
  highest: { label: "Highest", icon: ArrowUpIcon, className: "text-priority-high" },
  high: { label: "High", icon: ArrowUpIcon, className: "text-priority-high" },
  medium: { label: "Medium", icon: MinusIcon, className: "text-priority-medium" },
  low: { label: "Low", icon: ArrowDownIcon, className: "text-priority-low" },
};

interface PriorityIconProps {
  priority: Priority;
  className?: string;
}

export function PriorityIcon({ priority, className }: PriorityIconProps) {
  const { icon: Icon, label, className: colorClassName } =
    PRIORITY_CONFIG[priority];
  return (
    <Icon
      className={cn("size-3.5 shrink-0", colorClassName, className)}
      role="img"
      aria-label={`Priority: ${label}`}
    />
  );
}

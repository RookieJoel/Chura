"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { members } from "@/lib/mock-data";
import {
  BarChartIcon,
  CalendarIcon,
  ColumnsIcon,
  ListIcon,
  SettingsIcon,
} from "@/components/icons";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { cn } from "@/lib/cn";

const NAV_ITEMS = [
  // { href: "/board", label: "Board", icon: ColumnsIcon },
  { href: "/backlog", label: "Backlog", icon: ListIcon },
  { href: "/sprints", label: "Sprints", icon: CalendarIcon },
  // { href: "/reports", label: "Reports", icon: BarChartIcon },
  // { href: "/settings", label: "Settings", icon: SettingsIcon },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <nav className="flex w-[200px] shrink-0 flex-col gap-0.5 border-r border-border bg-surface p-3">
      <div className="px-2.5 py-1.5 text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
        Project
      </div>
      {NAV_ITEMS.map(({ href, label, icon: ItemIcon }) => {
        const active = pathname.startsWith(href);
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex items-center gap-2.5 rounded-[7px] px-2.5 py-2 text-[13px] font-semibold",
              active ? "bg-brand-tint text-brand-dark" : "text-text-secondary",
            )}
          >
            <ItemIcon className="size-3.5" />
            {label}
          </Link>
        );
      })}
      <div className="mx-1.5 my-3 h-px bg-border" />
      <div className="px-2.5 py-1.5 text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
        Team
      </div>
      {members.map((member) => (
        <div
          key={member.id}
          className="flex items-center gap-2 px-2.5 py-1.5 text-[12.5px] text-text-secondary"
        >
          <WorkItemAvatar member={member} size="sm" />
          {member.name.split(" ")[0]}
        </div>
      ))}
    </nav>
  );
}
